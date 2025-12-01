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
	"encoding/binary"
	"math"
	"math/big"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ava-labs/libevm/common"
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

type recorder struct {
	tb testing.TB

	gas                               uint64
	addr                              common.Address
	blockKey, prefetchKey, processKey common.Hash

	gotReceipts   types.Receipts
	gotAggregated aggregated
}

type aggregated struct {
	txOrder, processOrder []TxResult[recorded]
}

type recorded struct {
	HeaderExtra, TxData      []byte
	Block, Prefetch, Process common.Hash
}

type commonData struct {
	HeaderExtra         []byte
	BeforeBlockStateVal common.Hash
}

func (r *recorder) BeforeBlock(sdb libevm.StateReader, b *types.Block) commonData {
	return commonData{
		HeaderExtra:         slices.Clone(b.Header().Extra),
		BeforeBlockStateVal: sdb.GetState(r.addr, r.blockKey),
	}
}

func (r *recorder) Gas(tx *types.Transaction) (uint64, bool) {
	if to := tx.To(); to != nil && *to == r.addr {
		return r.gas, true
	}
	return 0, false
}

type prefetched struct {
	prefetchStateVal common.Hash
	common           commonData
}

func (r *recorder) Prefetch(sdb libevm.StateReader, tx IndexedTx, cd commonData) prefetched {
	return prefetched{
		common:           cd,
		prefetchStateVal: sdb.GetState(r.addr, r.prefetchKey),
	}
}

func (r *recorder) Process(sdb libevm.StateReader, tx IndexedTx, cd commonData, data prefetched) recorded {
	if diff := cmp.Diff(cd, data.common); diff != "" {
		r.tb.Errorf("Mismatched CommonData propagation to Handler methods; diff (-Process, +Prefetch):\n%s", diff)
	}

	return recorded{
		HeaderExtra: slices.Clone(cd.HeaderExtra),
		TxData:      slices.Clone(tx.Data()),
		Block:       cd.BeforeBlockStateVal,
		Prefetch:    data.prefetchStateVal,
		Process:     sdb.GetState(r.addr, r.processKey),
	}
}

func (r *recorded) asLog() *types.Log {
	return &types.Log{
		Topics: []common.Hash{
			r.Block, r.Prefetch, r.Process,
		},
		Data: slices.Concat(r.HeaderExtra, []byte("|"), r.TxData),
	}
}

func (r *recorder) PostProcess(res Results[recorded]) aggregated {
	defer res.WaitForAll()

	var out aggregated
	for res := range res.TxOrder {
		out.txOrder = append(out.txOrder, res)
	}
	for res := range res.ProcessOrder {
		out.processOrder = append(out.processOrder, res)
	}

	return out
}

func (r *recorder) AfterBlock(_ StateDB, agg aggregated, _ *types.Block, rs types.Receipts) {
	r.gotReceipts = slices.Clone(rs)
	r.gotAggregated = agg
}

func asHash(s string) (h common.Hash) {
	copy(h[:], []byte(s))
	return
}

func TestProcessor(t *testing.T) {
	handler := &recorder{
		tb:          t,
		addr:        common.Address{'c', 'o', 'n', 'c', 'a', 't'},
		gas:         1e6,
		blockKey:    asHash("block"),
		prefetchKey: asHash("prefetch"),
		processKey:  asHash("process"),
	}
	p := New(8, 8)
	getResult := AddHandler(p, handler)
	t.Cleanup(p.Close)

	type blockParams struct {
		numTxs                              int
		sendToAddrEvery, sufficientGasEvery int
	}

	// Each set of params is effectively a test case, but they are all run on
	// the same [Processor].
	tests := []blockParams{
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
		tests = append(tests, blockParams{
			numTxs:             rng.IntN(1000),
			sendToAddrEvery:    1 + rng.IntN(30),
			sufficientGasEvery: 1 + rng.IntN(30),
		})
	}

	_, _, sdb := ethtest.NewEmptyStateDB(t)
	h := handler
	blockVal := asHash("block_val")
	sdb.SetState(h.addr, h.blockKey, blockVal)
	prefetchVal := asHash("prefetch_val")
	sdb.SetState(h.addr, h.prefetchKey, prefetchVal)
	processVal := asHash("process_val")
	sdb.SetState(h.addr, h.processKey, processVal)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Logf("%+v", tt)

			var rules params.Rules
			txs := make(types.Transactions, tt.numTxs)
			wantProcessed := make([]bool, tt.numTxs)
			for i := range len(txs) {
				var (
					to       common.Address
					extraGas uint64
				)

				wantProcessed[i] = true
				if i%tt.sendToAddrEvery == 0 {
					to = handler.addr
				} else {
					wantProcessed[i] = false
				}
				if i%tt.sufficientGasEvery == 0 {
					extraGas = handler.gas
				} else {
					wantProcessed[i] = false
				}

				data := binary.BigEndian.AppendUint64(nil, uint64(i))
				gas, err := intrinsicGas(data, types.AccessList{}, &handler.addr, &rules)
				require.NoError(t, err, "core.IntrinsicGas(%#x, nil, false, ...)", data)

				txs[i] = types.NewTx(&types.LegacyTx{
					To:   &to,
					Data: data,
					Gas:  gas + extraGas,
				})
			}

			extra := []byte("extra")
			block := types.NewBlock(&types.Header{Extra: extra}, txs, nil, nil, trie.NewStackTrie(nil))
			require.NoError(t, p.StartBlock(sdb, rules, block), "StartBlock()")

			var wantPerTx []TxResult[recorded]
			for i, tx := range txs {
				wantOK := wantProcessed[i]

				var want recorded
				if wantOK {
					want = recorded{
						HeaderExtra: extra,
						Block:       blockVal,
						Prefetch:    prefetchVal,
						Process:     processVal,
						TxData:      tx.Data(),
					}
					wantPerTx = append(wantPerTx, TxResult[recorded]{
						Tx: IndexedTx{
							Index:       i,
							Transaction: tx,
						},
						Result: want,
					})
				}

				got, gotOK := getResult(i)
				if gotOK != wantOK {
					t.Errorf("Result(%d) got ok %t; want %t", i, gotOK, wantOK)
					continue
				}
				if diff := cmp.Diff(want, got.Result); diff != "" {
					t.Errorf("Result(%d) diff (-want +got):\n%s", i, diff)
				}
			}

			p.FinishBlock(sdb, block, nil)
			tests := []struct {
				name string
				got  []TxResult[recorded]
				opt  cmp.Option
			}{
				{
					name: "in_transaction_order",
					got:  h.gotAggregated.txOrder,
				},
				{
					name: "in_process_order",
					got:  h.gotAggregated.processOrder,
					opt: cmpopts.SortSlices(func(a, b TxResult[recorded]) bool {
						return a.Tx.Index < b.Tx.Index
					}),
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					opts := cmp.Options{
						tt.opt,
						cmp.Comparer(func(a, b *types.Transaction) bool {
							return a.Hash() == b.Hash()
						}),
					}
					if diff := cmp.Diff(wantPerTx, tt.got, opts); diff != "" {
						t.Errorf("handler.PostProcess() argument diff (-want +got):\n%s", diff)
					}
				})
			}
		})

		if t.Failed() {
			break
		}
	}
}

type vmHooks struct {
	vm.Preprocessor // the [Processor]
	vm.NOOPHooks
}

func (h *vmHooks) PreprocessingGasCharge(tx common.Hash) (uint64, error) {
	return h.Preprocessor.PreprocessingGasCharge(tx)
}

func TestIntegration(t *testing.T) {
	const handlerGas = 500
	handler := &recorder{
		tb:   t,
		addr: common.Address{'c', 'o', 'n', 'c', 'a', 't'},
		gas:  handlerGas,
	}
	sut := New(8, 8)
	getResult := AddHandler(sut, handler)
	t.Cleanup(sut.Close)

	vm.RegisterHooks(&vmHooks{Preprocessor: sut})
	t.Cleanup(vm.TestOnlyClearRegisteredHooks)

	stub := &hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			handler.addr: vm.NewStatefulPrecompile(func(env vm.PrecompileEnvironment, input []byte) (ret []byte, err error) {
				sdb := env.StateDB()
				txi, txh := sdb.TxIndex(), sdb.TxHash()

				// Precompiles MUST NOT charge gas for the preprocessing as it
				// would then be double-counted.
				got, ok := getResult(txi)
				if !ok {
					t.Errorf("no result for tx[%d] %v", txi, txh)
				}
				sdb.AddLog(got.Result.asLog())
				return nil, nil
			}),
		},
	}
	stub.Register(t)

	key, err := crypto.GenerateKey()
	require.NoErrorf(t, err, "crypto.GenerateKey()")
	eoa := crypto.PubkeyToAddress(key.PublicKey)

	state, evm := ethtest.NewZeroEVM(t)
	state.CreateAccount(eoa)
	state.SetBalance(eoa, new(uint256.Int).SetAllOne())

	var (
		txs  types.Transactions
		want types.Receipts
	)
	ignore := cmp.Options{
		cmpopts.IgnoreFields(
			types.Receipt{},
			"PostState", "CumulativeGasUsed", "BlockNumber", "BlockHash", "Bloom",
		),
		cmpopts.IgnoreFields(types.Log{}, "BlockHash"),
	}

	header := &types.Header{
		Number:  big.NewInt(0),
		BaseFee: big.NewInt(0),
	}
	config := evm.ChainConfig()
	rules := config.Rules(header.Number, true, header.Time)
	signer := types.MakeSigner(config, header.Number, header.Time)

	for i, addr := range []common.Address{
		{'o', 't', 'h', 'e', 'r'},
		handler.addr,
	} {
		ui := uint(i)
		data := []byte("hello, world")

		gas, err := intrinsicGas(data, types.AccessList{}, &addr, &rules)
		require.NoError(t, err, "core.IntrinsicGas(%#x, nil, false, ...)", data)
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
			want := (&recorded{
				TxData: tx.Data(),
			}).asLog()

			want.TxHash = tx.Hash()
			want.TxIndex = ui

			wantR.Logs = []*types.Log{want}
		}
		want = append(want, wantR)
	}

	block := types.NewBlock(header, txs, nil, nil, trie.NewStackTrie(nil))
	require.NoError(t, sut.StartBlock(state, rules, block), "StartBlock()")

	pool := core.GasPool(math.MaxUint64)
	var receipts types.Receipts
	for i, tx := range txs {
		state.SetTxContext(tx.Hash(), i)

		var usedGas uint64
		receipt, err := core.ApplyTransaction(
			evm.ChainConfig(),
			ethtest.DummyChainContext(),
			&block.Header().Coinbase,
			&pool,
			state,
			block.Header(),
			tx,
			&usedGas,
			vm.Config{},
		)
		require.NoError(t, err, "ApplyTransaction([%d])", i)
		receipts = append(receipts, receipt)
	}
	sut.FinishBlock(state, block, receipts)

	if diff := cmp.Diff(want, handler.gotReceipts, ignore); diff != "" {
		t.Errorf("%T diff (-want +got):\n%s", receipts, diff)
	}
}
