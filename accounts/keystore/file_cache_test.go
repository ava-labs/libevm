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

package keystore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

// TestFileCacheScan_inPlaceSizeChangeWithoutAfterLastMod ensures a key file is
// treated as updated when its size changes even if ModTime is not strictly after
// the previous scan's global lastMod (e.g. same wall-clock second, or tests that
// bump lastMod artificially).
func TestFileCacheScan_inPlaceSizeChangeWithoutAfterLastMod(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aaa")
	if err := os.WriteFile(path, []byte("short"), 0600); err != nil {
		t.Fatal(err)
	}

	fc := &fileCache{
		all:      mapset.NewThreadUnsafeSet[string](),
		fileStat: make(map[string]fileStat),
	}
	if _, _, _, err := fc.scan(dir); err != nil {
		t.Fatal(err)
	}

	// Make the global lastMod lie in the future so ModTime.After(lastMod) is false,
	// while the path is still known (in-place edit scenario).
	fc.lastMod = time.Now().Add(1 * time.Hour)

	if err := os.WriteFile(path, []byte("muuuuuuuuch longer content"), 0600); err != nil {
		t.Fatal(err)
	}

	creates, deletes, updates, err := fc.scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if creates.Cardinality() != 0 || deletes.Cardinality() != 0 {
		t.Fatalf("creates=%d deletes=%d; want 0,0", creates.Cardinality(), deletes.Cardinality())
	}
	if updates.Cardinality() != 1 || !updates.Contains(path) {
		t.Fatalf("updates=%v; want single path %q", updates, path)
	}
}

func TestFileCacheScan_noSpuriousUpdateWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aaa")
	content := []byte("fixed content")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	fc := &fileCache{
		all:      mapset.NewThreadUnsafeSet[string](),
		fileStat: make(map[string]fileStat),
	}
	if _, _, _, err := fc.scan(dir); err != nil {
		t.Fatal(err)
	}
	creates, deletes, updates, err := fc.scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if creates.Cardinality() != 0 || deletes.Cardinality() != 0 || updates.Cardinality() != 0 {
		t.Fatalf("creates=%d deletes=%d updates=%d; want all 0",
			creates.Cardinality(), deletes.Cardinality(), updates.Cardinality())
	}
}
