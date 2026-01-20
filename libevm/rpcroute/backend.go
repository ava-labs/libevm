package rpcroute

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cloudflare/backoff"
	"github.com/gorilla/websocket"

	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethclient"
	"github.com/ava-labs/libevm/event"
)

type backend struct {
	server *Server

	height atomic.Uint64
	back   Backend
	quit   chan chan struct{}
}

func withBackoff[T any](bo *backoff.Backoff, tries int, fn func() (T, error)) (T, error) {
	var err error
	for range tries {
		var res T
		res, err = fn()
		if err == nil { // NOTE == not !=
			return res, nil
		}
		d := bo.Duration()
		fmt.Println(d, err)
		time.Sleep(d)
	}
	var zero T
	return zero, err
}

func withDefaultBackoff[T any](fn func() (T, error)) (T, error) {
	// 2^16ms ~ 64s
	bo := backoff.New(0, time.Millisecond)
	return withBackoff(bo, 16, fn)
}

func (b *backend) dialWS(ctx context.Context) (*ethclient.Client, error) {
	ws, err := withDefaultBackoff(func() (*ethclient.Client, error) {
		return b.back.DialWS(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("%T{%q}.DialWS(): %v", b.back, b.back.Label(), err)
	}
	return ws, nil
}

func (b *backend) subscribe(ctx context.Context, ws *ethclient.Client, ch chan<- *types.Header) (event.Subscription, error) {
	return withDefaultBackoff(func() (event.Subscription, error) {
		return ws.SubscribeNewHead(ctx, ch)
	})
}

var ErrServerClosed = errors.New("server closed")

// `ready` MUST be buffered such that sending on it is non-blocking even when
// there is no longer a receiver.
func (b *backend) trackHeight(ctx context.Context) error {
	ready := make(chan error, 1)
	b.server.spawn(func() {
		b.heightLoop(ctx, ready)
	})
	return <-ready
}

func (b *backend) heightLoop(ctx context.Context, ready chan<- error) (retErr error) {
	defer func() {
		ready <- retErr // buffered
		b.server.removeBackend <- &removeBackend{
			label:  b.back.Label(),
			reason: retErr,
			err:    make(chan error, 1), // buffer -> black hole
		}
	}()

	headers := make(chan *types.Header, 16)
	subscribe := func() (event.Subscription, error) {
		ws, err := b.dialWS(ctx)
		if err != nil {
			return nil, err
		}
		sub, err := b.subscribe(ctx, ws, headers)
		if err != nil {
			ws.Close()
			return nil, err
		}
		runtime.AddCleanup(&sub, (*ethclient.Client).Close, ws)
		return sub, nil
	}

	sub, err := subscribe()
	if err != nil {
		return err
	}
	ready <- nil
	// defer sub.Unsubscribe()

	ctx = context.WithoutCancel(ctx) //b.server.reattachContext(ctx)

	for {
		select {
		case <-b.server.quit:
			return ErrServerClosed

		case ch := <-b.quit:
			defer close(ch)
			return nil

		case err := <-sub.Err():
			if websocket.IsUnexpectedCloseError(err) {
				fmt.Println("UnexpectedClose")
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				fmt.Println("EOF")
			}
			if err != nil {
				fmt.Println(err)
			}
			// No need to wait for the defer to trigger. The
			// [event.Subscription] interface documents this method as being
			// safe to call multiple times.
			// sub.Unsubscribe()

			// This MUST NOT be := as we are reestablishing the connection of
			// the outer variable.
			sub, err = subscribe()
			if err != nil {
				return err
			}
			// defer sub.Unsubscribe()

		case hdr := <-headers:
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
