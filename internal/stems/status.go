package stems

import (
	"fmt"
	"io"

	"github.com/skbolton/graft/internal/gitops"
)

type RepoStatus struct {
	Repo
	Dirty  bool
	Ahead  int
	Behind int
	// Problem is non-empty when health could not be determined.
	Problem string
}

type Report struct {
	Stem  Stem
	Repos []RepoStatus
}

// Status inspects each member repo of a stem: dirty worktrees (uncommitted +
// untracked) and ahead/behind counts against origin/HEAD as known locally.
// No network operation is performed.
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
			dirty, err := gitops.IsDirty(r.Path)
			if err != nil {
				rs.Problem = err.Error()
				break
			}
			rs.Dirty = dirty
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

// WriteStatus renders a status report in the human-readable format.
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
	}
}
