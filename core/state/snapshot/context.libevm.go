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

package snapshot

import (
	"errors"

	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/libevm/options"
)

// generatorPausing is embedded in [generatorContext] to add libevm-specific
// fields that allow [diskLayer.stopGeneration] to unblock more frequently.
type generatorPausing struct {
	cancel   <-chan struct{} // Closed when the generation is asked to stop
	progress []byte          // Last position the generation finished, as passed to checkAndFlush
}

type generatorContextOption = options.Option[generatorContext]

// withCancelFromDiskLayer configures [newGeneratorContext] to use the same
// cancellation channel as the [diskLayer].
func withCancelFromDiskLayer(dl *diskLayer) generatorContextOption {
	return options.Func[generatorContext](func(ctx *generatorContext) {
		ctx.generatorPausing.cancel = dl.cancel
	})
}

// errLibEVMIteratorAborted is returned by an [abortableIterator] once
// cancelled. It differs from [errAborted] so that [diskLayer.generate] knows
// the work since the last [diskLayer.checkAndFlush] is yet to be saved.
var errLibEVMIteratorAborted = errors.New("snapshot iterator aborted")

// abortableIterator stops with [errLibEVMIteratorAborted] once cancel is
// closed. On a hash-scheme database the snapshot iterators skip every trie node
// whose hash starts with the snapshot prefix, so one Next can run for minutes
// while stopGeneration waits.
type abortableIterator struct {
	ethdb.Iterator
	cancel <-chan struct{}
	err    error
}

func newAbortableIterator(it ethdb.Iterator, cancel <-chan struct{}) ethdb.Iterator {
	return &abortableIterator{Iterator: it, cancel: cancel}
}

func (it *abortableIterator) Next() bool {
	if it.err != nil {
		return false
	}
	select {
	case <-it.cancel:
		it.err = errLibEVMIteratorAborted
		return false
	default:
		return it.Iterator.Next()
	}
}

func (it *abortableIterator) Error() error {
	if it.err != nil {
		return it.err
	}
	return it.Iterator.Error()
}

// flushAfterIteratorAbort converts an [errLibEVMIteratorAborted] into the
// outcome had [diskLayer.checkAndFlush] itself seen the cancellation request,
// saving the work finished up to ctx.progress. Past it the batch holds only
// deletions of entries missing from the trie, which are safe to keep.
//
// This method never returns nil.
func (dl *diskLayer) flushAfterIteratorAbort(ctx *generatorContext) error {
	if ctx.progress == nil {
		// Nothing new to save, and calling [diskLayer.checkAndFlush] with a nil
		// argument would mark generation as completed, both via
		// [diskLayer.genMarker] and [journalProgress].
		return errAborted
	}
	// We can only reach here if the [abortableIterator] saw [diskLayer.cancel]
	// being closed. [diskLayer.checkAndFlush] will therefore always be
	// `aborting`.
	return dl.checkAndFlush(ctx, ctx.progress)
}
