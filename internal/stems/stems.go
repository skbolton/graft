// Package stems implements the daily-driver operations over existing stems:
// listing, per-repo status, and safe deletion.
package stems

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skbolton/graft/internal/config"
	"github.com/skbolton/graft/internal/discovery"
	"github.com/skbolton/graft/internal/gitops"
	"github.com/skbolton/graft/internal/manifest"
)

type Repo struct {
	Name   string
	Path   string
	Source string
	// Missing is true when the worktree is not registered in the source repo.
	Missing bool
	// SourceMissing is true when the source repo itself is gone, so
	// membership could not be verified.
	SourceMissing bool
}

type Stem struct {
	Name       string
	Dir        string
	Branch     string
	Collection string
	// Reason is non-empty for malformed stems (missing or invalid manifest).
	Reason string
	Repos  []Repo
}

// List reads every child directory of the stems directory, using each
// manifest as a hint for branch identity and membership, then verifying
// membership against git worktree list in the source repos.
func List(stemsDir string) ([]Stem, error) {
	entries, err := os.ReadDir(stemsDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read stems directory %s: %w", stemsDir, err)
	}
	var out []Stem
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		out = append(out, readStem(stemsDir, entry.Name()))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Find reads a single stem by name.
func Find(stemsDir, name string) (Stem, error) {
	dir := filepath.Join(stemsDir, name)
	info, err := os.Stat(dir)
	if err != nil {
		return Stem{}, fmt.Errorf("no stem named %s under %s", name, stemsDir)
	}
	if !info.IsDir() {
		return Stem{}, fmt.Errorf("%s is not a stem directory", dir)
	}
	return readStem(stemsDir, name), nil
}

func readStem(stemsDir, name string) Stem {
	dir := filepath.Join(stemsDir, name)
	s := Stem{Name: name, Dir: dir}
	m, err := manifest.Read(dir)
	if err != nil {
		s.Reason = err.Error()
		return s
	}
	s.Branch = m.Branch
	s.Collection = m.Collection
	for _, r := range m.Repos {
		repo := Repo{Name: r.Name, Path: filepath.Join(dir, r.Name), Source: r.Source}
		if r.Source == "" {
			repo.SourceMissing = true
		} else if _, err := os.Stat(r.Source); err != nil {
			repo.SourceMissing = true
		} else {
			paths, err := gitops.WorktreePaths(r.Source)
			if err != nil {
				repo.SourceMissing = true
			} else {
				repo.Missing = !containsPath(paths, repo.Path)
			}
		}
		s.Repos = append(s.Repos, repo)
	}
	return s
}

func containsPath(paths []string, want string) bool {
	want = filepath.Clean(want)
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

// WriteHuman renders stems in the human-readable list format.
func WriteHuman(w io.Writer, stems []Stem) {
	for _, s := range stems {
		if s.Reason != "" {
			fmt.Fprintf(w, "malformed  %s  (%s)\n", s.Name, s.Reason)
			continue
		}
		line := fmt.Sprintf("%s  branch: %s", s.Name, s.Branch)
		if s.Collection != "" {
			line += "  collection: " + s.Collection
		}
		fmt.Fprintln(w, line)
		for _, r := range s.Repos {
			note := ""
			switch {
			case r.SourceMissing:
				note = "  (source missing)"
			case r.Missing:
				note = "  (missing worktree)"
			}
			fmt.Fprintf(w, "  %s  %s%s\n", r.Name, r.Path, note)
		}
	}
}

// WritePorcelain emits the machine-readable record contract: stem, repo,
// collection, and source records, tab-separated, one per line.
func WritePorcelain(w io.Writer, cfg *config.Config, stems []Stem) error {
	for _, s := range stems {
		fmt.Fprintf(w, "stem\t%s\t%s\n", s.Name, s.Dir)
		for _, r := range s.Repos {
			fmt.Fprintf(w, "repo\t%s\t%s\t%s\n", s.Name, r.Name, r.Path)
		}
	}

	names := make([]string, 0, len(cfg.Collections))
	for name := range cfg.Collections {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(w, "collection\t%s\t%s\n", name, strings.Join(cfg.Collections[name], ","))
	}

	sources, err := discovery.Discover(cfg.Sources)
	if err != nil {
		return err
	}
	for _, s := range sources {
		fmt.Fprintf(w, "source\t%s\t%s\n", s.Name, s.Path)
	}
	return nil
}
