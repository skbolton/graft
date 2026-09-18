package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/e2e"
)

func branchExists(t *testing.T, repo, branch string) bool {
	t.Helper()
	err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run()
	return err == nil
}

func TestDeleteCleanStem(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "feat-x")
	if code != 0 {
		t.Fatalf("delete: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-x")); !os.IsNotExist(err) {
		t.Errorf("stem directory should be gone, stat err: %v", err)
	}
	for _, repo := range []string{"project_a", "project_b"} {
		source := filepath.Join(f.sources, repo)
		if branchExists(t, source, "feat-x") {
			t.Errorf("branch feat-x should be deleted from %s", repo)
		}
		out, err := exec.Command("git", "-C", source, "worktree", "list", "--porcelain").CombinedOutput()
		if err != nil {
			t.Fatalf("worktree list in %s: %v\n%s", repo, err, out)
		}
		if strings.Contains(string(out), filepath.Join(f.stems, "feat-x", repo)) {
			t.Errorf("worktree %s should be unregistered:\n%s", repo, out)
		}
	}
}

func TestDeleteDirtyRefused(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	if err := os.WriteFile(filepath.Join(f.stems, "feat-x", "project_a", "wip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("make dirty: %v", err)
	}

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "feat-x")
	if code == 0 {
		t.Fatalf("dirty stem should be refused\nstdout:\n%s", stdout)
	}
	if !strings.Contains(stdout, "project_a") || !strings.Contains(stdout, "dirty") {
		t.Errorf("summary should show where work exists:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-x", "project_a")); err != nil {
		t.Errorf("refusal must remove nothing: %v", err)
	}
	if !branchExists(t, filepath.Join(f.sources, "project_a"), "feat-x") {
		t.Error("refusal must not delete branches")
	}
}

func TestDeleteUnpushedRefused(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	wt := filepath.Join(f.stems, "feat-x", "project_a")
	gitIn(t, wt, "commit", "--allow-empty", "-q", "-m", "unpushed")

	_, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "feat-x")
	if code == 0 {
		t.Fatal("unpushed commits should block deletion")
	}
}

func TestDeleteUnpushedButOnOriginProceeds(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	wt := filepath.Join(f.stems, "feat-x", "project_a")
	gitIn(t, wt, "commit", "--allow-empty", "-q", "-m", "pushed work")
	sha, err := exec.Command("git", "-C", wt, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	gitIn(t, filepath.Join(f.sources, "project_a"),
		"update-ref", "refs/remotes/origin/feat-x", strings.TrimSpace(string(sha)))

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "feat-x")
	if code != 0 {
		t.Fatalf("delete should proceed when work is on origin: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-x")); !os.IsNotExist(err) {
		t.Errorf("stem directory should be gone, stat err: %v", err)
	}
}

func TestDeleteForcedProceeds(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	if err := os.WriteFile(filepath.Join(f.stems, "feat-x", "project_a", "wip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("make dirty: %v", err)
	}
	gitIn(t, filepath.Join(f.stems, "feat-x", "project_b"), "commit", "--allow-empty", "-q", "-m", "unpushed")

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "--force", "feat-x")
	if code != 0 {
		t.Fatalf("forced delete: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "project_a") || !strings.Contains(stdout, "project_b") {
		t.Errorf("summary should report what was discarded even with --force:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-x")); !os.IsNotExist(err) {
		t.Errorf("stem directory should be gone, stat err: %v", err)
	}
}

func TestDeleteUnknownStem(t *testing.T) {
	setupGrow(t)
	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "ghost")
	if code == 0 {
		t.Fatal("delete of unknown stem should exit non-zero")
	}
	if !strings.Contains(stderr, "ghost") {
		t.Errorf("error should name the stem:\n%s", stderr)
	}
}

func TestDeleteFullLoop(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-loop")

	wt := filepath.Join(f.stems, "feat-loop", "project_a")
	if err := os.WriteFile(filepath.Join(wt, "wip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("make dirty: %v", err)
	}

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "status", "feat-loop")
	if code != 0 {
		t.Fatalf("status: exit %d", code)
	}
	if !strings.Contains(stdout, "dirty") {
		t.Errorf("status should show the dirty repo:\n%s", stdout)
	}

	if _, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "feat-loop"); code == 0 {
		t.Fatal("dirty stem must be refused")
	}
	if _, err := os.Stat(wt); err != nil {
		t.Fatalf("refusal must remove nothing: %v", err)
	}

	if err := os.Remove(filepath.Join(wt, "wip.txt")); err != nil {
		t.Fatalf("clean up: %v", err)
	}

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "delete", "feat-loop")
	if code != 0 {
		t.Fatalf("clean delete: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-loop")); !os.IsNotExist(err) {
		t.Errorf("stem directory should be gone, stat err: %v", err)
	}
}
