package rpcroute

import (
	"errors"
	"net/http"
	"time"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

// RoutedToHeader
const RoutedToHeader = "X-RPC-Routed-To"

func (s *Server) reverseProxyDirector(r *http.Request) {
	var sticky string
	if c, err := r.Cookie(RoutedToHeader); err == nil { // NOTE == not !=
		sticky = c.Value
	}

	b, err := func() (*backend, error) {
		for range 10 {
			all := s.frontier.Load().backends
			if b, ok := all[sticky]; ok {
				return b, nil
			}
			for _, b := range s.frontier.Load().backends { // random order due to being a map
				return b, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
		return nil, errors.New("x")
	}()
	if err != nil {
		panic(err)
	}

	r.Host = ""
	bb := b.back
	r.Header.Add(RoutedToHeader, bb.Label())
	bb.Redirect(r.URL)
}

func (s *Server) reverseProxyTransport() http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		res, err := s.roundTripper.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		for _, h := range req.Header.Values(RoutedToHeader) {
			res.Header.Add(RoutedToHeader, h)
			c := &http.Cookie{
				Name:  RoutedToHeader,
				Value: h,
			}
			res.Header.Add("Set-Cookie", c.String())
			break
		}
		return res, nil
	})
}

var _ http.RoundTripper = roundTripperFunc(nil)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
