package hookhelpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type scratchRepo struct {
	root     string // directory containing the repo
	repo     string // the main worktree
	worktree string // a linked worktree
}

func newScratchRepo(t *testing.T) scratchRepo {
	t.Helper()

	root := t.TempDir()
	repo := filepath.Join(root, "main")
	runGit(t, root, "init", "-q", "--initial-branch=main", "main")
	runGit(t, repo, "commit", "--allow-empty", "-q", "-m", "init")

	worktree := filepath.Join(root, "stem", "wt")
	runGit(t, repo, "worktree", "add", "-b", "feat", worktree)
	return scratchRepo{root: root, repo: repo, worktree: worktree}
}

func runGit(t *testing.T, dir string, args ...string) {
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestMainWorktreeFromWorktree(t *testing.T) {
	r := newScratchRepo(t)
	got, err := mainWorktree(r.worktree)
	if err != nil {
		t.Fatalf("mainWorktree: %v", err)
	}
	if got != r.repo {
		t.Errorf("mainWorktree = %q, want %q", got, r.repo)
	}
}

func TestMainWorktreeFromMainCheckout(t *testing.T) {
	r := newScratchRepo(t)
	got, err := mainWorktree(r.repo)
	if err != nil {
		t.Fatalf("mainWorktree: %v", err)
	}
	if got != r.repo {
		t.Errorf("mainWorktree = %q, want %q", got, r.repo)
	}
}

func TestMainWorktreeOutsideRepo(t *testing.T) {
	if _, err := mainWorktree(t.TempDir()); err == nil {
		t.Fatal("mainWorktree outside a repo should fail")
	}
}

func TestCopyFileToWorktree(t *testing.T) {
	r := newScratchRepo(t)
	writeFile(t, filepath.Join(r.repo, ".envrc"), "SECRET=1")

	var stderr bytes.Buffer
	if err := runCopyIn(t, r.worktree, []string{".envrc"}, &stderr); err != nil {
		t.Fatalf("copy: %v\nstderr:\n%s", err, stderr.String())
	}
	got, err := os.ReadFile(filepath.Join(r.worktree, ".envrc"))
	if err != nil {
		t.Fatalf("read copied .envrc: %v", err)
	}
	if string(got) != "SECRET=1" {
		t.Errorf(".envrc = %q, want %q", got, "SECRET=1")
	}
}

func TestCopyMissingSourceSkipsQuietly(t *testing.T) {
	r := newScratchRepo(t)
	writeFile(t, filepath.Join(r.repo, ".envrc"), "SECRET=1")

	var stderr bytes.Buffer
	if err := runCopyIn(t, r.worktree, []string{".envrc", ".nope"}, &stderr); err != nil {
		t.Fatalf("copy: %v\nstderr:\n%s", err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("missing source should skip quietly, stderr:\n%s", stderr.String())
	}
}

func TestCopyRejectsUnsafePaths(t *testing.T) {
	r := newScratchRepo(t)
	writeFile(t, filepath.Join(r.repo, ".envrc"), "SECRET=1")

	for _, p := range []string{"/etc/passwd", "../.envrc", "a/../../b"} {
		var stderr bytes.Buffer
		if err := runCopyIn(t, r.worktree, []string{p}, &stderr); err == nil {
			t.Errorf("copy %q should fail", p)
		}
	}
	if _, err := os.Stat(filepath.Join(r.worktree, ".envrc")); err == nil {
		t.Error("nothing should have been copied")
	}
}

func TestCopyDirectoryPreservesSymlink(t *testing.T) {
	r := newScratchRepo(t)
	pkg := filepath.Join(r.repo, "node_modules", "pkg")
	writeFile(t, filepath.Join(pkg, "index.js"), "module.exports = 1")
	if err := os.Symlink("./pkg", filepath.Join(r.repo, "node_modules", "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	var stderr bytes.Buffer
	if err := runCopyIn(t, r.worktree, []string{"node_modules"}, &stderr); err != nil {
		t.Fatalf("copy: %v\nstderr:\n%s", err, stderr.String())
	}

	if got, err := os.ReadFile(filepath.Join(r.worktree, "node_modules", "pkg", "index.js")); err != nil || string(got) != "module.exports = 1" {
		t.Errorf("node_modules/pkg/index.js = %q, err %v", got, err)
	}
	link := filepath.Join(r.worktree, "node_modules", "link")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink %s: %v", link, err)
	}
	if target != "./pkg" {
		t.Errorf("symlink target = %q, want ./pkg", target)
	}
}

func TestCopyOverwritesExistingEntries(t *testing.T) {
	r := newScratchRepo(t)
	writeFile(t, filepath.Join(r.repo, ".envrc"), "SECRET=1")
	writeFile(t, filepath.Join(r.repo, "node_modules", "pkg", "index.js"), "module.exports = 1")

	writeFile(t, filepath.Join(r.worktree, ".envrc"), "STALE=1")
	writeFile(t, filepath.Join(r.worktree, "node_modules", "stale.js"), "stale")

	var stderr bytes.Buffer
	if err := runCopyIn(t, r.worktree, []string{".envrc", "node_modules"}, &stderr); err != nil {
		t.Fatalf("copy: %v\nstderr:\n%s", err, stderr.String())
	}

	if got, _ := os.ReadFile(filepath.Join(r.worktree, ".envrc")); string(got) != "SECRET=1" {
		t.Errorf(".envrc = %q, want refreshed content", got)
	}
	if _, err := os.Stat(filepath.Join(r.worktree, "node_modules", "stale.js")); err == nil {
		t.Error("stale file should not survive a directory re-copy")
	}
}

func TestCopyFromMainWorktreeIsNoop(t *testing.T) {
	r := newScratchRepo(t)
	writeFile(t, filepath.Join(r.repo, ".envrc"), "SECRET=1")

	var stderr bytes.Buffer
	if err := runCopyIn(t, r.repo, []string{".envrc"}, &stderr); err != nil {
		t.Fatalf("copy from main worktree should exit 0: %v\nstderr:\n%s", err, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("main worktree")) {
		t.Errorf("expected a stderr note, got:\n%s", stderr.String())
	}
}

func TestCopyReportsFailureAndAttemptsRemaining(t *testing.T) {
	r := newScratchRepo(t)
	writeFile(t, filepath.Join(r.repo, ".envrc"), "SECRET=1")
	// A path through a regular file cannot exist; requesting it fails while
	// the sibling valid path must still be attempted.
	writeFile(t, filepath.Join(r.repo, "blocker"), "x")

	var stderr bytes.Buffer
	err := runCopyIn(t, r.worktree, []string{"blocker/child", ".envrc"}, &stderr)
	if err == nil {
		t.Fatalf("copy should report failure\nstderr:\n%s", stderr.String())
	}
	if got, err := os.ReadFile(filepath.Join(r.worktree, ".envrc")); err != nil || string(got) != "SECRET=1" {
		t.Errorf("remaining path should still be copied: %q, %v", got, err)
	}
}

func TestCopyDirFallsBackWhenCpFails(t *testing.T) {
	r := newScratchRepo(t)
	pkg := filepath.Join(r.repo, "node_modules", "pkg")
	writeFile(t, filepath.Join(pkg, "index.js"), "module.exports = 1")
	if err := os.Symlink("./pkg", filepath.Join(r.repo, "node_modules", "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	orig := cpDir
	cpDir = func(src, dst string) error { return fmt.Errorf("cp unavailable") }
	defer func() { cpDir = orig }()

	var stderr bytes.Buffer
	if err := runCopyIn(t, r.worktree, []string{"node_modules"}, &stderr); err != nil {
		t.Fatalf("copy: %v\nstderr:\n%s", err, stderr.String())
	}

	if got, err := os.ReadFile(filepath.Join(r.worktree, "node_modules", "pkg", "index.js")); err != nil || string(got) != "module.exports = 1" {
		t.Errorf("stdlib fallback should still copy the tree: %q, %v", got, err)
	}
	target, err := os.Readlink(filepath.Join(r.worktree, "node_modules", "link"))
	if err != nil || target != "./pkg" {
		t.Errorf("stdlib fallback should preserve symlinks: %q, %v", target, err)
	}
}

// runCopyIn runs Copy with cwd set to dir.
func runCopyIn(t *testing.T, dir string, paths []string, stderr *bytes.Buffer) error {
	t.Helper()
	t.Chdir(dir)
	return Copy(paths, stderr)
}
