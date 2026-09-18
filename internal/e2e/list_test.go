package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/e2e"
)

func growStem(t *testing.T, f growFixture, name string, extra ...string) {
	t.Helper()
	args := append([]string{"grow", name, "-c", "backend"}, extra...)
	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), args...)
	if code != 0 {
		t.Fatalf("grow %s: exit %d\nstdout:\n%s\nstderr:\n%s", name, code, stdout, stderr)
	}
}

func TestListHumanShowsStemsAndRepos(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	growStem(t, f, "feat-y")

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "list")
	if code != 0 {
		t.Fatalf("list: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	for _, want := range []string{
		"feat-x", "branch: feat-x", "collection: backend",
		"feat-y", "branch: feat-y",
		"project_a", "project_b",
		filepath.Join(f.stems, "feat-x", "project_a"),
		filepath.Join(f.stems, "feat-y", "project_a"),
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("list output missing %q:\n%s", want, stdout)
		}
	}
}

func TestListMembershipVerifiedAgainstGit(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	if err := os.RemoveAll(filepath.Join(f.stems, "feat-x", "project_a")); err != nil {
		t.Fatalf("remove worktree dir: %v", err)
	}
	if out, err := exec.Command("git", "-C", filepath.Join(f.sources, "project_a"), "worktree", "prune").CombinedOutput(); err != nil {
		t.Fatalf("prune: %v\n%s", err, out)
	}

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "list")
	if code != 0 {
		t.Fatalf("list: exit %d", code)
	}
	if !strings.Contains(stdout, "missing worktree") {
		t.Errorf("list should flag the removed worktree:\n%s", stdout)
	}
}

func TestListMalformedStem(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	if err := os.RemoveAll(filepath.Join(f.stems, "feat-x", ".graft")); err != nil {
		t.Fatalf("remove .graft: %v", err)
	}

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "list")
	if code != 0 {
		t.Fatalf("list must not crash on a malformed stem; exit %d", code)
	}
	if !strings.Contains(stdout, "malformed") || !strings.Contains(stdout, "feat-x") {
		t.Errorf("list should report the malformed stem:\n%s", stdout)
	}
}

func TestListPorcelainRecords(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "list", "--porcelain")
	if code != 0 {
		t.Fatalf("list --porcelain: exit %d", code)
	}

	stemPath := filepath.Join(f.stems, "feat-x")
	want := map[string]bool{
		"stem\tfeat-x\t" + stemPath:                                   false,
		"repo\tfeat-x\tproject_a\t" + filepath.Join(stemPath, "project_a"): false,
		"repo\tfeat-x\tproject_b\t" + filepath.Join(stemPath, "project_b"): false,
		"collection\tbackend\tproject_a,project_b":                    false,
		"source\tproject_a\t" + filepath.Join(f.sources, "project_a"): false,
		"source\tproject_b\t" + filepath.Join(f.sources, "project_b"): false,
	}

	allowed := map[string]bool{"stem": true, "repo": true, "collection": true, "source": true}
	for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
		kind, _, _ := strings.Cut(line, "\t")
		if !allowed[kind] {
			t.Errorf("record %q has unexpected type %q", line, kind)
			continue
		}
		if seen, ok := want[line]; ok {
			if seen {
				t.Errorf("duplicate record %q", line)
			}
			want[line] = true
		}
	}
	for line, seen := range want {
		if !seen {
			t.Errorf("missing record %q in:\n%s", line, stdout)
		}
	}
}

func TestListPorcelainIncludesMalformedStem(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	if err := os.RemoveAll(filepath.Join(f.stems, "feat-x", ".graft")); err != nil {
		t.Fatalf("remove .graft: %v", err)
	}
	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "list", "--porcelain")
	if code != 0 {
		t.Fatalf("list --porcelain: exit %d", code)
	}
	want := "stem\tfeat-x\t" + filepath.Join(f.stems, "feat-x")
	if !strings.Contains(stdout, want) {
		t.Errorf("porcelain should still emit the stem record %q:\n%s", want, stdout)
	}
	if strings.Contains(stdout, "repo\tfeat-x\t") {
		t.Errorf("malformed stem should carry no repo records:\n%s", stdout)
	}
}
