package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// skipDirs names directories that are never graftable sources, mirroring the
// common vendor-directory exclusion in the source-discovery spec.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"target":       true,
	"dist":         true,
	"build":        true,
	"Pods":         true,
}

type Repo struct {
	Name string
	Path string
}

// Discover scans each sources directory exactly one level deep for git
// repositories. Repo identity is the directory name; duplicates across
// directories collapse to the first occurrence. The result is sorted by name.
func Discover(sourcesDirs []string) ([]Repo, error) {
	seen := map[string]bool{}
	var repos []Repo

	for _, dir := range sourcesDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("scan sources directory %s: %w", dir, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || skipDirs[name] || seen[name] {
				continue
			}
			if !entry.IsDir() || !isGitRepo(filepath.Join(dir, name)) {
				continue
			}
			seen[name] = true
			repos = append(repos, Repo{Name: name, Path: filepath.Join(dir, name)})
		}
	}

	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	return repos, nil
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
