package stems

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/manifest"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "user.name=graft-test", "-c", "user.email=graft@test.invalid"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// fixtureStem builds a source repo with locally resolvable origin refs, a
// worktree for it inside a stem, and the stem's manifest.
func fixtureStem(t *testing.T, repoName, stemName string) (stemsDir, stemDir, repoDir, sourceDir string) {
	t.Helper()

	root := t.TempDir()
	sources := filepath.Join(root, "sources")
	stemsDir = filepath.Join(root, "stems")
	sourceDir = filepath.Join(sources, repoName)
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, sourceDir, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(sourceDir, "seed.txt"), []byte("seed"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, sourceDir, "add", ".")
	gitRun(t, sourceDir, "commit", "-q", "-m", "seed")
	gitRun(t, sourceDir, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, sourceDir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	stemDir = filepath.Join(stemsDir, stemName)
	repoDir = filepath.Join(stemDir, repoName)
	gitRun(t, sourceDir, "worktree", "add", "-b", stemName, repoDir)

	m := &manifest.Manifest{
		Branch: stemName,
		Repos:  []manifest.Repo{{Name: repoName, Source: sourceDir}},
	}
	if err := manifest.Write(stemDir, m); err != nil {
		t.Fatal(err)
	}
	return stemsDir, stemDir, repoDir, sourceDir
}

func TestFindByPathStemRoot(t *testing.T) {
	stemsDir, stemDir, _, _ := fixtureStem(t, "project_a", "feat-x")
	s, err := FindByPath(stemsDir, stemDir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "feat-x" {
		t.Errorf("resolved %q, want feat-x", s.Name)
	}
}

func TestFindByPathMemberWorktree(t *testing.T) {
	stemsDir, _, repoDir, _ := fixtureStem(t, "project_a", "feat-x")
	s, err := FindByPath(stemsDir, repoDir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "feat-x" {
		t.Errorf("resolved %q, want feat-x", s.Name)
	}
}

func TestFindByPathDeepInside(t *testing.T) {
	stemsDir, _, repoDir, _ := fixtureStem(t, "project_a", "feat-x")
	deep := filepath.Join(repoDir, "cmd", "graft")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	s, err := FindByPath(stemsDir, deep)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "feat-x" {
		t.Errorf("resolved %q, want feat-x", s.Name)
	}
}

func TestFindByPathSymlinkedEntry(t *testing.T) {
	stemsDir, stemDir, _, _ := fixtureStem(t, "project_a", "feat-x")
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(stemDir, link); err != nil {
		t.Fatal(err)
	}
	s, err := FindByPath(stemsDir, filepath.Join(link, "project_a"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "feat-x" {
		t.Errorf("resolved %q, want feat-x", s.Name)
	}
}

func TestFindByPathOutsideStems(t *testing.T) {
	stemsDir, _, _, _ := fixtureStem(t, "project_a", "feat-x")
	_, err := FindByPath(stemsDir, t.TempDir())
	if err == nil {
		t.Fatal("expected error for a path outside any stem")
	}
	if !strings.Contains(err.Error(), "feat-x") {
		t.Errorf("error should list available stems:\n%v", err)
	}
}

func TestFindByPathNonexistentPath(t *testing.T) {
	stemsDir, _, _, _ := fixtureStem(t, "project_a", "feat-x")
	_, err := FindByPath(stemsDir, filepath.Join(stemsDir, "nope"))
	if err == nil {
		t.Fatal("expected error for a nonexistent path")
	}
	if !strings.Contains(err.Error(), "feat-x") {
		t.Errorf("error should list available stems:\n%v", err)
	}
}

func TestFindByPathStemsDirItself(t *testing.T) {
	stemsDir, _, _, _ := fixtureStem(t, "project_a", "feat-x")
	_, err := FindByPath(stemsDir, stemsDir)
	if err == nil {
		t.Fatal("expected error for the stems directory itself")
	}
}

func TestFindByPathNoStems(t *testing.T) {
	root := t.TempDir()
	stemsDir := filepath.Join(root, "stems")
	if err := os.MkdirAll(stemsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := FindByPath(stemsDir, root)
	if err == nil {
		t.Fatal("expected error when no stems exist")
	}
	if !strings.Contains(err.Error(), "no stems under") {
		t.Errorf("unexpected error:\n%v", err)
	}
}
