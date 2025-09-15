// Copyright 2025 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

package parallel

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/consensus"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/hookstest"
	"github.com/ava-labs/libevm/params"
	"github.com/ava-labs/libevm/trie"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreCurrent())
}

type shaHandler struct {
	addr common.Address
	gas  uint64
}

func (h *shaHandler) Gas(tx *types.Transaction) (uint64, bool) {
	if to := tx.To(); to == nil || *to != h.addr {
		return 0, false
	}
	return h.gas, true
}

func (*shaHandler) Process(i int, tx *types.Transaction) [sha256.Size]byte {
	return sha256.Sum256(tx.Data())
}

func TestProcessor(t *testing.T) {
	handler := &shaHandler{
		addr: common.Address{'s', 'h', 'a', 2, 5, 6},
		gas:  1e6,
	}
	p := New(handler, 8)
	t.Cleanup(p.Close)

	type blockParams struct {
		numTxs                              int
		sendToAddrEvery, sufficientGasEvery int
	}

	// Each set of params is effectively a test case, but they are all run on
	// the same [Processor].
	params := []blockParams{
		{
			numTxs: 0,
		},
		{
			numTxs:             500,
			sendToAddrEvery:    7,
			sufficientGasEvery: 5,
		},
		{
			numTxs:             1_000,
			sendToAddrEvery:    7,
			sufficientGasEvery: 5,
		},
		{
			numTxs:             1_000,
			sendToAddrEvery:    11,
			sufficientGasEvery: 3,
		},
		{
			numTxs:             100,
			sendToAddrEvery:    1,
			sufficientGasEvery: 1,
		},
		{
			numTxs: 0,
		},
	}

	rng := rand.New(rand.NewPCG(0, 0)) //nolint:gosec // Reproducibility is useful for testing
	for range 100 {
		params = append(params, blockParams{
			numTxs:             rng.IntN(1000),
			sendToAddrEvery:    1 + rng.IntN(30),
			sufficientGasEvery: 1 + rng.IntN(30),
		})
	}

	for _, tc := range params {
		t.Run("", func(t *testing.T) {
			t.Logf("%+v", tc)

			txs := make(types.Transactions, tc.numTxs)
			wantProcessed := make([]bool, tc.numTxs)
			for i := range len(txs) {
				var (
					to       common.Address
					extraGas uint64
				)

				wantProcessed[i] = true
				if i%tc.sendToAddrEvery == 0 {
					to = handler.addr
				} else {
					wantProcessed[i] = false
				}
				if i%tc.sufficientGasEvery == 0 {
					extraGas = handler.gas
				} else {
					wantProcessed[i] = false
				}

				data := binary.BigEndian.AppendUint64(nil, uint64(i))
				gas, err := core.IntrinsicGas(data, nil, false, true, true, true)
				require.NoError(t, err, "core.IntrinsicGas(%#x, nil, false, true, true, true)", data)

				txs[i] = types.NewTx(&types.LegacyTx{
					To:   &to,
					Data: data,
					Gas:  gas + extraGas,
				})
			}

			block := types.NewBlock(&types.Header{}, txs, nil, nil, trie.NewStackTrie(nil))
			require.NoError(t, p.StartBlock(block), "BeforeBlock()")
			defer p.FinishBlock(block)

			for i, tx := range txs {
				wantOK := wantProcessed[i]

				var want [sha256.Size]byte
				if wantOK {
					want = handler.Process(i, tx)
				}

				got, gotOK := p.Result(i)
				if got != want || gotOK != wantOK {
					t.Errorf("Result(%d) got (%#x, %t); want (%#x, %t)", i, got, gotOK, want, wantOK)
				}
			}
		})

		if t.Failed() {
			break
		}
	}
}

type noopHooks struct{}

func (noopHooks) OverrideNewEVMArgs(a *vm.NewEVMArgs) *vm.NewEVMArgs {
	return a
}

func (noopHooks) OverrideEVMResetArgs(_ params.Rules, a *vm.EVMResetArgs) *vm.EVMResetArgs {
	return a
}

type vmHooks struct {
	vm.Preprocessor // the [Processor]
	noopHooks
}

func TestIntegration(t *testing.T) {
	const handlerGas = 500
	handler := &shaHandler{
		addr: common.Address{'s', 'h', 'a', 2, 5, 6},
		gas:  handlerGas,
	}
	sut := New(handler, 8)
	t.Cleanup(sut.Close)

	vm.RegisterHooks(vmHooks{Preprocessor: sut})
	t.Cleanup(vm.TestOnlyClearRegisteredHooks)

	stub := &hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			handler.addr: vm.NewStatefulPrecompile(func(env vm.PrecompileEnvironment, input []byte) (ret []byte, err error) {
				sdb := env.StateDB()
				txi, txh := sdb.TxIndex(), sdb.TxHash()

				// Precompiles MUST NOT charge gas for the preprocessing as it
				// would then be double-counted.
				got, ok := sut.Result(txi)
				if !ok {
					t.Errorf("no result for tx[%d] %v", txi, txh)
				}
				env.StateDB().AddLog(&types.Log{
					Data: got[:],
				})
				return nil, nil
			}),
		},
	}
	stub.Register(t)

	state, evm := ethtest.NewZeroEVM(t)

	key, err := crypto.GenerateKey()
	require.NoErrorf(t, err, "crypto.GenerateKey()")
	eoa := crypto.PubkeyToAddress(key.PublicKey)
	state.CreateAccount(eoa)
	state.AddBalance(eoa, uint256.NewInt(10*params.Ether))

	var (
		txs  types.Transactions
		want []*types.Receipt
	)
	ignore := cmp.Options{
		cmpopts.IgnoreFields(
			types.Receipt{},
			"PostState", "CumulativeGasUsed", "BlockNumber", "BlockHash", "Bloom",
		),
		cmpopts.IgnoreFields(types.Log{}, "BlockHash"),
	}

	signer := types.LatestSigner(evm.ChainConfig())
	for i, addr := range []common.Address{
		{'o', 't', 'h', 'e', 'r'},
		handler.addr,
	} {
		ui := uint(i) //nolint:gosec // Known value that won't overflow
		data := []byte("hello, world")

		// Having all arguments `false` is equivalent to what
		// [core.ApplyTransaction] will do.
		gas, err := core.IntrinsicGas(data, types.AccessList{}, false, false, false, false)
		require.NoError(t, err, "core.IntrinsicGas(%#x, nil, false, false, false, false)", data)
		if addr == handler.addr {
			gas += handlerGas
		}

		tx := types.MustSignNewTx(key, signer, &types.LegacyTx{
			Nonce: uint64(ui),
			To:    &addr,
			Data:  data,
			Gas:   gas,
		})
		txs = append(txs, tx)

		wantR := &types.Receipt{
			Status:           types.ReceiptStatusSuccessful,
			TxHash:           tx.Hash(),
			GasUsed:          gas,
			TransactionIndex: ui,
		}
		if addr == handler.addr {
			res := handler.Process(i, tx)
			wantR.Logs = []*types.Log{{
				TxHash:  tx.Hash(),
				TxIndex: ui,
				Data:    res[:],
			}}
		}
		want = append(want, wantR)
	}

	block := types.NewBlock(&types.Header{}, txs, nil, nil, trie.NewStackTrie(nil))
	require.NoError(t, sut.StartBlock(block), "StartBlock()")
	defer sut.FinishBlock(block)

	pool := core.GasPool(math.MaxUint64)
	var got []*types.Receipt
	for i, tx := range txs {
		state.SetTxContext(tx.Hash(), i)

		var usedGas uint64
		receipt, err := core.ApplyTransaction(
			evm.ChainConfig(),
			chainContext{},
			&block.Header().Coinbase,
			&pool,
			state,
			block.Header(),
			tx,
			&usedGas,
			vm.Config{},
		)
		require.NoError(t, err, "ApplyTransaction([%d])", i)
		got = append(got, receipt)
	}

	if diff := cmp.Diff(want, got, ignore); diff != "" {
		t.Errorf("%T diff (-want +got):\n%s", got, diff)
	}
}

// Dummy implementations of interfaces required by [core.ApplyTransaction].
type (
	chainContext struct{}
	engine       struct{ consensus.Engine }
)

func (chainContext) Engine() consensus.Engine                    { return engine{} }
func (chainContext) GetHeader(common.Hash, uint64) *types.Header { panic("unimplemented") }
func (engine) Author(h *types.Header) (common.Address, error)    { return common.Address{}, nil }
