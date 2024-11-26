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
	"fmt"

	"github.com/ava-labs/libevm/common"
)

// A LibEVMOption configures default behaviour of this package.
type LibEVMOption interface {
	apply(*libevmConfig)
}

type libevmConfig struct {
	preserveDescendantsOnCapZero bool
	hash, parentHash             common.Hash
}

func asLibEVMConfig(opts []LibEVMOption) *libevmConfig {
	c := new(libevmConfig)
	for _, o := range opts {
		o.apply(c)
	}
	return c
}

type libevmFuncOpt func(*libevmConfig)

func (f libevmFuncOpt) apply(c *libevmConfig) { f(c) }

// PreserveDescendantsOnCapZero signals to [Tree.Cap], if capping to zero layers
// (i.e. flattening), that descendants of the flattened layer must be kept.
// Without this option, the entire tree is replaced with the new base (disk)
// layer.
func PreserveDescendantsOnCapZero() LibEVMOption {
	return libevmFuncOpt(func(c *libevmConfig) {
		c.preserveDescendantsOnCapZero = true
	})
}

func WithBlockHashes(hash, parentHash common.Hash) LibEVMOption {
	return libevmFuncOpt(func(c *libevmConfig) {
		c.hash = hash
		c.parentHash = parentHash
	})
}

func (t *Tree) updateLayersAfterCapZero(base *diskLayer, flattened *diffLayer, opts ...LibEVMOption) {
	// Original geth behaviour
	opt := asLibEVMConfig(opts)
	if !opt.preserveDescendantsOnCapZero {
		t.layers = map[common.Hash]snapshot{base.root: base}
		t.children = make(map[common.Hash][]common.Hash)
		return
	}

	children := t.children
	baseHash := base.root
	if opt.hash != (common.Hash{}) {
		baseHash = opt.hash
	}

	newLayers := map[common.Hash]snapshot{baseHash: base}
	newChildren := make(map[common.Hash][]common.Hash)
	var keepChildren func(root common.Hash)
	keepChildren = func(root common.Hash) {
		for _, child := range children[root] {
			childLayer := t.layers[child]
			newLayers[child] = childLayer
			newChildren[root] = append(newChildren[root], child)
			keepChildren(child)
		}
	}
	keepChildren(baseHash)

	for _, child := range children[baseHash] {
		d, ok := t.layers[child].(*diffLayer)
		if !ok {
			continue
		}
		d.lock.Lock()
		d.parent = base
		d.lock.Unlock()
	}
	t.layers = newLayers
	t.children = newChildren
}

func (t *Tree) Flatten(hash common.Hash) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Retrieve the head snapshot to cap from
	snap := t.layers[hash]
	if snap == nil {
		return fmt.Errorf("snapshot [%#x] missing", hash)
	}
	diff, ok := snap.(*diffLayer)
	if !ok {
		return fmt.Errorf("snapshot [%#x] is disk layer", hash)
	}

	diff.lock.RLock()
	base := diffToDisk(diff.flatten().(*diffLayer))
	diff.lock.RUnlock()

	t.updateLayersAfterCapZero(
		base,
		diff,
		PreserveDescendantsOnCapZero(),
		WithBlockHashes(hash, common.Hash{}),
	)
	return nil
}

func (t *Tree) Discard(hash common.Hash) error {
	return nil
}

func (t *Tree) AbortGeneration() {
	t.lock.Lock()
	defer t.lock.Unlock()

	dl := t.disklayer()

	dl.lock.Lock()
	if dl.genAbort == nil {
		dl.lock.Unlock()
		return
	}
	if dl.genAbort != nil {
		abort := make(chan *generatorStats)
		dl.genAbort <- abort
		dl.genAbort = nil
		dl.lock.Unlock()
		<-abort
	}
}

func (t *Tree) NumStateLayers() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.layers)
}

func (t *Tree) NumBlockLayers() int {
	return t.NumStateLayers()
}
func (t *Tree) DiskAccountIterator(seek common.Hash) AccountIterator {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.disklayer().AccountIterator(seek)
}

func (t *Tree) DiskStorageIterator(account common.Hash, seek common.Hash) StorageIterator {
	t.lock.Lock()
	defer t.lock.Unlock()

	it, _ := t.disklayer().StorageIterator(account, seek)
	return it
}
