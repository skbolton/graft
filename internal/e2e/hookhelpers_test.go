package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/e2e"
)

const zerosSha = "0000000000000000000000000000000000000000"

// TestHookOnCreateInPostCheckout drives the compiled binary through a real
// repo's post-checkout hook: the helper must detect worktree creation and
// reject ordinary branch switches.
func TestHookOnCreateInPostCheckout(t *testing.T) {
	home := e2e.TempHOME(t)
	bin := e2e.GraftBinary(t)
	repo := e2e.FixtureRepo(t, home, "project_hooky")

	logPath := filepath.Join(home, "hook.log")
	hook := fmt.Sprintf(`#!/bin/sh
echo "$1" >> %s
exec %s hook on-create "$1"
`, logPath, bin)
	e2e.FixtureRepoWithHook(t, home, "project_hooky", hook)

	worktree := filepath.Join(home, "stem", "project_hooky")
	if out, err := exec.Command("git", "-C", repo, "worktree", "add", "-b", "feat", worktree).CombinedOutput(); err != nil {
		t.Fatalf("worktree add should succeed when the hook detects creation: %v\n%s", err, out)
	}
	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("worktree missing: %v", err)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read hook log: %v", err)
	}
	if got := strings.TrimSpace(string(log)); got != zerosSha {
		t.Errorf("creation should log the zeros sha, got %q", got)
	}

	if err := os.Remove(logPath); err != nil {
		t.Fatalf("clear hook log: %v", err)
	}

	if out, err := exec.Command("git", "-C", repo, "checkout", "-q", "-b", "switch-away").CombinedOutput(); err == nil {
		t.Fatalf("branch switch should fail while the hook rejects it:\n%s", out)
	}
	log, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read hook log: %v", err)
	}
	if got := strings.TrimSpace(string(log)); got == zerosSha || got == "" {
		t.Errorf("branch switch should log a real sha, got %q", got)
	}
}

// TestGrowHookCopyHelpers grows a stem whose repo bootstraps itself through
// the helper idiom: untracked .envrc and an ignored node_modules directory
// (with a symlink inside) must appear in the new worktree.
func TestGrowHookCopyHelpers(t *testing.T) {
	home := e2e.TempHOME(t)
	bin := e2e.GraftBinary(t)
	repo := e2e.FixtureRepo(t, home, "project_hooky")

	hook := fmt.Sprintf(`#!/bin/sh
if %s hook on-create "$1"; then
  %s hook copy .envrc node_modules
fi
exit 0
`, bin, bin)
	e2e.FixtureRepoWithHook(t, home, "project_hooky", hook)

	writeFile(t, filepath.Join(repo, ".envrc"), "SECRET=1")
	if err := os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte("node_modules\n"), 0o644); err != nil {
		t.Fatalf("write exclude: %v", err)
	}
	writeFile(t, filepath.Join(repo, "node_modules", "pkg", "index.js"), "module.exports = 1")
	if err := os.Symlink("./pkg", filepath.Join(repo, "node_modules", "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	sources := filepath.Join(home, "sources")
	stems := filepath.Join(home, "stems")
	e2e.WriteConfig(t, home, fmt.Sprintf("sources = [%q]\nstems = %q\n", sources, stems))

	stdout, stderr, code := e2e.RunGraft(t, bin, "grow", "feat-c", "-s", "project_hooky")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	worktree := filepath.Join(stems, "feat-c", "project_hooky")
	if got, err := os.ReadFile(filepath.Join(worktree, ".envrc")); err != nil || string(got) != "SECRET=1" {
		t.Errorf(".envrc in worktree = %q, err %v", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(worktree, "node_modules", "pkg", "index.js")); err != nil || string(got) != "module.exports = 1" {
		t.Errorf("node_modules/pkg/index.js = %q, err %v", got, err)
	}
	target, err := os.Readlink(filepath.Join(worktree, "node_modules", "link"))
	if err != nil {
		t.Fatalf("readlink in worktree: %v", err)
	}
	if target != "./pkg" {
		t.Errorf("symlink target = %q, want ./pkg", target)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
