package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const dirName = ".graft"
const fileName = "manifest.toml"

type Repo struct {
	Name   string `toml:"name"`
	Source string `toml:"source"`
}

type Manifest struct {
	Branch     string `toml:"branch"`
	Collection string `toml:"collection"`
	CreatedAt  string `toml:"created_at"`
	Repos      []Repo `toml:"repos"`
}

// Path returns the location of the manifest inside a stem.
func Path(stemDir string) string {
	return filepath.Join(stemDir, dirName, fileName)
}

// Write creates .graft/ in the stem and writes the manifest, stamping
// created_at with the current UTC time. Emission is hand-rolled to match the
// documented schema byte-for-byte; parsing tolerates any valid TOML.
func Write(stemDir string, m *Manifest) error {
	if m.CreatedAt == "" {
		m.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if err := os.MkdirAll(filepath.Join(stemDir, dirName), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Join(stemDir, dirName), err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "branch = %s\n", strconv.Quote(m.Branch))
	fmt.Fprintf(&b, "collection = %s\n", strconv.Quote(m.Collection))
	fmt.Fprintf(&b, "created_at = %s\n", strconv.Quote(m.CreatedAt))
	for _, r := range m.Repos {
		fmt.Fprintf(&b, "\n[[repos]]\nname = %s\nsource = %s\n", strconv.Quote(r.Name), strconv.Quote(r.Source))
	}
	if err := os.WriteFile(Path(stemDir), []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", Path(stemDir), err)
	}
	return nil
}

// Read parses the manifest inside a stem. Missing and malformed manifests are
// reported as errors; callers treat them as malformed-stem hints.
func Read(stemDir string) (*Manifest, error) {
	data, err := os.ReadFile(Path(stemDir))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", Path(stemDir), err)
	}
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", Path(stemDir), err)
	}
	return &m, nil
}
