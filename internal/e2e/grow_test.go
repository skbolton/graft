package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/e2e"
	"github.com/skbolton/graft/internal/manifest"
)

type growFixture struct {
	home    string
	sources string
	stems   string
}

func setupGrow(t *testing.T) growFixture {
	t.Helper()

	home := e2e.TempHOME(t)
	f := growFixture{
		home:    home,
		sources: filepath.Join(home, "sources"),
		stems:   filepath.Join(home, "stems"),
	}
	e2e.FixtureRepo(t, home, "project_a")
	e2e.FixtureRepo(t, home, "project_b")
	e2e.WriteConfig(t, home, fmt.Sprintf(`
sources = ["%s"]
stems = "%s"

[collections]
backend = ["project_a", "project_b"]
`, f.sources, f.stems))
	return f
}

func currentBranch(t *testing.T, repo string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatalf("current branch of %s: %v", repo, err)
	}
	return strings.TrimSpace(string(out))
}

func TestGrowTwoRepos(t *testing.T) {
	f := setupGrow(t)
	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-x", "-s", "project_a,project_b")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	stem := filepath.Join(f.stems, "feat-x")
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if last := lines[len(lines)-1]; last != stem {
		t.Errorf("final stdout line = %q, want %q", last, stem)
	}

	for _, repo := range []string{"project_a", "project_b"} {
		worktree := filepath.Join(stem, repo)
		if _, err := os.Stat(worktree); err != nil {
			t.Errorf("worktree %s missing: %v", worktree, err)
			continue
		}
		if got := currentBranch(t, worktree); got != "feat-x" {
			t.Errorf("%s on branch %q, want feat-x", repo, got)
		}
	}

	m, err := manifest.Read(stem)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if m.Branch != "feat-x" || m.Collection != "" || len(m.Repos) != 2 {
		t.Errorf("manifest mismatch: %+v", m)
	}
}

func TestGrowCollectionSelection(t *testing.T) {
	f := setupGrow(t)
	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-c", "-c", "backend")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstderr:\n%s", code, stderr)
	}
	for _, repo := range []string{"project_a", "project_b"} {
		if _, err := os.Stat(filepath.Join(f.stems, "feat-c", repo)); err != nil {
			t.Errorf("worktree %s missing: %v", repo, err)
		}
	}
}

func TestGrowCwdAwareSingleRepo(t *testing.T) {
	f := setupGrow(t)
	repo := filepath.Join(f.sources, "project_a")
	_, stderr, code := e2e.RunGraftIn(t, repo, e2e.GraftBinary(t), "grow", "feat-y")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstderr:\n%s", code, stderr)
	}
	worktree := filepath.Join(f.stems, "feat-y", "project_a")
	if _, err := os.Stat(worktree); err != nil {
		t.Errorf("single-repo worktree missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-y", "project_b")); err == nil {
		t.Error("unselected repo should not be in the stem")
	}
}

func TestGrowOverlappingSelectionDedupes(t *testing.T) {
	f := setupGrow(t)
	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-o", "-c", "backend", "-s", "project_a")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstderr:\n%s", code, stderr)
	}
	m, err := manifest.Read(filepath.Join(f.stems, "feat-o"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if len(m.Repos) != 2 {
		t.Errorf("overlap should dedupe to 2 repos, got %+v", m.Repos)
	}
}

func TestGrowNoSelectionHint(t *testing.T) {
	setupGrow(t)
	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-h")
	if code == 0 {
		t.Fatalf("grow with no selection should exit non-zero\nstdout:\n%s", stdout)
	}
	for _, want := range []string{"-c", "-s", "inside a source repo"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("hint missing %q:\n%s", want, stderr)
		}
	}
}

func TestGrowSlashedBranchRejected(t *testing.T) {
	f := setupGrow(t)
	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat/login", "-s", "project_a")
	if code == 0 {
		t.Fatalf("slashed branch should exit non-zero\nstderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "directory name") {
		t.Errorf("error should explain directory-name requirement:\n%s", stderr)
	}
	if _, err := os.Stat(f.stems); err == nil {
		entries, _ := os.ReadDir(f.stems)
		if len(entries) > 0 {
			t.Errorf("no stem should be created, found %v", entries)
		}
	}
}

func TestGrowExistingStemRefused(t *testing.T) {
	setupGrow(t)
	if _, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-x", "-s", "project_a"); code != 0 {
		t.Fatalf("first grow: exit %d\n%s", code, stderr)
	}
	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-x", "-s", "project_b")
	if code == 0 {
		t.Fatalf("second grow should exit non-zero\nstderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "already exists") {
		t.Errorf("error should identify existing stem:\n%s", stderr)
	}
}

func TestGrowUnresolvableOriginHEAD(t *testing.T) {
	f := setupGrow(t)
	e2e.FixtureRepoWithoutOrigin(t, f.home, "project_broken")
	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-o", "-s", "project_broken,project_a")
	if code == 0 {
		t.Fatalf("grow with unresolvable origin/HEAD should exit non-zero\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "git remote set-head origin -a") {
		t.Errorf("failure should be actionable:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-o", "project_a")); err != nil {
		t.Errorf("healthy repo should still grow: %v", err)
	}
}

func TestGrowHookFailureIsWarning(t *testing.T) {
	f := setupGrow(t)
	e2e.FixtureRepoWithHook(t, f.home, "project_hooky", "#!/bin/sh\nexit 1\n")
	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-hf", "-s", "project_hooky,project_a")
	if code != 0 {
		t.Fatalf("hook failure is a warning, not an error; exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "warning") {
		t.Errorf("summary should report the hook failure as a warning:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-hf", "project_hooky")); err != nil {
		t.Errorf("worktree should exist despite hook failure: %v", err)
	}
}

func TestGrowSameBranchCollision(t *testing.T) {
	f := setupGrow(t)
	repo := filepath.Join(f.sources, "project_b")
	if out, err := exec.Command("git", "-C", repo, "checkout", "-q", "-b", "feat-b").CombinedOutput(); err != nil {
		t.Fatalf("create colliding branch: %v\n%s", err, out)
	}
	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-b", "-s", "project_a,project_b")
	if code == 0 {
		t.Fatalf("collision should exit non-zero\nstderr:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-b", "project_a")); err != nil {
		t.Errorf("other repo should proceed: %v", err)
	}
}

func TestGrowAllFailLeavesStemAndManifest(t *testing.T) {
	home := e2e.TempHOME(t)
	sources := filepath.Join(home, "sources")
	stems := filepath.Join(home, "stems")
	e2e.FixtureRepoWithoutOrigin(t, home, "project_a")
	e2e.FixtureRepoWithoutOrigin(t, home, "project_b")
	e2e.WriteConfig(t, home, fmt.Sprintf(`
sources = ["%s"]
stems = "%s"
`, sources, stems))

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-f", "-s", "project_a,project_b")
	if code == 0 {
		t.Fatalf("all-fail grow should exit non-zero\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stem := filepath.Join(stems, "feat-f")
	m, err := manifest.Read(stem)
	if err != nil {
		t.Fatalf("manifest should exist on total failure: %v", err)
	}
	if m.Branch != "feat-f" || len(m.Repos) != 0 {
		t.Errorf("manifest mismatch: %+v", m)
	}
}

func TestGrowPostcreateHookContext(t *testing.T) {
	f := setupGrow(t)
	e2e.WriteConfig(t, f.home, fmt.Sprintf(`
sources = ["%s"]
stems = "%s"
postcreate = "printf '%%s\\n' \"$GRAFT_WORKSPACE\" > ws.txt && printf '%%s\\n' \"$GRAFT_PATH\" > path.txt && printf '%%s\\n' \"$GRAFT_BRANCH\" > branch.txt && printf '%%s\\n' \"$GRAFT_COLLECTION\" > collection.txt && printf '%%s\\n' \"$GRAFT_REPOS\" > repos.txt"

[collections]
backend = ["project_a", "project_b"]
`, f.sources, f.stems))

	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-h", "-c", "backend")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	stem := lines[len(lines)-1]
	assertFile := func(name, want string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(stem, name))
		if err != nil {
			t.Errorf("hook artifact %s: %v", name, err)
			return
		}
		if got := strings.TrimSpace(string(data)); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	assertFile("ws.txt", "feat-h")
	assertFile("path.txt", stem)
	assertFile("branch.txt", "feat-h")
	assertFile("collection.txt", "backend")
	assertFile("repos.txt", "project_a\nproject_b")
}

func TestGrowPostcreateFailureIsWarning(t *testing.T) {
	f := setupGrow(t)
	e2e.WriteConfig(t, f.home, fmt.Sprintf(`
sources = ["%s"]
stems = "%s"
postcreate = "exit 1"

[collections]
backend = ["project_a"]
`, f.sources, f.stems))

	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-pf", "-s", "project_a")
	if code != 0 {
		t.Fatalf("postcreate failure is a warning; exit %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "postcreate hook failed") {
		t.Errorf("warning should mention the hook failure:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-pf", "project_a")); err != nil {
		t.Errorf("stem should remain intact: %v", err)
	}
}

func TestGrowNoHooks(t *testing.T) {
	f := setupGrow(t)
	e2e.WriteConfig(t, f.home, fmt.Sprintf(`
sources = ["%s"]
stems = "%s"
postcreate = "touch hook-ran"

[collections]
backend = ["project_a"]
`, f.sources, f.stems))

	_, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-nh", "-s", "project_a", "--no-hooks")
	if code != 0 {
		t.Fatalf("grow: exit %d\nstderr:\n%s", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-nh", "hook-ran")); err == nil {
		t.Error("--no-hooks should skip the postcreate hook")
	}
}

func TestGrowUnknownRepoReported(t *testing.T) {
	f := setupGrow(t)
	stdout, stderr, code := e2e.RunGraft(t, e2e.GraftBinary(t), "grow", "feat-u", "-s", "project_a,ghost")
	if code == 0 {
		t.Fatalf("unknown repo should exit non-zero\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "ghost") {
		t.Errorf("summary should name the unknown repo:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(f.stems, "feat-u", "project_a")); err != nil {
		t.Errorf("remaining repos should proceed: %v", err)
	}
}
