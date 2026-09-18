package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const Example = `sources = ["~/repos"]   # directories of pristine source checkouts
stems = "~/workspaces"         # where workstream directories are created

[collections]
backend = ["project_a", "project_b"]`

type Config struct {
	Sources     []string            `toml:"sources"`
	Stems       string              `toml:"stems"`
	Collections map[string][]string `toml:"collections"`
	Postcreate  string              `toml:"postcreate"`
}

// Path returns the XDG-aware location of the graft config directory's config.toml.
func Path() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "graft", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}
	return filepath.Join(home, ".config", "graft", "config.toml"), nil
}

// Load reads and validates config.toml at the XDG-aware location.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no graft config found at %s; create it with at least:\n\n%s", path, Example)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if len(cfg.Sources) == 0 {
		return nil, fmt.Errorf("%s: sources must list at least one directory", path)
	}
	if cfg.Stems == "" {
		return nil, fmt.Errorf("%s: stems directory is not set", path)
	}
	for i, s := range cfg.Sources {
		cfg.Sources[i] = expandPath(s)
	}
	cfg.Stems = expandPath(cfg.Stems)
	return &cfg, nil
}

// Collection returns the repo names in a named collection, erroring with the
// available names when the collection does not exist.
func (c *Config) Collection(name string) ([]string, error) {
	repos, ok := c.Collections[name]
	if !ok {
		return nil, fmt.Errorf("unknown collection %q (available: %s)", name, c.collectionNames())
	}
	return repos, nil
}

func (c *Config) collectionNames() string {
	names := make([]string, 0, len(c.Collections))
	for n := range c.Collections {
		names = append(names, n)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "none configured"
	}
	return strings.Join(names, ", ")
}

func expandPath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}
