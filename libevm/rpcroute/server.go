package rpcroute

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/ava-labs/libevm/ethclient"
	"github.com/ava-labs/libevm/libevm/options"
)

// A Backend
type Backend interface {
	Label() string
	DialWS(context.Context) (*ethclient.Client, error)
	Redirect(*url.URL)
	Removed(error)
}

// A Server
type Server struct {
	height atomic.Uint64

	addBackend     chan *newBackend
	removeBackend  chan *removeBackend
	frontier       atomic.Pointer[frontier]
	updateFrontier chan struct{}

	spawned sync.WaitGroup
	quit    chan struct{}

	proxy        *httputil.ReverseProxy
	roundTripper http.RoundTripper
}

type frontier struct {
	height   uint64
	backends map[string]*backend
}

// NewServer
func NewServer(ctx context.Context, opts ...Option) (*Server, error) {
	c := options.ApplyTo(&config{
		roundTripper: http.DefaultTransport,
	}, opts...)

	s := &Server{
		addBackend:     make(chan *newBackend),
		removeBackend:  make(chan *removeBackend),
		updateFrontier: make(chan struct{}, 1),
		quit:           make(chan struct{}),
		roundTripper:   c.roundTripper,
	}
	s.frontier.Store(new(frontier))
	s.proxy = &httputil.ReverseProxy{
		Director:  s.reverseProxyDirector,
		Transport: s.reverseProxyTransport(),
	}

	s.spawn(s.manageBackends)
	return s, nil
}

type config struct {
	roundTripper http.RoundTripper
}

// An Option
type Option = options.Option[config]

// WithRoudnTripper
func WithRoundTripper(rt http.RoundTripper) Option {
	return options.Func[config](func(c *config) {
		c.roundTripper = rt
	})
}

// spawn is equivalent to [sync.WaitGroup.Go].
// TODO(arr4n) replace when using 1.25+.
func (s *Server) spawn(fn func()) {
	s.spawned.Add(1)
	go func() {
		fn()
		s.spawned.Done()
	}()
}

// Close
func (s *Server) Close() {
	close(s.quit)
	s.spawned.Wait()

	// Open connections are closed via GC cleanups. We know that all goroutines
	// have now returned so the only remaining references are in the frontier.
	s.frontier.Store(nil)
}

// reattachContext
func (s *Server) reattachContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancelCause(context.WithoutCancel(ctx))
	s.spawn(func() {
		<-s.quit
		cancel(ErrServerClosed)
	})
	return ctx
}

func (s *Server) Ready() bool {
	return len(s.frontier.Load().backends) > 0
}
