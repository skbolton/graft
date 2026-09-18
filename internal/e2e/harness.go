// Package e2e drives the compiled graft binary against real git fixture
// repositories in an isolated temp HOME. No git mocking.
//
// The Go build caches are pinned at package init because tests replace HOME
// with an isolated temp directory; without pinning, a build triggered from
// inside a test would write its module cache into that directory and break
// TempDir cleanup.
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	hostHome       = os.Getenv("HOME")
	hostModCache   = os.Getenv("GOMODCACHE")
	hostGoCache    = os.Getenv("GOCACHE")
	binaryOnce     sync.Once
	binaryPath     string
	binaryBuildErr error
)

func init() {
	if hostModCache == "" {
		hostModCache = filepath.Join(hostHome, "go", "pkg", "mod")
	}
	if hostGoCache == "" {
		hostGoCache = filepath.Join(hostHome, ".cache", "go-build")
	}
}

// GraftBinary builds the graft binary and returns its path. The build runs
// once per test process; set GRAFT_E2E_BIN to point at a pre-built binary
// instead of building.
func GraftBinary(t *testing.T) string {
	t.Helper()

	if env := os.Getenv("GRAFT_E2E_BIN"); env != "" {
		if _, err := os.Stat(env); err != nil {
			t.Fatalf("GRAFT_E2E_BIN %s: %v", env, err)
		}
		return env
	}

	binaryOnce.Do(buildBinary)
	if binaryBuildErr != nil {
		t.Fatalf("build graft binary: %v", binaryBuildErr)
	}
	return binaryPath
}

func buildBinary() {
	dir, err := os.MkdirTemp("", "graft-e2e-bin-")
	if err != nil {
		binaryBuildErr = err
		return
	}

	binaryPath = filepath.Join(dir, "graft")
	cmd := exec.Command("go", "build", "-o", binaryPath, "github.com/skbolton/graft/cmd/graft")
	cmd.Env = append(os.Environ(),
		"GOMODCACHE="+hostModCache,
		"GOCACHE="+hostGoCache,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		binaryBuildErr = fmt.Errorf("%v\n%s", err, out)
	}
}

// RunGraft invokes the graft binary with the given arguments, inheriting the
// test process environment (so t.Setenv("HOME", ...) isolates it), and
// returns stdout, stderr, and the exit code.
func RunGraft(t *testing.T, bin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run graft %v: %v", args, err)
		}
		exitCode = exitErr.ExitCode()
	}
	return out.String(), errOut.String(), exitCode
}

// TempHOME creates an isolated HOME for the test process and any graft
// subprocesses it spawns. The directory is registered for cleanup.
func TempHOME(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

// FixtureRepo creates a git repository at sources/name under home with an
// initial commit and deterministic identity, isolated from host git
// configuration. It returns the repository path.
func FixtureRepo(t *testing.T, home, name string) string {
	t.Helper()

	path := filepath.Join(home, "sources", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}

	git := func(args ...string) {
		t.Helper()
		full := append([]string{
			"-c", "user.name=graft-test",
			"-c", "user.email=graft@test.invalid",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = path
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, path, err, out)
		}
	}

	git("init", "-q", "--initial-branch=main", path)
	git("commit", "--allow-empty", "-q", "-m", fmt.Sprintf("initial commit for %s", name))
	return path
}
