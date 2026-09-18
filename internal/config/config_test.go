package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, home, content string) {
	t.Helper()

	dir := filepath.Join(home, ".config", "graft")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestLoadParsesSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	writeConfig(t, home, `
sources = ["~/repos"]
stems = "~/workspaces"
postcreate = "echo done"

[collections]
backend = ["project_a", "project_b"]
`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.Sources) != 1 || cfg.Sources[0] != filepath.Join(home, "repos") {
		t.Errorf("sources = %v, want ~/repos expanded to %s", cfg.Sources, home)
	}
	if cfg.Stems != filepath.Join(home, "workspaces") {
		t.Errorf("stems = %q, want expanded path", cfg.Stems)
	}
	if cfg.Postcreate != "echo done" {
		t.Errorf("postcreate = %q", cfg.Postcreate)
	}
	got, err := cfg.Collection("backend")
	if err != nil {
		t.Fatalf("collection backend: %v", err)
	}
	if strings.Join(got, ",") != "project_a,project_b" {
		t.Errorf("backend = %v", got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !strings.Contains(err.Error(), ".config/graft/config.toml") {
		t.Errorf("error should identify expected path, got: %v", err)
	}
	if !strings.Contains(err.Error(), "sources = ") {
		t.Errorf("error should include the example config, got: %v", err)
	}
}

func TestXDGConfigHome(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	path, err := Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if path != filepath.Join(xdg, "graft", "config.toml") {
		t.Errorf("path = %q, want %q", path, filepath.Join(xdg, "graft", "config.toml"))
	}

	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`sources = ["/s"]
stems = "/t"`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(); err != nil {
		t.Errorf("load via XDG_CONFIG_HOME: %v", err)
	}
}

func TestMissingRequiredFields(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeConfig(t, home, "stems = \"/t\"\n")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "sources") {
		t.Errorf("expected sources error, got: %v", err)
	}

	writeConfig(t, home, "sources = [\"/s\"]\n")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "stems") {
		t.Errorf("expected stems error, got: %v", err)
	}
}

func TestUnknownCollectionListsAvailable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	writeConfig(t, t.TempDir(), "")
	home, _ := os.UserHomeDir()
	writeConfig(t, home, `
sources = ["/s"]
stems = "/t"

[collections]
backend = ["a"]
frontend = ["b"]
`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	_, err = cfg.Collection("mobile")
	if err == nil {
		t.Fatal("expected unknown-collection error")
	}
	for _, want := range []string{"mobile", "backend", "frontend"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q: %v", want, err)
		}
	}
}
