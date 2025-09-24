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
	"bytes"
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

type reverser struct {
	headerExtra []byte
	addr        common.Address
	stateKey    common.Hash
	gas         uint64
}

func (r *reverser) BeforeBlock(h *types.Header) {
	r.headerExtra = slices.Clone(h.Extra)
}

func (r *reverser) Gas(tx *types.Transaction) (uint64, bool) {
	if to := tx.To(); to == nil || *to != r.addr {
		return 0, false
	}
	return r.gas, true
}

func reverserOutput(txData []byte, state common.Hash, extra []byte) []byte {
	out := slices.Concat(txData, state[:], extra)
	slices.Reverse(out)
	return out
}

func (r *reverser) Process(sdb libevm.StateReader, i int, tx *types.Transaction) []byte {
	return reverserOutput(
		tx.Data(),
		sdb.GetTransientState(r.addr, r.stateKey),
		r.headerExtra,
	)
}

func TestProcessor(t *testing.T) {
	handler := &reverser{
		addr:     common.Address{'r', 'e', 'v', 'e', 'r', 's', 'e'},
		stateKey: common.Hash{'k', 'e', 'y'},
		gas:      1e6,
	}
	p := New(handler, 8)
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
	stateVal := common.Hash{'s', 't', 'a', 't', 'e'}
	sdb.SetTransientState(handler.addr, handler.stateKey, stateVal)

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
				require.NoError(t, err, "core.IntrinsicGas(%#x, nil, false, true, true, true)", data)

				txs[i] = types.NewTx(&types.LegacyTx{
					To:   &to,
					Data: data,
					Gas:  gas + extraGas,
				})
			}

			extra := []byte("extra")
			block := types.NewBlock(&types.Header{Extra: extra}, txs, nil, nil, trie.NewStackTrie(nil))
			require.NoError(t, p.StartBlock(block, rules, sdb), "StartBlock()")
			defer p.FinishBlock(block)

			for i, tx := range txs {
				wantOK := wantProcessed[i]

				var want []byte
				if wantOK {
					want = reverserOutput(tx.Data(), stateVal, extra)
				}

				got, gotOK := p.Result(i)
				if !bytes.Equal(got, want) || gotOK != wantOK {
					t.Errorf("Result(%d) got (%#x, %t); want (%#x, %t)", i, got, gotOK, want, wantOK)
				}
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
	handler := &reverser{
		addr: common.Address{'r', 'e', 'v', 'e', 'r', 's', 'e'},
		gas:  handlerGas,
	}
	sut := New(handler, 8)
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
				got, ok := sut.Result(txi)
				if !ok {
					t.Errorf("no result for tx[%d] %v", txi, txh)
				}
				sdb.AddLog(&types.Log{
					Data: got[:],
				})
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
		want []*types.Receipt
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
			wantR.Logs = []*types.Log{{
				TxHash:  tx.Hash(),
				TxIndex: ui,
				Data:    reverserOutput(data, common.Hash{}, nil),
			}}
		}
		want = append(want, wantR)
	}

	block := types.NewBlock(header, txs, nil, nil, trie.NewStackTrie(nil))
	require.NoError(t, sut.StartBlock(block, rules, state), "StartBlock()")
	defer sut.FinishBlock(block)

	pool := core.GasPool(math.MaxUint64)
	var got []*types.Receipt
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
		got = append(got, receipt)
	}

	if diff := cmp.Diff(want, got, ignore); diff != "" {
		t.Errorf("%T diff (-want +got):\n%s", got, diff)
	}
}
