package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/e2e"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "user.name=graft-test", "-c", "user.email=graft@test.invalid"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func TestStatusCleanStem(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "status", f.stems+"/feat-x")
	if code != 0 {
		t.Fatalf("status: exit %d", code)
	}
	if strings.Contains(stdout, "dirty") {
		t.Errorf("freshly grown stem should be clean:\n%s", stdout)
	}
	if got := strings.Count(stdout, "clean"); got != 2 {
		t.Errorf("expected 2 clean repos, got %d:\n%s", got, stdout)
	}
}

func TestStatusDirtyRepo(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	untracked := filepath.Join(f.stems, "feat-x", "project_a", "scratch.txt")
	if err := os.WriteFile(untracked, []byte("hi"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}
	modified := filepath.Join(f.stems, "feat-x", "project_b", "edit.txt")
	if err := os.WriteFile(modified, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	gitIn(t, filepath.Join(f.stems, "feat-x", "project_b"), "add", "edit.txt")

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "status", f.stems+"/feat-x")
	if code != 0 {
		t.Fatalf("status: exit %d", code)
	}
	lines := statusLines(t, stdout)
	if !strings.Contains(lines["project_a"], "dirty") {
		t.Errorf("untracked file should mark project_a dirty:\n%s", stdout)
	}
	if !strings.Contains(lines["project_b"], "dirty") {
		t.Errorf("uncommitted change should mark project_b dirty:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Untracked files:") || !strings.Contains(stdout, "scratch.txt") {
		t.Errorf("dirty repo should show its full git status output:\n%s", stdout)
	}
}

func TestStatusFromInsideStem(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	stem := filepath.Join(f.stems, "feat-x")
	worktree := filepath.Join(stem, "project_a")
	if err := os.WriteFile(filepath.Join(worktree, "scratch.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}

	for _, dir := range []string{stem, worktree} {
		stdout, stderr, code := e2e.RunGraftIn(t, dir, e2e.GraftBinary(t), "status")
		if code != 0 {
			t.Fatalf("status from %s: exit %d\nstderr:\n%s", dir, code, stderr)
		}
		if !strings.Contains(stdout, "feat-x") || !strings.Contains(stdout, "project_a  dirty") {
			t.Errorf("bare status from %s should report the containing stem:\n%s", dir, stdout)
		}
		if !strings.Contains(stdout, "scratch.txt") {
			t.Errorf("bare status from %s should show the dirty repo's git status:\n%s", dir, stdout)
		}
	}

	stdout, stderr, code := e2e.RunGraftIn(t, stem, e2e.GraftBinary(t), "status", "./")
	if code != 0 {
		t.Fatalf("status ./: exit %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "project_a  dirty") {
		t.Errorf("status ./ from stem root should report the stem:\n%s", stdout)
	}
}

func TestStatusAheadBehind(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	wt := filepath.Join(f.stems, "feat-x", "project_a")
	gitIn(t, wt, "commit", "--allow-empty", "-q", "-m", "work in the stem")

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "status", f.stems+"/feat-x")
	if code != 0 {
		t.Fatalf("status: exit %d", code)
	}
	if !strings.Contains(statusLines(t, stdout)["project_a"], "ahead 1") {
		t.Errorf("expected ahead 1 for project_a:\n%s", stdout)
	}

	gitIn(t, filepath.Join(f.sources, "project_a"), "commit", "--allow-empty", "-q", "-m", "upstream move")
	gitIn(t, filepath.Join(f.sources, "project_a"), "update-ref", "refs/remotes/origin/main", "HEAD")

	stdout, _, code = e2e.RunGraft(t, e2e.GraftBinary(t), "status", f.stems+"/feat-x")
	if code != 0 {
		t.Fatalf("status: exit %d", code)
	}
	line := statusLines(t, stdout)["project_a"]
	if !strings.Contains(line, "ahead 1") || !strings.Contains(line, "behind 1") {
		t.Errorf("expected ahead 1 behind 1 for project_a:\n%s", stdout)
	}
}

func TestStatusNoNetwork(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")
	wt := filepath.Join(f.stems, "feat-x", "project_a")
	gitIn(t, wt, "commit", "--allow-empty", "-q", "-m", "local work")

	stdout, stderr, code := e2e.RunGraftEnv(t,
		[]string{"GIT_ALLOW_PROTOCOL=none", "GIT_TERMINAL_PROMPT=0"},
		e2e.GraftBinary(t), "status", f.stems+"/feat-x")
	if code != 0 {
		t.Fatalf("status must not need network access; exit %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(statusLines(t, stdout)["project_a"], "ahead 1") {
		t.Errorf("status should report against locally known refs:\n%s", stdout)
	}
}

func TestStatusOutsideAnyStem(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	home := f.home
	for _, tc := range []struct {
		name string
		dir  string
		arg  string
	}{
		{"bare outside", home, ""},
		{"path outside", home, home},
		{"nonexistent path", home, filepath.Join(home, "nope")},
	} {
		var stdout, stderr string
		var code int
		if tc.arg == "" {
			stdout, stderr, code = e2e.RunGraftIn(t, tc.dir, e2e.GraftBinary(t), "status")
		} else {
			stdout, stderr, code = e2e.RunGraft(t, e2e.GraftBinary(t), "status", tc.arg)
		}
		if code == 0 {
			t.Fatalf("%s: status should exit non-zero:\n%s", tc.name, stdout)
		}
		if !strings.Contains(stderr, "feat-x") {
			t.Errorf("%s: error should name the available stems:\n%s", tc.name, stderr)
		}
	}
}

func TestStatusPorcelainFraming(t *testing.T) {
	f := setupGrow(t)
	growStem(t, f, "feat-x")

	untracked := filepath.Join(f.stems, "feat-x", "project_a", "scratch.txt")
	if err := os.WriteFile(untracked, []byte("hi"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}

	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "status", "--porcelain", f.stems+"/feat-x")
	if code != 0 {
		t.Fatalf("status --porcelain: exit %d", code)
	}

	blocks := porcelainBlocks(t, stdout)
	if len(blocks) != 2 {
		t.Fatalf("expected one block per repo, got %d:\n%s", len(blocks), stdout)
	}
	wantPath := filepath.Join(f.stems, "feat-x", "project_a")
	if blocks[wantPath] == "" {
		t.Errorf("dirty repo block should follow its header:\n%s", stdout)
	}
	if !strings.Contains(blocks[wantPath], "?? scratch.txt") {
		t.Errorf("block should contain git status --porcelain lines:\n%s", stdout)
	}
	cleanPath := filepath.Join(f.stems, "feat-x", "project_b")
	if blocks[cleanPath] != "" {
		t.Errorf("clean repo should have an empty block:\n%q", blocks[cleanPath])
	}
}

func porcelainBlocks(t *testing.T, out string) map[string]string {
	t.Helper()
	blocks := map[string]string{}
	current := ""
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if rest, ok := strings.CutPrefix(line, "repo\t"); ok {
			fields := strings.Split(rest, "\t")
			if len(fields) != 3 {
				t.Fatalf("malformed repo header record: %q", line)
			}
			current = fields[2]
			blocks[current] = ""
			continue
		}
		if current == "" {
			t.Fatalf("git status line outside any repo block: %q", line)
		}
		if blocks[current] != "" {
			blocks[current] += "\n"
		}
		blocks[current] += line
	}
	return blocks
}

func statusLines(t *testing.T, out string) map[string]string {
	t.Helper()
	lines := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if !strings.HasPrefix(line, "  ") {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}
		lines[fields[0]] = line
	}
	return lines
}
