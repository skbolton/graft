package hookhelpers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/skbolton/graft/internal/gitops"
)

// mainWorktree returns the path of the repo's main worktree as seen from
// dir. It depends only on git, so it works in any worktree of the repo
// regardless of how the worktree was created.
func mainWorktree(dir string) (string, error) {
	out, err := gitops.Run(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err == nil {
		gitDir := strings.TrimSpace(out)
		if filepath.IsAbs(gitDir) {
			return filepath.Dir(gitDir), nil
		}
	}
	out, err = gitops.Run(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("not inside a git repository (run from a worktree of the repo): %w", err)
	}
	for _, line := range strings.Split(out, "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			return filepath.Clean(p), nil
		}
	}
	return "", fmt.Errorf("could not resolve the main worktree of the repo containing %s", dir)
}
