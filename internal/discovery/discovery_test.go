package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func makeRepo(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if out, err := exec.Command("git", "init", "-q", path).CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", path, err, out)
	}
}

func TestScanFlatLayout(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, filepath.Join(root, "project_a"))
	makeRepo(t, filepath.Join(root, "project_b"))

	repos, err := Discover([]string{root})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	var names []string
	for _, r := range repos {
		names = append(names, r.Name)
	}
	if len(names) != 2 || names[0] != "project_a" || names[1] != "project_b" {
		t.Errorf("discovered %v, want [project_a project_b]", names)
	}
}

func TestHiddenAndVendorDirsExcluded(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, filepath.Join(root, "real"))
	makeRepo(t, filepath.Join(root, "node_modules"))
	makeRepo(t, filepath.Join(root, "vendor"))
	makeRepo(t, filepath.Join(root, ".hiddot"))
	if err := os.MkdirAll(filepath.Join(root, ".hidden"), 0o755); err != nil {
		t.Fatalf("mkdir hidden: %v", err)
	}
	makeRepo(t, filepath.Join(root, ".hidden", "secret"))
	makeRepo(t, filepath.Join(root, "target"))

	repos, err := Discover([]string{root})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "real" {
		t.Errorf("discovered %v, want only [real]", repos)
	}
}

func TestNestedReposIgnored(t *testing.T) {
	root := t.TempDir()
	outer := filepath.Join(root, "outer")
	if err := os.MkdirAll(outer, 0o755); err != nil {
		t.Fatalf("mkdir outer: %v", err)
	}
	makeRepo(t, filepath.Join(outer, "inner"))

	repos, err := Discover([]string{root})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("discovered %v, want none: nested repos are out of scope", repos)
	}
}

func TestNonRepoEntriesSkipped(t *testing.T) {
	root := t.TempDir()
	makeRepo(t, filepath.Join(root, "project_a"))
	if err := os.MkdirAll(filepath.Join(root, "plain_dir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "loose_file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	repos, err := Discover([]string{root})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "project_a" {
		t.Errorf("discovered %v, want [project_a]", repos)
	}
}

func TestDuplicateNamesCollapse(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	makeRepo(t, filepath.Join(first, "project_a"))
	makeRepo(t, filepath.Join(second, "project_a"))
	makeRepo(t, filepath.Join(second, "project_b"))

	repos, err := Discover([]string{first, second})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("discovered %v, want 2 repos with duplicate name collapsed", repos)
	}
}

func TestMissingSourcesDirErrors(t *testing.T) {
	_, err := Discover([]string{filepath.Join(t.TempDir(), "nope")})
	if err == nil {
		t.Fatal("expected error for missing sources directory")
	}
}
