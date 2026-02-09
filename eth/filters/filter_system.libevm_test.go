// Copyright 2026 the libevm authors.
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

package filters

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/consensus/ethash"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/bloombits"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/event"
	"github.com/ava-labs/libevm/params"
)

type overrideBloomsTestBackend struct {
	*testBackend
	blockFeed event.FeedOf[core.ChainHeadEvent]
	s         *BloomIndexerService
}

// SubscribeChainHeadEvent forwards accepted blocks to the [core.ChainIndexer].
func (o *overrideBloomsTestBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return o.blockFeed.Subscribe(ch)
}

func (o *overrideBloomsTestBackend) BloomStatus() (uint64, uint64) {
	return o.s.BloomStatus()
}

func (o *overrideBloomsTestBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	o.s.ServiceFilter(ctx, session)
}

type bloomOverrider struct {
	blooms map[uint64]types.Bloom
}

// OverrideHeaderBloom replaces the bloom of the given header if we know a custom one for its block number.
// This is because [core.GenerateChainWithGenesis] doesn't let us set blooms directly.
func (o *bloomOverrider) OverrideHeaderBloom(hdr *types.Header) types.Bloom {
	bloom, ok := o.blooms[hdr.Number.Uint64()]
	if !ok {
		return hdr.Bloom
	}
	return bloom
}

func TestOverrideBlooms(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent(), goleak.IgnoreTopFunction("github.com/ava-labs/libevm/eth/filters.(*EventSystem).eventLoop"))

	const (
		sectionSize = 8
		numBlocks   = 10
	)

	var (
		db = rawdb.NewMemoryDatabase()
	)

	var (
		signer  = types.HomesteadSigner{}
		key, _  = crypto.GenerateKey()
		addr    = crypto.PubkeyToAddress(key.PublicKey)
		genesis = &core.Genesis{Config: params.TestChainConfig,
			Alloc: types.GenesisAlloc{
				addr: {Balance: big.NewInt(params.Ether)},
			},
		}
		firstAddr = common.HexToAddress("0x1111111111111111111111111111111111111111")

		receipts []*types.Receipt
		blooms   map[uint64]types.Bloom = make(map[uint64]types.Bloom)
	)

	// [core.GenerateChainWithGenesis] doesn't let us set blooms directly,
	// so we create receipts with logs that produce the desired blooms.
	// This also tests BloomOverrider in the process.
	for i := range numBlocks {
		blockNum := uint64(i + 1) //nolint:gosec // guaranteed by loop
		log := &types.Log{Address: firstAddr, Topics: []common.Hash{}, Data: []byte{}, BlockNumber: blockNum, Index: 0}
		receipt := &types.Receipt{
			Logs: []*types.Log{log},
		}
		bloom := types.CreateBloom(types.Receipts{receipt})
		blooms[blockNum] = bloom
		receipt.Bloom = bloom
		receipts = append(receipts, receipt)
	}

	// Doesn't set the bloom
	_, blocks, _ := core.GenerateChainWithGenesis(genesis, ethash.NewFaker(), numBlocks, func(i int, b *core.BlockGen) {
		b.AddUncheckedReceipt(receipts[i])
		tx, _ := types.SignTx(types.NewTx(&types.LegacyTx{
			Nonce:    uint64(i), //nolint:gosec // verified above
			To:       &common.Address{},
			Value:    big.NewInt(1000),
			Gas:      params.TxGas,
			GasPrice: b.BaseFee(),
			Data:     nil,
		}), signer, key)
		b.AddTx(tx)
	})

	writeBlock := func(block *types.Block) {
		rawdb.WriteBlock(db, block)
		rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
		rawdb.WriteHeadBlockHash(db, block.Hash())
	}

	// Write genesis block to start bloom indexer from there
	writeBlock(genesis.ToBlock())

	var (
		backend = &overrideBloomsTestBackend{
			testBackend: &testBackend{db: db},
		}
		sys = NewFilterSystem(backend, Config{})
		api = NewFilterAPI(sys, true)
	)
	// TODO(arr4n): DO NOT MERGE: this circular dependency needs to be addressed.
	backend.s = NewBloomIndexerService(db, backend, &bloomOverrider{blooms}, sectionSize)
	defer func() {
		CloseAPI(api)
		require.NoError(t, CloseBloomIndexerService(backend.s))
	}()

	for i, block := range blocks {
		writeBlock(block)
		rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), []*types.Receipt{receipts[i]})
		backend.blockFeed.Send(core.ChainHeadEvent{Block: block})
	}

	require.Eventually(t, func() bool {
		_, indexedSections := backend.BloomStatus()
		return indexedSections > 0
	}, 10*time.Second, 100*time.Millisecond, "bloom indexer did not index all blocks in time")

	filter := FilterCriteria{Addresses: []common.Address{firstAddr}, FromBlock: big.NewInt(1), ToBlock: big.NewInt(int64(numBlocks + 1))}
	results, err := api.GetLogs(t.Context(), filter)
	require.NoErrorf(t, err, "%T.GetLogs", api)
	require.Len(t, results, numBlocks, "expected to find logs for all blocks")
}
