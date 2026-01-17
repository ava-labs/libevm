package rpcroute

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/exp/slog"

	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/eth/filters"
	"github.com/ava-labs/libevm/ethclient"
	"github.com/ava-labs/libevm/event"
	"github.com/ava-labs/libevm/libevm/ethapi"
	"github.com/ava-labs/libevm/log"
	"github.com/ava-labs/libevm/rpc"
)

func TestMain(m *testing.M) {
	var opts []goleak.Option
	for _, ignore := range []string{
		// All leaked by upstream geth code
		"github.com/ava-labs/libevm/eth/filters.(*EventSystem).eventLoop",
		"github.com/ava-labs/libevm/rpc.(*Client).dispatch",
		"github.com/ava-labs/libevm/metrics.(*meterArbiter).tick",
		"github.com/ava-labs/libevm/core.(*txSenderCacher).cache",
	} {
		opts = append(opts, goleak.IgnoreTopFunction(ignore))
	}
	goleak.VerifyTestMain(m, opts...)
}

var _ Backend = (*stubBackend)(nil)

type stubBackend struct {
	id      int
	height  uint64
	newHead event.FeedOf[core.ChainEvent]

	http, ws *httptest.Server
	httpURL  *url.URL

	// We only implement the RPC backend methods necessary to serve the
	// HTTP/WS connections, and embed the interface to satisfy the rest.
	ethapi.Backend
}

func newBackend(tb testing.TB, id int) *stubBackend {
	tb.Helper()

	r := rpc.NewServer()
	tb.Cleanup(r.Stop)
	b := &stubBackend{
		id:   id,
		http: httptest.NewServer(r),
		ws:   httptest.NewServer(r.WebsocketHandler([]string{"*"})),
	}
	tb.Cleanup(b.http.Close)
	tb.Cleanup(b.ws.Close)

	u, err := url.Parse(b.http.URL)
	require.NoErrorf(tb, err, "url.Parse(%T.URL)", b.http)
	b.httpURL = u

	events := filters.NewFilterAPI(
		filters.NewFilterSystem(b, filters.Config{}),
		false, /*lightMode*/
	)
	tb.Cleanup(func() { filters.CloseAPI(events) })

	for _, api := range []any{
		ethapi.NewBlockChainAPI(b), // BlockNumber
		events,                     // SubscribeNewHead()
	} {
		require.NoErrorf(tb, r.RegisterName("eth", api), "%T.RegisterName(%q, %T)", r, "eth", api)
	}
	return b
}

// SubscribeChainEvent implements the respective method of [ethapi.Backend],
// used to serve [ethclient.Client.SubscribeNewHead].
func (b *stubBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) sub {
	return b.newHead.Subscribe(ch)
}

func (b *stubBackend) increment() {
	b.height++
	bl := types.NewBlockWithHeader(b.header())
	b.newHead.Send(core.ChainEvent{
		Block: bl,
		Hash:  bl.Hash(),
	})
}

func (b *stubBackend) header() *types.Header {
	return &types.Header{
		Number: uint256.NewInt(b.height).ToBig(),
	}
}

// HeaderByNumber implements the respective method of [ethapi.Backend], used to
// serve [ethclient.Client.BlockNumber].
func (b *stubBackend) HeaderByNumber(ctx context.Context, n rpc.BlockNumber) (*types.Header, error) {
	if n != rpc.LatestBlockNumber {
		return nil, fmt.Errorf("unsupported block number %v", n)
	}
	return b.header(), nil
}

func (b *stubBackend) Label() string {
	return fmt.Sprintf("%T[%d]", b, b.id)
}

func (b *stubBackend) DialWS(ctx context.Context) (*ethclient.Client, error) {
	url := fmt.Sprintf("ws://%s", b.ws.Listener.Addr().String())
	c, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("rpc.Dial(%T.srv.Listener.Addr() = %q): %v", b, url, err)
	}
	return ethclient.NewClient(c), nil
}

func (b *stubBackend) Redirect(u *url.URL) {
	*u = *b.httpURL
}

func TestServer(t *testing.T) {
	ctx := t.Context()

	{
		old := log.Root()
		defer func() {
			log.SetDefault(old)
		}()
		log.SetDefault(log.NewLogger(slog.NewTextHandler(os.Stderr, nil)))
	}

	var bs []*stubBackend
	for i := range 10 {
		b := newBackend(t, i)
		bs = append(bs, b)
	}
	sut, err := NewServer(ctx, bs...)
	require.NoError(t, err, "NewServer()")
	t.Cleanup(sut.Close)

	srv := httptest.NewServer(sut)
	t.Cleanup(srv.Close)
	proxy, err := ethclient.DialContext(ctx, srv.URL)
	require.NoErrorf(t, err, "ethclient.DialContext(ctx, %T{%T}.URL)", srv, sut)
	t.Cleanup(proxy.Close)

	assert.False(t, sut.Ready(), "Ready() before any blocks")

	requireFrontier := func(t *testing.T, wantHeight uint64, wantFrontier []int) {
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			require.True(t, sut.Ready(), "Ready()")
			assert.Equal(c, wantHeight, sut.height.Load(), "height")
		}, 500*time.Millisecond, time.Millisecond)

		var gotFrontier []int
		for _, b := range *sut.frontier.Load() {
			gotFrontier = append(gotFrontier, b.back.(*stubBackend).id)
		}
		require.Equal(t, wantFrontier, gotFrontier, "backend indices in frontier set")

		t.Run("BlockNumber_from_reverse_proxy", func(t *testing.T) {
			// The backend that serves the request is chosen at random from the
			// frontier, so we can't test each one directly. However, sufficient
			// sampling without failure gives a very reliable p-score. The most
			// conservative estimate comes from assuming only a single incorrect
			// node: even with 10 in the set, the probability of not hitting the
			// bad node with 300 trials is vanishingly small: 0.9^300 = 1.9e-14.
			for range 300 {
				got, err := proxy.BlockNumber(ctx)
				require.NoError(t, err, "%T.BlockNumber()", proxy)
				require.Equal(t, wantHeight, got)
			}
		})
	}

	bs[0].increment()
	requireFrontier(t, 1, []int{0})

	bs[3].increment()
	requireFrontier(t, 1, []int{0, 3})
	bs[2].increment()
	requireFrontier(t, 1, []int{0, 2, 3})

	for range 2 {
		bs[1].increment()
	}
	requireFrontier(t, 2, []int{1})
	bs[4].increment()
	requireFrontier(t, 2, []int{1})
	bs[4].increment()
	requireFrontier(t, 2, []int{1, 4})

	bs[0].increment()
	bs[2].increment()
	bs[3].increment()
	for _, b := range bs[5:] {
		for range 2 {
			b.increment()
		}
	}
	requireFrontier(t, 2, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
}

// Everything below here is necessary to satisfy [ethapi.Backend] methods that
// are called by [filters.FilterAPI] but not actually relevant to the tests.

type (
	sub     = event.Subscription // for brevity
	noopSub struct {
		errs chan error
		once sync.Once
	}
)

func newNoopSub() sub {
	return &noopSub{errs: make(chan error)}
}

func (s *noopSub) Unsubscribe() {
	s.once.Do(func() { close(s.errs) })
}

func (s *noopSub) Err() <-chan error {
	return s.errs
}

func (*stubBackend) SubscribeNewTxsEvent(chan<- core.NewTxsEvent) sub           { return newNoopSub() }
func (*stubBackend) SubscribeLogsEvent(chan<- []*types.Log) sub                 { return newNoopSub() }
func (*stubBackend) SubscribeRemovedLogsEvent(chan<- core.RemovedLogsEvent) sub { return newNoopSub() }
func (*stubBackend) SubscribePendingLogsEvent(chan<- []*types.Log) sub          { return newNoopSub() }
