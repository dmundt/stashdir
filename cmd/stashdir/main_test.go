package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// setTestConfigEnv ensures the CLI uses a temp config directory
func setTestConfigEnv(t *testing.T) string {
	t.Helper()
	cfg := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", cfg)
	case "darwin":
		t.Setenv("HOME", cfg)
	default:
		t.Setenv("XDG_CONFIG_HOME", cfg)
	}
	return cfg
}

// runMain invokes main() with the provided args, optionally changing working directory,
// and captures stdout/stderr. Returns stdout, stderr, and an exit code (0 on success).
// Note: if main() calls log.Fatalf, the process exits; tests are written to avoid that path.
func runMain(t *testing.T, workDir string, args ...string) (string, string, int) {
	t.Helper()
	// Save/restore cwd
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("chdir failed: %v", err)
		}
		defer func() { _ = os.Chdir(oldWD) }()
	}

	// Capture stdout/stderr
	oldStdout, oldStderr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()

	// Set args
	oldArgs := os.Args
	os.Args = append([]string{"stashdir"}, args...)
	defer func() { os.Args = oldArgs }()

	// Run main
	code := 0
	main()

	// Close writers and read
	_ = wOut.Close()
	_ = wErr.Close()
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	_ = rOut.Close()
	_ = rErr.Close()
	return string(outBytes), string(errBytes), code
}

func normForOS(p string) string {
	// Always resolve symlinks for expected path to match CLI output
	rp := evalSymlinksIfPossible(p)
	if runtime.GOOS == "windows" {
		rp = strings.ToLower(rp)
	}
	return rp
}

func evalSymlinksIfPossible(p string) string {
	if rp, err := filepath.EvalSymlinks(p); err == nil {
		return rp
	}
	return p
}

func TestCLI_AddListSelectRemove_ByIndex(t *testing.T) {
	setTestConfigEnv(t)

	// Use a unique working directory so add . resolves predictably
	wd := t.TempDir()

	// Add current directory using '.'
	if out, errOut, code := runMain(t, wd, "add", "."); code != 0 {
		t.Fatalf("add . exit=%d stderr=%s stdout=%s", code, errOut, out)
	}

	// List should show exactly one entry
	out, errOut, code := runMain(t, "", "list")
	if code != 0 {
		t.Fatalf("list exit=%d err=%s", code, errOut)
	}

	// On macOS, temp dirs may appear as /var/... but getwd resolves to /private/var/...
	// Match the CLI's resolved form by evaluating symlinks for the expected path.
	expectedPath := normForOS(evalSymlinksIfPossible(filepath.Clean(wd)))
	// Parse the first line and extract the path after the tab
	firstLine := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")[0]
	parts := strings.SplitN(firstLine, "\t", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected list output format: %q", out)
	}
	listedPath := strings.TrimSpace(parts[1])
	if runtime.GOOS == "windows" {
		// On Windows, handle 8.3 short names by comparing file identity
		ei, err1 := os.Stat(expectedPath)
		li, err2 := os.Stat(listedPath)
		if err1 != nil || err2 != nil || !os.SameFile(ei, li) {
			t.Fatalf("list path mismatch\n got: %q\nwant (same file as): %q\nstat errors: %v %v", listedPath, expectedPath, err1, err2)
		}
	} else {
		if listedPath != expectedPath {
			t.Fatalf("list path mismatch\n got: %q\nwant: %q", listedPath, expectedPath)
		}
	}

	// Select by index should print the path (no newline guaranteed)
	out, errOut, code = runMain(t, "", "select", "1")
	if code != 0 {
		t.Fatalf("select 1 exit=%d err=%s", code, errOut)
	}
	sel := strings.TrimSpace(out)
	if runtime.GOOS == "windows" {
		ei, err1 := os.Stat(expectedPath)
		li, err2 := os.Stat(sel)
		if err1 != nil || err2 != nil || !os.SameFile(ei, li) {
			t.Fatalf("select path mismatch\n got: %q\nwant (same file as): %q\nstat errors: %v %v", sel, expectedPath, err1, err2)
		}
	} else {
		if sel != expectedPath {
			t.Fatalf("select output mismatch\n got: %q\nwant: %q", sel, expectedPath)
		}
	}

	// Remove by index
	if out, errOut, code = runMain(t, "", "remove", "1"); code != 0 {
		t.Fatalf("remove 1 exit=%d err=%s out=%s", code, errOut, out)
	}

	// List should now be empty (no lines)
	out, errOut, code = runMain(t, "", "list")
	if code != 0 {
		t.Fatalf("list after remove exit=%d err=%s", code, errOut)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty list after remove, got: %q", out)
	}
}

func TestCLI_Path_ShowsDBAbsolutePath(t *testing.T) {
	cfg := setTestConfigEnv(t)

	// Trigger DB creation implicitly by any command that opens the DB
	// Using list is safe and shouldn't error
	if out, errOut, code := runMain(t, "", "list"); code != 0 {
		t.Fatalf("list exit=%d err=%s out=%s", code, errOut, out)
	}

	// Now query the path
	out, errOut, code := runMain(t, "", "path")
	if code != 0 {
		t.Fatalf("path exit=%d err=%s out=%s", code, errOut, out)
	}
	got := strings.TrimSpace(out)
	if got == "" {
		t.Fatalf("expected non-empty db path")
	}
	// The path should be inside the config dir
	// Windows: APPDATA\stashdir\config.json; Unix: XDG_CONFIG_HOME/stashdir/config.json; macOS: $HOME/Library/Application Support/stashdir/config.json
	// Instead of re-implementing os.UserConfigDir logic, just assert it has the config temp dir as a prefix after cleaning and case-normalizing on Windows.
	cleaned := filepath.Clean(got)
	cfgClean := filepath.Clean(cfg)
	if runtime.GOOS == "windows" {
		cleaned = strings.ToLower(cleaned)
		cfgClean = strings.ToLower(cfgClean)
	}
	if !strings.HasPrefix(cleaned, cfgClean) {
		t.Fatalf("db path not under test config dir\n got: %q\nwant prefix: %q", cleaned, cfgClean)
	}
	// Ensure file name is config.json
	if filepath.Base(cleaned) != "config.json" {
		t.Fatalf("unexpected db filename: %q", filepath.Base(cleaned))
	}
}
