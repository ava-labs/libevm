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

package release

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"

	_ "embed"
)

var (
	//go:embed cherrypicks
	cherryPicks  string
	lineFormatRE = regexp.MustCompile(`^([a-fA-F0-9]{40}) #.+$`)
)

func TestCherryPicksFormat(t *testing.T) {
	var commits []string

	for i, line := range strings.Split(cherryPicks, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch matches := lineFormatRE.FindStringSubmatch(line); len(matches) {
		case 2:
			commits = append(commits, matches[1])
		default:
			t.Errorf("Line %d is improperly formatted: %s", i, line)
		}
	}
	if t.Failed() {
		t.Fatalf("Required line regexp: %s", lineFormatRE.String())
	}

	opts := &git.PlainOpenOptions{DetectDotGit: true}
	repo, err := git.PlainOpenWithOptions("./", opts)
	require.NoErrorf(t, err, "git.PlainOpenWithOptions(./, %+v", opts)

	fetch := &git.FetchOptions{
		RemoteURL: "https://github.com/ethereum/go-ethereum.git",
	}
	err = repo.Fetch(fetch)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		t.Fatalf("%T.Fetch(%+v) error %v", repo, fetch, err)
	}

	var (
		lastHash string
		lastAt   time.Time
	)
	for _, hash := range commits {
		obj, err := repo.CommitObject(plumbing.NewHash(hash))
		require.NoErrorf(t, err, "%T.CommitObject(%q)", repo, hash)

		at := obj.Committer.When
		if !at.After(lastAt) {
			t.Errorf("Commit %s (%s) is not after %s (%s)", hash, at, lastHash, lastAt)
		}
		lastHash = hash
		lastAt = at
	}
}
