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
	"math"

	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/bloombits"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/eth"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/params"
)

// BloomIndexerService tracks all necessary components to run a bloom indexer
// service alongside the Ethereum node, independent of the [eth.Ethereum] struct.
// The methods returned can be used to implement the [Backend] interface, but
// this CANNOT be embedded into the backend struct directly, as it would
// expose the [BloomIndexerService.Close] method publicly. The Close method must be called once
// the service is no longer needed to gracefully terminate all goroutines.
type BloomIndexerService struct {
	indexer  *core.ChainIndexer
	size     uint64
	requests chan chan *bloombits.Retrieval
	quit     chan struct{}
}

// NewBloomIndexerService creates and starts a bloom indexer service with the given
// backend and section size. If the section size is 0 or too large, it defaults
// to [params.BloomBitsBlocks].
// The returned service immediately starts indexing the canonical chain and
// servicing bloom filter retrieval requests.
// Once done, the service should be closed with [CloseBloomIndexerService].
// The [BloomOverrider] MAY be nil, in which case the [types.Header] bloom is
// always used.
func NewBloomIndexerService(db ethdb.Database, chain core.ChainIndexerChain, override BloomOverrider, size uint64) *BloomIndexerService {
	if size == 0 || size > math.MaxInt32 {
		size = params.BloomBitsBlocks
	}
	backend := &bloomBackend{
		BloomIndexer: core.NewBloomIndexerBackend(db, size),
		override:     override,
	}
	table := rawdb.NewTable(db, string(rawdb.BloomBitsIndexPrefix))
	s := &BloomIndexerService{
		indexer:  core.NewChainIndexer(db, table, backend, size, 0, core.BloomThrottling, "bloombits"),
		size:     size,
		requests: make(chan chan *bloombits.Retrieval),
		quit:     make(chan struct{}),
	}

	s.indexer.Start(chain)
	eth.StartBloomHandlers(
		db,
		s.requests,
		s.quit,
		size,
	)

	return s
}

// BloomStatus returns the section size and the number of sections indexed so far.
// Can be used as a [Backend] implementation.
func (s *BloomIndexerService) BloomStatus() (uint64, uint64) {
	sections, _, _ := s.indexer.Sections()
	return s.size, sections
}

// ServiceFilter starts servicing bloom filter retrieval requests for the given
// matcher session. Can be used as a [Backend] implementation.
func (s *BloomIndexerService) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for range eth.BloomFilterThreads {
		go session.Multiplex(eth.BloomRetrievalBatch, eth.BloomRetrievalWait, s.requests)
	}
}

// Close terminates the bloom indexer, current bloom filter retrieval requests,
// and the bloom retrieval server. It is defined as a function instead of a
// method to allow embedding of a [BloomIndexerService] without exposing it as
// an RPC method.
func CloseBloomIndexerService(s *BloomIndexerService) error {
	close(s.quit)
	return s.indexer.Close()
}

// BloomOverrider is an optional extension to [Backend], allowing arbitrary
// bloom filters to be returned for a header. If not implemented,
// [types.Header.Bloom] is used instead.
type BloomOverrider interface {
	OverrideHeaderBloom(*types.Header) types.Bloom
}

func maybeOverrideBloom(header *types.Header, backend Backend) types.Bloom {
	if bo, ok := backend.(BloomOverrider); ok {
		return bo.OverrideHeaderBloom(header)
	}
	return header.Bloom
}

var _ core.ChainIndexerBackend = (*bloomBackend)(nil)

// bloomBackend is a wrapper around a [core.BloomIndexer] that
// overrides Process() to allow for custom bloom filter generation.
type bloomBackend struct {
	*core.BloomIndexer
	override BloomOverrider
}

func (b *bloomBackend) bloom(h *types.Header) types.Bloom {
	if b.override == nil {
		return h.Bloom
	}
	return b.override.OverrideHeaderBloom(h)
}

// Process adds a new header's bloom into the index, possibly overriding
// it using the backend's [BloomOverrider] implementation.
func (b *bloomBackend) Process(ctx context.Context, header *types.Header) error {
	return b.ProcessWithBloomOverride(header, b.bloom(header))
}
