package stems

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/skbolton/graft/internal/gitops"
	"github.com/skbolton/graft/internal/manifest"
)

type DeleteOptions struct {
	Name   string
	Force  bool
	Stdout io.Writer
	Stderr io.Writer
}

type repoHealth struct {
	repo     Repo
	dirty    bool
	unpushed int
	problem  string
}

// Delete removes a stem's worktrees, workspace branches, and directory. The
// per-repo summary is always printed before removal; deletion is refused
// while any repo blocks unless Force is set.
func Delete(stemsDir string, opts DeleteOptions) error {
	stemDir := filepath.Join(stemsDir, opts.Name)
	if _, err := os.Stat(stemDir); err != nil {
		return fmt.Errorf("no stem named %s under %s", opts.Name, stemsDir)
	}

	m, err := manifest.Read(stemDir)
	if err != nil {
		return fmt.Errorf("stem %s is malformed (%v); resolve or remove it manually", opts.Name, err)
	}
	branch := m.Branch
	if branch == "" {
		branch = opts.Name
	}

	var healths []repoHealth
	blocked := false
	for _, r := range m.Repos {
		h := inspectRepo(stemDir, r, branch)
		if h.problem == "" && (h.dirty || h.unpushed > 0) {
			blocked = true
		}
		healths = append(healths, h)
	}
	printSummary(opts.Stdout, healths)

	if blocked && !opts.Force {
		return fmt.Errorf("stem %s has uncommitted or unpushed work; use --force to delete anyway", opts.Name)
	}

	var failed []string
	for _, h := range healths {
		if h.problem == "source repo missing" {
			failed = append(failed, fmt.Sprintf("  %s: source repo missing; worktree %s cannot be removed safely", h.repo.Name, h.repo.Path))
			continue
		}
		if !h.repo.Missing {
			if err := gitops.WorktreeRemove(h.repo.Source, h.repo.Path, opts.Force); err != nil {
				failed = append(failed, fmt.Sprintf("  %s: %v", h.repo.Name, err))
				continue
			}
		}
		if err := gitops.BranchDelete(h.repo.Source, branch); err != nil {
			failed = append(failed, fmt.Sprintf("  %s: %v", h.repo.Name, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("stem %s was not fully removed; fix manually:\n%s", opts.Name, strings.Join(failed, "\n"))
	}

	if err := os.RemoveAll(stemDir); err != nil {
		return fmt.Errorf("remove stem directory %s: %w", stemDir, err)
	}
	fmt.Fprintf(opts.Stdout, "deleted  %s\n", stemDir)
	return nil
}

func inspectRepo(stemDir string, r manifest.Repo, branch string) repoHealth {
	h := repoHealth{repo: Repo{Name: r.Name, Path: filepath.Join(stemDir, r.Name), Source: r.Source}}
	if r.Source == "" {
		h.problem = "source repo missing"
		return h
	}
	if _, err := os.Stat(r.Source); err != nil {
		h.problem = "source repo missing"
		return h
	}
	if gitops.HasWorktree(h.repo.Path) {
		dirty, err := gitops.IsDirty(h.repo.Path)
		if err != nil {
			h.problem = err.Error()
			return h
		}
		h.dirty = dirty
	} else {
		h.repo.Missing = true
	}
	n, err := unpushed(r.Source, branch)
	if err != nil {
		h.problem = err.Error()
		return h
	}
	h.unpushed = n
	return h
}

// unpushed counts commits on the workspace branch not available on origin:
// not reachable from refs/remotes/origin/<branch> when that ref exists,
// otherwise commits beyond origin/HEAD.
func unpushed(sourceRepo, branch string) (int, error) {
	if !gitops.HasRef(sourceRepo, "refs/heads/"+branch) {
		return 0, nil
	}
	if gitops.HasRef(sourceRepo, "refs/remotes/origin/"+branch) {
		return gitops.CountRange(sourceRepo, "refs/remotes/origin/"+branch+"..refs/heads/"+branch)
	}
	return gitops.CountRange(sourceRepo, "origin/HEAD..refs/heads/"+branch)
}

func printSummary(w io.Writer, healths []repoHealth) {
	for _, h := range healths {
		parts := []string{h.repo.Name}
		switch {
		case h.problem != "":
			parts = append(parts, h.problem)
		case h.dirty:
			parts = append(parts, "dirty")
		default:
			parts = append(parts, "clean")
		}
		if h.unpushed > 0 {
			parts = append(parts, fmt.Sprintf("%d unpushed", h.unpushed))
		}
		fmt.Fprintln(w, "  "+strings.Join(parts, "  "))
	}
}
