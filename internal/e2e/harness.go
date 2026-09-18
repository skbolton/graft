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
	return runGraftDir(t, "", bin, args...)
}

// RunGraftIn is RunGraft with the graft process running in dir.
func RunGraftIn(t *testing.T, dir, bin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	return runGraftDir(t, dir, bin, args...)
}

func runGraftDir(t *testing.T, dir, bin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
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
// configuration. origin/HEAD is locally resolvable, as in a normal clone.
func FixtureRepo(t *testing.T, home, name string) string {
	t.Helper()

	path := fixtureRepoDir(t, home, name)
	git := gitRunner(t, path)
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	git("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	return path
}

// FixtureRepoWithoutOrigin is FixtureRepo but leaves origin/HEAD unset.
func FixtureRepoWithoutOrigin(t *testing.T, home, name string) string {
	return fixtureRepoDir(t, home, name)
}

// FixtureRepoWithHook creates a fixture repo whose post-checkout hook runs
// script on worktree creation.
func FixtureRepoWithHook(t *testing.T, home, name, script string) string {
	t.Helper()

	path := FixtureRepo(t, home, name)
	hookPath := filepath.Join(path, ".git", "hooks", "post-checkout")
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write post-checkout hook in %s: %v", path, err)
	}
	return path
}

// WriteConfig writes graft.toml into home's config directory.
func WriteConfig(t *testing.T, home, content string) {
	t.Helper()

	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "graft.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func fixtureRepoDir(t *testing.T, home, name string) string {
	t.Helper()

	path := filepath.Join(home, "sources", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}

	git := gitRunner(t, path)
	git("init", "-q", "--initial-branch=main", path)
	git("commit", "--allow-empty", "-q", "-m", fmt.Sprintf("initial commit for %s", name))
	return path
}

func gitRunner(t *testing.T, dir string) func(args ...string) {
	t.Helper()
	return func(args ...string) {
		t.Helper()
		full := append([]string{
			"-c", "user.name=graft-test",
			"-c", "user.email=graft@test.invalid",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
}
