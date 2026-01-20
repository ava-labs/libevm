package rpcroute

import (
	"context"
	"errors"
	"fmt"
)

type (
	newBackend struct {
		*backend
		err chan error
	}
	removeBackend struct {
		label  string
		reason error
		err    chan error
	}
)

func (s *Server) manageBackends() {
	all := make(map[string]*backend)

	add := func(b *backend) error {
		lbl := b.back.Label()
		if _, ok := all[lbl]; ok {
			return fmt.Errorf("duplicate Backend %q", lbl)
		}
		all[lbl] = b
		return nil
	}

	remove := func(lbl string, reason error) error {
		b, ok := all[lbl]
		if !ok {
			return fmt.Errorf("unknown Backend %q", lbl)
		}
		delete(all, lbl)
		go b.back.Removed(reason)
		return nil
	}

	updateFrontier := func() {
		f := &frontier{
			height:   s.height.Load(),
			backends: make(map[string]*backend),
		}
		for _, b := range all {
			if b.height.Load() >= f.height { // MUST be >= and not just == as backend is incremented first
				f.backends[b.back.Label()] = b
			}
		}
		s.frontier.Store(f)
	}

	for {
		select {
		case b := <-s.addBackend:
			// The backend will itself trigger a frontier update so there's no
			// need for an explicit one as done on removal.
			b.err <- add(b.backend)

		case b := <-s.removeBackend:
			err := remove(b.label, b.reason)
			updateFrontier()
			b.err <- err

		case <-s.updateFrontier:
			updateFrontier()

		case <-s.quit:
			// backends remove themselves when receiving the `quit` signal.
			for len(all) > 0 {
				b := <-s.removeBackend
				remove(b.label, b.reason)
			}
			return
		}
	}
}

func (s *Server) triggerFrontierUpdate() {
	select {
	case s.updateFrontier <- struct{}{}:
		_ = 0
	default:
		_ = 0
	}
}

// AddBackend
func (s *Server) AddBackend(ctx context.Context, be Backend) error {
	b := &backend{
		server: s,
		back:   be,
		quit:   make(chan chan struct{}),
	}
	if err := b.trackHeight(ctx); err != nil {
		return err
	}

	errCh := make(chan error)
	s.addBackend <- &newBackend{
		backend: b,
		err:     errCh,
	}
	return <-errCh
}

// RemoveBackend
func (s *Server) RemoveBackend(label string) error {
	return errors.New("unimplemented")
}
