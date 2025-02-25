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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/params"

	_ "embed"
)

const defaultBranch = "main"

var triggerOrPRTargetBranch = flag.String(
	"analyse_branch",
	defaultBranch,
	"Target branch if triggered by a PR (github.base_ref), otherwise triggering branch (github.ref)",
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var (
	//go:embed cherrypicks
	cherryPicks  string
	lineFormatRE = regexp.MustCompile(`^([a-fA-F0-9]{40}) # (.*)$`)
)

type parsedLine struct {
	hash, commitMsg string
}

func parseCherryPicks(t *testing.T) (rawLines []string, lines []parsedLine) {
	t.Helper()
	for i, line := range strings.Split(cherryPicks, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch matches := lineFormatRE.FindStringSubmatch(line); len(matches) {
		case 3:
			rawLines = append(rawLines, line)
			lines = append(lines, parsedLine{
				hash:      matches[1],
				commitMsg: matches[2],
			})

		default:
			t.Errorf("Line %d is improperly formatted: %s", i, line)
		}
	}
	return rawLines, lines
}

func TestCherryPicksFormat(t *testing.T) {
	rawLines, lines := parseCherryPicks(t)
	if t.Failed() {
		t.Fatalf("Required line regexp: %s", lineFormatRE.String())
	}

	commits := make([]struct {
		obj  *object.Commit
		line parsedLine
	}, len(lines))

	repo := openGitRepo(t)
	for i, line := range lines {
		obj, err := repo.CommitObject(plumbing.NewHash(line.hash))
		require.NoErrorf(t, err, "%T.CommitObject(%q)", repo, line.hash)

		commits[i].obj = obj
		commits[i].line = line
	}
	sort.Slice(commits, func(i, j int) bool {
		ci, cj := commits[i].obj, commits[j].obj
		return ci.Committer.When.Before(cj.Committer.When)
	})

	var want []string
	for _, c := range commits {
		msg := strings.Split(c.obj.Message, "\n")[0]
		want = append(
			want,
			fmt.Sprintf("%s # %s", c.line.hash, msg),
		)
	}
	if diff := cmp.Diff(want, rawLines); diff != "" {
		t.Errorf("Commits in `cherrypicks` file out of order or have incorrect commit message(s);\n(-want +got):\n%s", diff)
		t.Logf("To fix, copy:\n%s", strings.Join(want, "\n"))
	}
}

func TestBranchProperties(t *testing.T) {
	branch := strings.TrimPrefix(*triggerOrPRTargetBranch, "refs/heads/")

	switch {
	case strings.HasPrefix(branch, "release/"):
		// Tests continue below
	case branch == defaultBranch:
		if rt := params.LibEVMReleaseType; rt.ForReleaseBranch() {
			t.Errorf("On default branch; params.LibEVMReleaseType = %q, which is reserved for release branches", rt)
		}
		return
	default:
		t.Skipf("Branch %q is neither default nor release branch", branch)
	}

	// Testing a release branch

	want := fmt.Sprintf("release/v%s", params.LibEVMVersion)
	assert.Equal(t, want, branch)
	if rt := params.LibEVMReleaseType; !rt.ForReleaseBranch() {
		t.Errorf("On release branch; params.LibEVMReleaseType = %q, which is unsuitable for release branches", rt)
	}

	repo := openGitRepo(t)
	headRef, err := repo.Head()
	require.NoError(t, err)

	head := commitFromRef(t, repo, headRef)
	main := branchTipCommit(t, repo, "main")

	forks, err := head.MergeBase(main)
	require.NoError(t, err)
	require.Len(t, forks, 1)
	fork := forks[0]
	t.Logf("Forked from default branch at commit %s (%s)", fork.Hash, commitMsgFirstLine(fork))

	history, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderDFS,
	})
	require.NoError(t, err)
	newCommits := commitsSince(t, history, fork)
	logCommits(t, "History since fork", newCommits)

	_, cherryPick := parseCherryPicks(t)
	wantCommits := commitsFromHashes(t, repo, cherryPick, fork)
	logCommits(t, "Expected cherry-picks", wantCommits)
	require.Len(t, newCommits, len(wantCommits)+1)

	opt := cmp.Transformer("gitCommit", pertinentCommitProperties)
	if diff := cmp.Diff(wantCommits, newCommits[:len(wantCommits)], opt); diff != "" {
		t.Error(diff)
	}

	n := len(newCommits)
	lastDiffs, err := object.DiffTree(
		treeFromCommit(t, newCommits[n-1]),
		treeFromCommit(t, newCommits[n-2]),
	)
	require.NoError(t, err)

	wantFilesModified := []string{
		"version.libevm.go",
		"version.libevm_test.go",
	}
	opt = cmpopts.SortSlices(func(a, b string) bool { return a < b })
	if diff := cmp.Diff(wantFilesModified, changedFilesByName(t, lastDiffs), opt); diff != "" {
		t.Error(diff)
	}
}

func openGitRepo(t *testing.T) *git.Repository {
	t.Helper()

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

	return repo
}

func branchTipCommit(t *testing.T, repo *git.Repository, name string) *object.Commit {
	t.Helper()

	branch, err := repo.Branch(name)
	require.NoError(t, err)
	ref, err := repo.Reference(branch.Merge, false)
	require.NoError(t, err)
	return commitFromRef(t, repo, ref)
}

func commitFromRef(t *testing.T, repo *git.Repository, ref *plumbing.Reference) *object.Commit {
	t.Helper()
	c, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	return c
}

func commitsSince(t *testing.T, iter object.CommitIter, since *object.Commit) []*object.Commit {
	t.Helper()

	var commits []*object.Commit
	errReachedSince := fmt.Errorf("%T reached terminal commit %v", iter, since)

	err := iter.ForEach(func(c *object.Commit) error {
		if c.Hash == since.Hash {
			return errReachedSince
		}
		commits = append(commits, c)
		return nil
	})
	require.ErrorIs(t, err, errReachedSince)

	slices.Reverse(commits)
	return commits
}

func commitsFromHashes(t *testing.T, repo *git.Repository, lines []parsedLine, skipAncestorsOf *object.Commit) []*object.Commit {
	t.Helper()

	var commits []*object.Commit
	for _, l := range lines {
		c, err := repo.CommitObject(plumbing.NewHash(l.hash))
		require.NoError(t, err)

		skip, err := c.IsAncestor(skipAncestorsOf)
		require.NoError(t, err)
		if skip || c.Hash == skipAncestorsOf.Hash {
			continue
		}
		commits = append(commits, c)
	}

	return commits
}

func commitMsgFirstLine(c *object.Commit) string {
	return strings.Split(c.Message, "\n")[0]
}

func logCommits(t *testing.T, header string, commits []*object.Commit) {
	t.Logf("### %s (%d commits):", header, len(commits))
	for _, c := range commits {
		t.Logf("%s by %s", commitMsgFirstLine(c), c.Author.String())
	}
}

type comparableCommit struct {
	MessageFirstLine, Author string
	Authored                 time.Time
}

func pertinentCommitProperties(c *object.Commit) comparableCommit {
	return comparableCommit{
		MessageFirstLine: commitMsgFirstLine(c),
		Author:           c.Author.String(),
		Authored:         c.Author.When,
	}
}

func treeFromCommit(t *testing.T, commit *object.Commit) *object.Tree {
	t.Helper()
	tree, err := commit.Tree()
	require.NoError(t, err)
	return tree
}

func changedFilesByName(t *testing.T, changes object.Changes) []string {
	t.Helper()

	var files []string
	for _, c := range changes {
		from, to, err := c.Files()
		require.NoError(t, err)
		require.NotNilf(t, from, "file %q inserted", to.Name)
		require.NotNilf(t, to, "file %q deleted", from.Name)
		require.Equalf(t, from.Name, to.Name, "file renamed; expect modified file's name to equal original")

		// [object.File.Name] is documented as being either the name or a path,
		// depending on how it was generated. We only need to protect against
		// accidental changes to the wrong files, so it's sufficient to just
		// check the names.
		files = append(files, filepath.Base(from.Name))
	}
	return files
}
