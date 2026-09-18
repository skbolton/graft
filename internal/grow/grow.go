package grow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skbolton/graft/internal/config"
	"github.com/skbolton/graft/internal/discovery"
	"github.com/skbolton/graft/internal/gitops"
	"github.com/skbolton/graft/internal/hooks"
	"github.com/skbolton/graft/internal/manifest"
)

type Options struct {
	Branch     string
	Collection string
	Sources    []string
	NoHooks    bool
	Dir        string
	Stdout     io.Writer
	Stderr     io.Writer
}

type outcome struct {
	name string
	ok   bool
	warn string
	err  string
}

const noSelectionHint = `no repos selected: pass -c <collection>, -s <repo,repo>, or run graft grow from inside a source repo`

// Run grows a stem: one directory per branch under the configured stems
// directory, containing one worktree per selected source repo, all checked
// out to the branch. Failures on individual repos never abort the others.
func Run(opts Options) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := validateStemName(opts.Branch); err != nil {
		return err
	}

	stemDir := filepath.Join(cfg.Stems, opts.Branch)
	if _, err := os.Stat(stemDir); err == nil {
		return fmt.Errorf("stem %q already exists at %s", opts.Branch, stemDir)
	}

	discovered, err := discovery.Discover(cfg.Sources)
	if err != nil {
		return err
	}
	byName := make(map[string]discovery.Repo, len(discovered))
	for _, r := range discovered {
		byName[r.Name] = r
	}

	selected, err := selectRepos(cfg, opts, byName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(stemDir, 0o755); err != nil {
		return fmt.Errorf("create stem directory %s: %w", stemDir, err)
	}

	sort.Strings(selected)
	var outcomes []outcome
	var created []discovery.Repo
	for _, name := range selected {
		repo := byName[name]
		o := growRepo(stemDir, opts.Branch, name, repo)
		outcomes = append(outcomes, o)
		if o.ok {
			created = append(created, repo)
		}
	}

	m := &manifest.Manifest{
		Branch:     opts.Branch,
		Collection: opts.Collection,
		Repos:      make([]manifest.Repo, 0, len(created)),
	}
	for _, r := range created {
		m.Repos = append(m.Repos, manifest.Repo{Name: r.Name, Source: r.Path})
	}
	if err := manifest.Write(stemDir, m); err != nil {
		return err
	}

	var failed bool
	for _, o := range outcomes {
		switch {
		case o.err != "":
			failed = true
			fmt.Fprintf(opts.Stderr, "failed   %s: %s\n", o.name, o.err)
		case o.warn != "":
			fmt.Fprintf(opts.Stderr, "warning  %s: %s\n", o.name, o.warn)
		default:
			fmt.Fprintf(opts.Stderr, "created  %s\n", o.name)
		}
	}

	if !opts.NoHooks && cfg.Postcreate != "" {
		ctx := hooks.Context{
			Workspace:  opts.Branch,
			Path:       stemDir,
			Branch:     opts.Branch,
			Collection: opts.Collection,
			Repos:      repoNames(created),
		}
		if err := hooks.RunPostcreate(cfg.Postcreate, ctx); err != nil {
			fmt.Fprintf(opts.Stderr, "warning: %v\n", err)
		}
	}

	fmt.Fprintln(opts.Stdout, stemDir)

	if failed {
		return fmt.Errorf("stem %s grew with %d failing repo(s); see the summary above", opts.Branch, countFailures(outcomes))
	}
	return nil
}

func growRepo(stemDir, branch, name string, repo discovery.Repo) outcome {
	if repo.Name == "" {
		return outcome{name: name, err: "no source repo found with this name"}
	}
	if _, err := gitops.ResolveOriginHEAD(repo.Path); err != nil {
		return outcome{name: name, err: err.Error()}
	}
	path := filepath.Join(stemDir, name)
	if err := gitops.WorktreeAdd(repo.Path, path, branch); err != nil {
		if gitops.HasWorktree(path) {
			return outcome{name: name, ok: true, warn: fmt.Sprintf("worktree created but the repo's post-checkout hook failed: %v", err)}
		}
		return outcome{name: name, err: err.Error()}
	}
	return outcome{name: name, ok: true}
}

func selectRepos(cfg *config.Config, opts Options, byName map[string]discovery.Repo) ([]string, error) {
	var names []string
	if opts.Collection != "" {
		repos, err := cfg.Collection(opts.Collection)
		if err != nil {
			return nil, err
		}
		names = append(names, repos...)
	}
	if len(opts.Sources) > 0 {
		names = append(names, opts.Sources...)
	}

	if len(names) == 0 {
		if repo, ok := cwdRepo(opts.Dir, byName); ok {
			return []string{repo}, nil
		}
		return nil, fmt.Errorf(noSelectionHint)
	}
	return dedupe(names), nil
}

// cwdRepo finds the deepest discovered repo containing dir.
func cwdRepo(dir string, byName map[string]discovery.Repo) (string, bool) {
	if dir == "" {
		return "", false
	}
	best := ""
	for name, r := range byName {
		if dir == r.Path || strings.HasPrefix(dir, r.Path+string(filepath.Separator)) {
			if len(r.Path) > len(best) {
				best = name
			}
		}
	}
	return best, best != ""
}

func dedupe(names []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return out
}

func validateStemName(branch string) error {
	if branch == "" || branch == "." || branch == ".." || strings.ContainsAny(branch, "/\x00") {
		return fmt.Errorf("branch %q must be usable as a directory name (no slashes)", branch)
	}
	return nil
}

func repoNames(repos []discovery.Repo) []string {
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name
	}
	return names
}

func countFailures(outcomes []outcome) int {
	n := 0
	for _, o := range outcomes {
		if o.err != "" {
			n++
		}
	}
	return n
}
