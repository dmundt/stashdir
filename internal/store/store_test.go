package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// setTestConfigDir redirects os.UserConfigDir() lookups to a temp directory
// appropriate for the current OS so tests do not touch real user config.
func setTestConfigDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmp)
	case "darwin":
		// On macOS, UserConfigDir = $HOME/Library/Application Support
		t.Setenv("HOME", tmp)
	default:
		// On Unix, UserConfigDir prefers $XDG_CONFIG_HOME
		t.Setenv("XDG_CONFIG_HOME", tmp)
	}
	return tmp
}

func TestAddListPersistence(t *testing.T) {
	setTestConfigDir(t)

	db, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	if err := db.Add(filepath.Clean("/tmp/a")); err != nil {
		t.Fatalf("add a: %v", err)
	}
	if err := db.Add(filepath.Clean("/tmp/b")); err != nil {
		t.Fatalf("add b: %v", err)
	}

	got := db.List()
	if len(got) != 2 {
		t.Fatalf("list len=%d, want 2; got=%v", len(got), got)
	}

	// Ensure it saved to disk
	if _, err := os.Stat(db.Path); err != nil {
		t.Fatalf("expected db file to exist at %s: %v", db.Path, err)
	}

	// Reload and verify persistence
	db2, err := Open()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got2 := db2.List()
	if len(got2) != 2 {
		t.Fatalf("reloaded list len=%d, want 2; got=%v", len(got2), got2)
	}
}

func TestOrderAndRemoveIndex_Sorted(t *testing.T) {
	setTestConfigDir(t)
	db, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Add out of order, expect sorted storage
	raw := []string{filepath.Clean("/p3"), filepath.Clean("/p1"), filepath.Clean("/p2")}
	for _, p := range raw {
		if err := db.Add(p); err != nil {
			t.Fatalf("add %s: %v", p, err)
		}
	}

	got := db.List()
	want := []string{filepath.Clean("/p1"), filepath.Clean("/p2"), filepath.Clean("/p3")}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("order mismatch: got=%v want=%v", got, want)
	}

	if err := db.RemoveIndex(1); err != nil { // remove second
		t.Fatalf("remove index 1: %v", err)
	}
	got = db.List()
	if len(got) != 2 || got[0] != want[0] || got[1] != want[2] {
		t.Fatalf("after remove index, got=%v", got)
	}
}

func TestPersistentSortedOnLoad(t *testing.T) {
	setTestConfigDir(t)
	db, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// Add out-of-order and ensure it's saved sorted; then reopen and validate order
	_ = db.Add(filepath.Clean("/zeta"))
	_ = db.Add(filepath.Clean("/alpha"))
	_ = db.Add(filepath.Clean("/beta"))

	// Reopen
	db2, err := Open()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got := db2.List()
	want := []string{filepath.Clean("/alpha"), filepath.Clean("/beta"), filepath.Clean("/zeta")}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch at %d: got=%v want=%v", i, got, want)
		}
	}
}

func TestRemovePathCaseSensitivity(t *testing.T) {
	setTestConfigDir(t)
	db, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Choose a path with mixed case and platform-specific separators
	base := filepath.Join(string(filepath.Separator), "Var", "Data", "Proj")
	if err := db.Add(base); err != nil {
		t.Fatalf("add: %v", err)
	}

	var toRemove string
	if runtime.GOOS == "windows" {
		// On Windows, removal should be case-insensitive; flip the case to test
		toRemove = swapCase(base)
	} else {
		// On non-Windows, must match exactly
		toRemove = base
	}

	if err := db.RemovePath(toRemove); err != nil {
		if runtime.GOOS == "windows" {
			t.Fatalf("windows RemovePath should succeed case-insensitively: %v", err)
		}
		// On non-windows should have succeeded because exact
		if runtime.GOOS != "windows" {
			t.Fatalf("RemovePath failed: %v", err)
		}
	}
}

func swapCase(s string) string {
	r := []rune(s)
	for i := range r {
		if r[i] >= 'a' && r[i] <= 'z' {
			r[i] = r[i] - 32
		} else if r[i] >= 'A' && r[i] <= 'Z' {
			r[i] = r[i] + 32
		}
	}
	return string(r)
}

func TestSelectInteractiveEmptyList(t *testing.T) {
	setTestConfigDir(t)
	db, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// Ensure empty list
	if len(db.List()) != 0 {
		t.Fatalf("expected empty list at start")
	}
	// Should not block or error; returns empty choice
	choice, err := db.SelectInteractive()
	if err != nil {
		t.Fatalf("SelectInteractive on empty: %v", err)
	}
	if choice != "" {
		t.Fatalf("expected empty choice, got %q", choice)
	}
}

func TestDBFileJSONShape(t *testing.T) {
	setTestConfigDir(t)
	db, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = db.Add(filepath.Clean("/alpha"))
	_ = db.Add(filepath.Clean("/beta"))

	// Read the file and ensure it has a valid JSON with items array
	b, err := os.ReadFile(db.Path)
	if err != nil {
		t.Fatalf("read db file: %v", err)
	}
	var raw struct {
		Items []string `json:"items"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v\njson: %s", err, string(b))
	}
	if len(raw.Items) != 2 {
		t.Fatalf("json items len=%d want 2", len(raw.Items))
	}
}
