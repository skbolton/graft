package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRoundTrip(t *testing.T) {
	stem := t.TempDir()
	in := &Manifest{
		Branch:     "feat-x",
		Collection: "backend",
		CreatedAt:  "2026-09-18T12:00:00Z",
		Repos: []Repo{
			{Name: "project_a", Source: "/home/user/repos/project_a"},
			{Name: "project_b", Source: "/home/user/repos/project_b"},
		},
	}
	if err := Write(stem, in); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(Path(stem))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	text := string(data)
	for _, want := range []string{
		`branch = "feat-x"`,
		`collection = "backend"`,
		`created_at = "2026-09-18T12:00:00Z"`,
		"[[repos]]",
		`name = "project_a"`,
		`source = "/home/user/repos/project_a"`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("manifest missing %s:\n%s", want, text)
		}
	}

	out, err := Read(stem)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if out.Branch != in.Branch || out.Collection != in.Collection {
		t.Errorf("round trip mismatch: %+v", out)
	}
	if len(out.Repos) != 2 || out.Repos[1].Name != "project_b" {
		t.Errorf("repos round trip mismatch: %+v", out.Repos)
	}
}

func TestWriteStampsCreatedTime(t *testing.T) {
	stem := t.TempDir()
	before := time.Now().UTC().Add(-time.Minute)
	if err := Write(stem, &Manifest{Branch: "b"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	m, err := Read(stem)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	ts, err := time.Parse(time.RFC3339, m.CreatedAt)
	if err != nil {
		t.Fatalf("created_at %q not RFC 3339: %v", m.CreatedAt, err)
	}
	if ts.Before(before) {
		t.Errorf("created_at %q predates the write", m.CreatedAt)
	}
}

func TestReadMissing(t *testing.T) {
	if _, err := Read(t.TempDir()); err == nil {
		t.Fatal("expected error reading missing manifest")
	}
}

func TestReadInvalid(t *testing.T) {
	stem := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stem, ".graft"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(Path(stem), []byte("not [valid toml"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := Read(stem); err == nil {
		t.Fatal("expected error reading invalid manifest")
	}
}
