package stems

import (
	"fmt"
	"io"
	"strings"

	"github.com/skbolton/graft/internal/gitops"
)

type RepoStatus struct {
	Repo
	Dirty  bool
	Ahead  int
	Behind int
	// Changes holds `git status --porcelain` lines; empty when clean.
	Changes []string
	// Problem is non-empty when health could not be determined.
	Problem string
}

type Report struct {
	Stem  Stem
	Repos []RepoStatus
}

// Status inspects each member repo of a stem: dirty worktrees (uncommitted +
// untracked, detected via `git status --porcelain`) and ahead/behind counts
// against origin/HEAD as known locally. No network operation is performed.
func Status(stem Stem) Report {
	st := Report{Stem: stem}
	for _, r := range stem.Repos {
		rs := RepoStatus{Repo: r}
		switch {
		case r.SourceMissing:
			rs.Problem = "source repo missing"
		case r.Missing:
			rs.Problem = "worktree missing"
		default:
			out, err := gitops.StatusPorcelain(r.Path)
			if err != nil {
				rs.Problem = err.Error()
				break
			}
			if out != "" {
				rs.Dirty = true
				rs.Changes = strings.Split(out, "\n")
			}
			behind, ahead, err := gitops.AheadBehind(r.Path, "origin/HEAD")
			if err != nil {
				rs.Problem = err.Error()
				break
			}
			rs.Ahead = ahead
			rs.Behind = behind
		}
		st.Repos = append(st.Repos, rs)
	}
	return st
}

// WriteStatus renders a status report in the human-readable format: a
// one-line summary per repo, with git's own `git status` output shown
// verbatim beneath dirty repos so user customizations carry through.
func WriteStatus(w io.Writer, st Report) {
	s := st.Stem
	if s.Reason != "" {
		fmt.Fprintf(w, "malformed  %s  (%s)\n", s.Name, s.Reason)
		return
	}
	line := fmt.Sprintf("%s  branch: %s", s.Name, s.Branch)
	if s.Collection != "" {
		line += "  collection: " + s.Collection
	}
	fmt.Fprintln(w, line)
	for _, rs := range st.Repos {
		if rs.Problem != "" {
			fmt.Fprintf(w, "  %s  (%s)\n", rs.Name, rs.Problem)
			continue
		}
		state := "clean"
		if rs.Dirty {
			state = "dirty"
		}
		line := fmt.Sprintf("  %s  %s", rs.Name, state)
		if rs.Ahead > 0 {
			line += fmt.Sprintf("  ahead %d", rs.Ahead)
		}
		if rs.Behind > 0 {
			line += fmt.Sprintf("  behind %d", rs.Behind)
		}
		fmt.Fprintln(w, line)
		if !rs.Dirty {
			continue
		}
		out, err := gitops.StatusLong(rs.Path)
		if err != nil {
			fmt.Fprintf(w, "    (git status failed: %v)\n", err)
			continue
		}
		fmt.Fprintln(w, out)
	}
}

// WriteStatusPorcelain emits the machine-readable framing: one repo header
// record per member repo, followed by that repo's `git status --porcelain`
// lines verbatim. Clean repos contribute only the header.
func WriteStatusPorcelain(w io.Writer, st Report) {
	s := st.Stem
	if s.Reason != "" {
		return
	}
	for _, rs := range st.Repos {
		fmt.Fprintf(w, "repo\t%s\t%s\t%s\n", s.Name, rs.Name, rs.Path)
		for _, line := range rs.Changes {
			fmt.Fprintln(w, line)
		}
	}
}
