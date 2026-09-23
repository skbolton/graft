package stems

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/manifest"
)

func TestStatusDirtyAndClean(t *testing.T) {
	stemsDir, stemDir, repoDir, _ := fixtureStem(t, "project_a", "feat-x")

	clean, err := Find(stemsDir, "feat-x")
	if err != nil {
		t.Fatal(err)
	}
	report := Status(clean)
	if report.Repos[0].Dirty {
		t.Errorf("fresh worktree should be clean")
	}
	if len(report.Repos[0].Changes) != 0 {
		t.Errorf("clean repo should have no changes: %v", report.Repos[0].Changes)
	}

	if err := os.WriteFile(filepath.Join(repoDir, "scratch.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "seed.txt"), []byte("edited"), 0o644); err != nil {
		t.Fatal(err)
	}

	report = Status(clean)
	rs := report.Repos[0]
	if !rs.Dirty {
		t.Errorf("untracked + modified files should mark the repo dirty")
	}
	joined := strings.Join(rs.Changes, "\n")
	if !strings.Contains(joined, "scratch.txt") || !strings.Contains(joined, "seed.txt") {
		t.Errorf("changes should include both files:\n%s", joined)
	}
	_ = stemDir
}

func TestStatusProblemRepo(t *testing.T) {
	stemsDir, stemDir, _, sourceDir := fixtureStem(t, "project_a", "feat-x")

	// A manifest entry with a real source but no worktree registered.
	m, err := manifest.Read(stemDir)
	if err != nil {
		t.Fatal(err)
	}
	m.Repos = append(m.Repos, manifest.Repo{Name: "ghost", Source: sourceDir})
	if err := manifest.Write(stemDir, m); err != nil {
		t.Fatal(err)
	}
	s, err := Find(stemsDir, "feat-x")
	if err != nil {
		t.Fatal(err)
	}

	report := Status(s)
	var ghost *RepoStatus
	for i := range report.Repos {
		if report.Repos[i].Name == "ghost" {
			ghost = &report.Repos[i]
		}
	}
	if ghost == nil {
		t.Fatal("ghost repo missing from report")
	}
	if ghost.Problem == "" {
		t.Errorf("missing worktree should report a problem")
	}
	if ghost.Dirty {
		t.Errorf("problem repo should not be marked dirty")
	}
	_ = stemsDir
}

func TestWriteStatusShowsGitStatusForDirtyRepos(t *testing.T) {
	stemsDir, _, repoDir, _ := fixtureStem(t, "project_a", "feat-x")

	if err := os.WriteFile(filepath.Join(repoDir, "scratch.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Find(stemsDir, "feat-x")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	WriteStatus(&buf, Status(s))
	out := buf.String()

	if !strings.Contains(out, "project_a  dirty") {
		t.Errorf("summary should mark project_a dirty:\n%s", out)
	}
	if !strings.Contains(out, "Untracked files:") || !strings.Contains(out, "scratch.txt") {
		t.Errorf("dirty repo should show its full git status output:\n%s", out)
	}
}

func TestWriteStatusCleanReposStayOneLine(t *testing.T) {
	stemsDir, _, _, _ := fixtureStem(t, "project_a", "feat-x")

	s, err := Find(stemsDir, "feat-x")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	WriteStatus(&buf, Status(s))
	out := buf.String()

	if !strings.Contains(out, "project_a  clean") {
		t.Errorf("summary should mark project_a clean:\n%s", out)
	}
	if strings.Contains(out, "nothing to commit") {
		t.Errorf("clean repos should not show git status output:\n%s", out)
	}
}

func TestWriteStatusPorcelainFraming(t *testing.T) {
	stemsDir, stemDir, repoDir, sourceDir := fixtureStem(t, "project_a", "feat-x")

	if err := os.WriteFile(filepath.Join(repoDir, "scratch.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := manifest.Read(stemDir)
	if err != nil {
		t.Fatal(err)
	}
	m.Repos = append(m.Repos, manifest.Repo{Name: "project_b", Source: sourceDir})
	if err := manifest.Write(stemDir, m); err != nil {
		t.Fatal(err)
	}
	// project_b has no worktree, so it is a problem repo: header only.
	s, err := Find(stemsDir, "feat-x")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	WriteStatusPorcelain(&buf, Status(s))
	out := buf.String()

	wantHeader := "repo\tfeat-x\tproject_a\t" + repoDir + "\n"
	if !strings.Contains(out, wantHeader) {
		t.Errorf("expected repo header record:\n%s", out)
	}
	header, body, _ := strings.Cut(out, wantHeader)
	if header != "" {
		t.Errorf("first record should be the header for project_a:\n%s", out)
	}
	if !strings.Contains(body, "?? scratch.txt") {
		t.Errorf("porcelain block should contain git status lines verbatim:\n%s", out)
	}
	if !strings.Contains(body, "repo\tfeat-x\tproject_b\t") {
		t.Errorf("every member repo should get a header record:\n%s", out)
	}
}
