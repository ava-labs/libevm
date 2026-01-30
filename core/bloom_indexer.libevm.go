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
// <http://www.gnu.org/licenses/>

package core

import (
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb"
)

// BloomThrottling is the time to wait between processing two consecutive index sections.
const BloomThrottling = bloomThrottling

func NewBloomIndexerBackend(db ethdb.Database, size uint64) *BloomIndexer {
	return &BloomIndexer{
		db:   db,
		size: size,
	}
}

// ProcessBloom is the same as Process, but takes the header and bloom separately.
func (b *BloomIndexer) ProcessBloom(header *types.Header, bloom types.Bloom) error {
	b.gen.AddBloom(uint(header.Number.Uint64()-b.section*b.size), bloom)
	b.head = header.Hash()
	return nil
}
