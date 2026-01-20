package rpcroute

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/eth/filters"
	"github.com/ava-labs/libevm/ethclient"
	"github.com/ava-labs/libevm/event"
	"github.com/ava-labs/libevm/libevm/ethapi"
	"github.com/ava-labs/libevm/log"
	"github.com/ava-labs/libevm/rpc"
)

// func TestMain(m *testing.M) {
// 	var opts []goleak.Option
// 	for _, ignore := range []string{
// 		// All leaked by upstream geth code
// 		"github.com/ava-labs/libevm/eth/filters.(*EventSystem).eventLoop",
// 		"github.com/ava-labs/libevm/rpc.(*Client).dispatch",
// 		"github.com/ava-labs/libevm/metrics.(*meterArbiter).tick",
// 		"github.com/ava-labs/libevm/core.(*txSenderCacher).cache",
// 	} {
// 		opts = append(opts, goleak.IgnoreTopFunction(ignore))
// 	}
// 	goleak.VerifyTestMain(m, opts...)
// }

var _ Backend = (*stubBackend)(nil)

type stubBackend struct {
	id      int
	height  uint64
	newHead event.FeedOf[core.ChainEvent]

	*bufconn.Listener
	rpc      *rpc.Server
	http, ws *httptest.Server
	httpURL  *url.URL

	// We only implement the RPC backend methods necessary to serve the
	// HTTP/WS connections, and embed the interface to satisfy the rest.
	ethapi.Backend
}

func (b *stubBackend) Addr() net.Addr {
	return b
}

func (*stubBackend) Network() string {
	return "stub"
}

func (b *stubBackend) String() string {
	return fmt.Sprintf("stub:%d", b.id)
}

func newHTTPServer(tb testing.TB, lis net.Listener, h http.Handler) *httptest.Server {
	tb.Helper()
	s := httptest.NewUnstartedServer(h)
	s.Listener = lis
	s.Start()
	return s
}

func newStubBackend(tb testing.TB, id int) *stubBackend {
	tb.Helper()

	lis := bufconn.Listen(1 << 10)
	tb.Cleanup(func() {
		lis.Close()
	})

	r := rpc.NewServer()
	tb.Cleanup(r.Stop)
	b := &stubBackend{
		id:       id,
		rpc:      r,
		Listener: lis,
	}
	b.http = newHTTPServer(tb, b, r)
	b.ws = newHTTPServer(tb, b, r.WebsocketHandler([]string{"*"}))
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
	return ethclient.NewClient(rpc.DialInProc(b.rpc)), nil
}

func (b *stubBackend) Redirect(u *url.URL) {
	*u = *b.httpURL
}

func (*stubBackend) Removed(error) {}

// sut is the system under test.
type sut struct {
	server   *Server
	proxy    *ethclient.Client
	backends []*stubBackend
	url      *url.URL
	cookies  http.CookieJar
}

func newSUT(tb testing.TB, numBackends int) sut {
	tb.Helper()
	ctx := tb.Context()

	var backends []*stubBackend
	for i := range numBackends {
		b := newStubBackend(tb, i)
		backends = append(backends, b)
	}

	tx := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				panic(err)
				return nil, err
			}
			if host != "stub" {
				panic(host)
				return nil, fmt.Errorf("unknown network host %q", host)
			}
			p, err := strconv.Atoi(port)
			if err != nil {
				panic(err)
				return nil, err
			}
			if p >= len(backends) {
				panic(fmt.Sprintf("%d >= %d", p, len(backends)))
				return nil, fmt.Errorf("port %d higher than greatest ID", p)
			}
			return backends[p].Dial()
		},
	}

	srv, err := NewServer(ctx, WithRoundTripper(tx))
	require.NoError(tb, err, "NewServer()")
	tb.Cleanup(func() {
		srv.Close()
		// Close the RPC connections so [goleak] doesn't complain.
		runtime.GC()
	})

	for _, b := range backends {
		require.NoError(tb, srv.AddBackend(ctx, b), "AddBackend()")
	}

	httpSrv := httptest.NewServer(srv)
	tb.Cleanup(httpSrv.Close)

	u, err := url.Parse(httpSrv.URL)
	require.NoError(tb, err)

	cl := httpSrv.Client()
	jar, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(tb, err)
	cl.Jar = jar

	proxy, err := rpc.DialOptions(ctx, httpSrv.URL, rpc.WithHTTPClient(cl))
	require.NoErrorf(tb, err, "ethclient.DialContext(ctx, %T{%T}.URL)", httpSrv, srv)
	tb.Cleanup(proxy.Close)

	return sut{
		server:   srv,
		proxy:    ethclient.NewClient(proxy),
		backends: backends,
		url:      u,
		cookies:  jar,
	}
}

func TestServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	{
		old := log.Root()
		defer func() {
			log.SetDefault(old)
		}()
		log.SetDefault(log.NewLogger(slog.NewTextHandler(os.Stderr, nil)))
	}

	sut := newSUT(t, 10)
	srv := sut.server

	assert.False(t, srv.Ready(), "Ready() before any blocks")

	requireFrontier := func(t *testing.T, wantHeight uint64, wantFrontier []int) {
		t.Helper()

		require.EventuallyWithT(t, func(c *assert.CollectT) {
			require.True(t, srv.Ready(), "Ready()")
			assert.Equal(c, wantHeight, srv.height.Load(), "height")
		}, 500*time.Millisecond, 20*time.Millisecond)

		t.Run("BlockNumber_from_reverse_proxy", func(t *testing.T) {
			// The backend that serves the request is chosen at random from the
			// frontier, so we can't test each one directly. However, sufficient
			// sampling without failure gives a very reliable p-score. The most
			// conservative estimate comes from assuming only a single incorrect
			// node: even with 10 in the set, the probability of not hitting the
			// bad node with 300 trials is vanishingly small: 0.9^300 = 1.9e-14.
			for range 300 {
				got, err := sut.proxy.BlockNumber(ctx)
				require.NoError(t, err, "%T.BlockNumber()", sut.proxy)
				require.Equal(t, wantHeight, got)
			}
		})

		return
		opts := cmp.Options{
			cmp.Transformer("backendIDs", func(bs []*backend) []int {
				var out []int
				for _, b := range bs {
					out = append(out, b.back.(*stubBackend).id)
				}
				return out
			}),
			cmpopts.SortSlices(func(a, b int) bool {
				return a < b
			}),
		}
		if diff := cmp.Diff(wantFrontier, *sut.server.frontier.Load(), opts); diff != "" {
			t.Errorf("backend indices in frontier set (-want +got):\n%s", diff)
		}
	}

	bs := sut.backends
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

func TestMonotonicBlockNumber(t *testing.T) {
	t.Skip("")
	ctx := t.Context()

	const numBackends = 10
	sut := newSUT(t, numBackends)

	sut.backends[0].increment()
	for !sut.server.Ready() {
		time.Sleep(10 * time.Millisecond)
	}

	var lastBlockNum uint64
	for range 1_000 {
		sut.backends[rand.IntN(numBackends)].increment()

		got, err := sut.proxy.BlockNumber(ctx)
		require.NoError(t, err)

		require.GreaterOrEqual(t, got, lastBlockNum)
		lastBlockNum = got
	}
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
