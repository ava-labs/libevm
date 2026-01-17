package rpcroute

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethclient"
	"golang.org/x/sync/errgroup"
)

type Backend interface {
	Label() string
	DialWS(context.Context) (*ethclient.Client, error)
	Redirect(*url.URL)
}

const RoutedToHeader = "X-RPC-Routed-To"

func NewServer[B Backend](ctx context.Context, backends ...B) (*Server, error) {
	s := &Server{
		backends:       make([]*backend, len(backends)),
		updateFrontier: make(chan struct{}, 1),
		quit:           make(chan struct{}),
	}

	s.proxy = &httputil.ReverseProxy{
		Director: s.reverseProxyDirector,
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			res, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				return nil, err
			}
			for _, h := range req.Header.Values(RoutedToHeader) {
				res.Header.Add(RoutedToHeader, h)
			}
			return res, nil
		}),
	}

	{
		g, ctx := errgroup.WithContext(ctx)
		for i, b := range backends {
			g.Go(func() error {
				ws, err := b.DialWS(ctx)
				if err != nil {
					return fmt.Errorf("%T{%q}.DialWS(): %v", b, b.Label(), err)
				}
				s.backends[i] = &backend{
					server: s,
					back:   b,
					ws:     ws,
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			for _, b := range s.backends {
				if b != nil {
					b.ws.Close()
				}
			}
			return nil, err
		}
	}

	{
		// The Context is only relevant to the scope of creating the server, but
		// not for its entire lifetime. We must therefore replace the
		// cancellation mechanism with an appropriate one.
		ctx := context.WithoutCancel(ctx)
		ctx, cancel := context.WithCancel(ctx)
		s.done.Add(1)
		go func() {
			<-s.quit
			cancel()
			s.done.Done()
		}()

		for _, b := range s.backends {
			s.ready.Add(1)
			s.done.Add(1)
			go b.trackHeight(ctx)
		}
	}

	s.done.Add(1)
	go s.manageFrontierSet()

	s.ready.Wait()
	return s, nil
}

type Server struct {
	height atomic.Uint64

	backends []*backend
	ready    sync.WaitGroup
	quit     chan struct{}
	done     sync.WaitGroup

	frontier       atomic.Pointer[[]*backend]
	updateFrontier chan struct{}

	proxy *httputil.ReverseProxy
}

func (s *Server) Close() {
	close(s.quit)
	s.done.Wait()
}

func (s *Server) manageFrontierSet() {
	defer s.done.Done()

	for {
		select {
		case <-s.updateFrontier:
			h := s.height.Load()
			frontier := make([]*backend, 0, len(s.backends))
			for _, b := range s.backends {
				if b.height.Load() >= h { // MUST be >= and not just == as backend is incremented first
					frontier = append(frontier, b)
				}
			}
			s.frontier.Store(&frontier)

		case <-s.quit:
			return
		}
	}
}

func (s *Server) triggerFrontierUpdate() {
	select {
	case s.updateFrontier <- struct{}{}:
	default:
	}
}

func (s *Server) Ready() bool {
	return s.frontier.Load() != nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (s *Server) reverseProxyDirector(r *http.Request) {
	f := *s.frontier.Load()
	b := f[rand.IntN(len(f))].back

	r.Host = ""
	r.Header.Add(RoutedToHeader, b.Label())
	b.Redirect(r.URL)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type backend struct {
	server *Server

	height atomic.Uint64
	back   Backend
	ws     *ethclient.Client
}

func (b *backend) trackHeight(ctx context.Context) {
	defer func() {
		b.ws.Close()
		b.server.done.Done()
	}()

	head := make(chan *types.Header, 16)
	sub, err := b.ws.SubscribeNewHead(ctx, head)
	if err != nil {
		return
	}
	defer sub.Unsubscribe()
	b.server.ready.Done()

	for {
		select {
		case <-b.server.quit:
			return
		case <-sub.Err():
			return

		case hdr := <-head:
			num := hdr.Number.Uint64()
			b.height.Store(num)

			old := b.server.height.Load()
			if num == old {
				// Add this node to the frontier.
				b.server.triggerFrontierUpdate()
			}
			if num <= old || !b.server.height.CompareAndSwap(old, num) {
				continue
			}
			// Increase the height of the frontier.
			b.server.triggerFrontierUpdate()

			// TODO(arr4n) explore pre-warming of receipt caches, resulting in a
			// single RPC node being responsible for serving of the latest
			// block.
		}
	}
}
