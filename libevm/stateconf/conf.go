// Copyright 2024 the libevm authors.
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

// Package stateconf configures state management.
package stateconf

import (
	"github.com/ava-labs/libevm/libevm/options"
)

// A StateDBCommitOption configures the behaviour of all update calls within the
// state.StateDB.Commit() implementations.
type StateDBCommitOption = options.Option[stateDBCommitConfig]

type stateDBCommitConfig struct {
	snapshotOpts []SnapshotUpdateOption
	triedbOpts   []TrieDBUpdateOption
}

// WithSnapshotUpdateOpts returns a StateDBCommitOption carrying a list of
func WithSnapshotUpdateOpts(opts ...SnapshotUpdateOption) StateDBCommitOption {
	return options.Func[stateDBCommitConfig](func(c *stateDBCommitConfig) {
		// I don't like append() because there's no way to remove options, but that's a weakly held opinion
		c.snapshotOpts = opts
	})
}

func ExtractSnapshotUpdateOpts(opts ...StateDBCommitOption) []SnapshotUpdateOption {
	return options.As(opts...).snapshotOpts
}

func WithTrieDBUpdateOpts(opts ...TrieDBUpdateOption) StateDBCommitOption {
	return options.Func[stateDBCommitConfig](func(c *stateDBCommitConfig) {
		// I don't like append() because there's no way to remove options, but that's a weakly held opinion
		c.triedbOpts = opts
	})
}

func ExtractTrieDBUpdateOpts(opts ...StateDBCommitOption) []TrieDBUpdateOption {
	return options.As(opts...).triedbOpts
}

// A SnapshotUpdateOption configures the behaviour of
// state.SnapshotTree.Update() implementations. This will be removed along with
// state.SnapshotTree.
type SnapshotUpdateOption = options.Option[snapshotUpdateConfig]

type snapshotUpdateConfig struct {
	payload any
}

// WithSnapshotUpdatePayload returns a SnapshotUpdateOption carrying an arbitrary
// payload. It acts only as a carrier to exploit existing function plumbing and
// the effect on behaviour is left to the implementation receiving it.
func WithSnapshotUpdatePayload(p any) SnapshotUpdateOption {
	return options.Func[snapshotUpdateConfig](func(c *snapshotUpdateConfig) {
		c.payload = p
	})
}

// ExtractSnapshotUpdatePayload returns the payload carried by a [WithSnapshotUpdatePayload]
// option. Only one such option can be used at once; behaviour is otherwise
// undefined.
func ExtractSnapshotUpdatePayload(opts ...SnapshotUpdateOption) any {
	return options.As(opts...).payload
}

type TrieDBUpdateOption = options.Option[triedbUpdateConfig]

type triedbUpdateConfig struct {
	payload any
}

func WithTrieDBUpdatePayload(p any) TrieDBUpdateOption {
	return options.Func[triedbUpdateConfig](func(c *triedbUpdateConfig) {
		c.payload = p
	})
}

// ExtractTrieDBUpdatePayload returns the payload carried by a [WithSnapshotUpdatePayload]
// option. Only one such option can be used at once; behaviour is otherwise
// undefined.
func ExtractTrieDBUpdatePayload(opts ...TrieDBUpdateOption) any {
	return options.As(opts...).payload
}
