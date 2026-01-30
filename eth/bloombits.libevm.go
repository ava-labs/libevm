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

package eth

import (
	"github.com/ava-labs/libevm/core/bloombits"
	"github.com/ava-labs/libevm/ethdb"
)

// StartBloomHandlers starts a batch of goroutines to accept bloom bit database
// retrievals from possibly a range of filters and serving the data to satisfy.
// This is identical to [Ethereum.startBloomHandlers], but exposed for use separately.
func StartBloomHandlers(db ethdb.Database, bloomRequests chan chan *bloombits.Retrieval, closeBloomHandler chan struct{}, sectionSize uint64) {
	(&Ethereum{
		bloomRequests:     bloomRequests,
		closeBloomHandler: closeBloomHandler,
		chainDb:           db,
	}).startBloomHandlers(sectionSize)
}
