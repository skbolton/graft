package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Run executes a git command in dir and returns trimmed stdout.
func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ResolveOriginHEAD returns the commit that origin/HEAD points at in repo.
func ResolveOriginHEAD(repo string) (string, error) {
	out, err := Run(repo, "rev-parse", "origin/HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot resolve origin/HEAD in %s; run `git remote set-head origin -a` in that repo to fix it: %w", repo, err)
	}
	return out, nil
}

// WorktreeAdd creates a worktree at path on a new branch based on origin/HEAD.
func WorktreeAdd(repo, path, branch string) error {
	_, err := Run(repo, "worktree", "add", "-b", branch, path, "origin/HEAD")
	return err
}

// HasWorktree reports whether path looks like a checked-out worktree.
func HasWorktree(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(strings.TrimRight(path, "/") + "/.git")
	return err == nil
}

// WorktreePaths returns the paths of worktrees registered in repo.
func WorktreePaths(repo string) ([]string, error) {
	out, err := Run(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			paths = append(paths, filepath.Clean(p))
		}
	}
	return paths, nil
}

// HasRef reports whether ref resolves in repo.
func HasRef(repo, ref string) bool {
	_, err := Run(repo, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

// CountRange counts commits in a rev-list range expression.
func CountRange(repo, revRange string) (int, error) {
	out, err := Run(repo, "rev-list", "--count", revRange)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(out)
	if err != nil {
		return 0, fmt.Errorf("unexpected rev-list output %q in %s: %w", out, repo, err)
	}
	return n, nil
}

// AheadBehind counts commits behind and ahead of base in repo's HEAD.
func AheadBehind(repo, base string) (behind, ahead int, err error) {
	out, err := Run(repo, "rev-list", "--left-right", "--count", base+"...HEAD")
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(out)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q in %s", out, repo)
	}
	behind, err = strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q in %s: %w", out, repo, err)
	}
	ahead, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q in %s: %w", out, repo, err)
	}
	return behind, ahead, nil
}

// StatusPorcelain returns the `git status --porcelain` output of the
// worktree at dir, empty when clean.
func StatusPorcelain(dir string) (string, error) {
	return Run(dir, "status", "--porcelain")
}

// IsDirty reports whether the worktree at dir has uncommitted or untracked
// changes.
func IsDirty(dir string) (bool, error) {
	out, err := StatusPorcelain(dir)
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// StatusLong returns the long-form `git status` output of the worktree at
// dir, as git renders it under the user's configuration.
func StatusLong(dir string) (string, error) {
	return Run(dir, "status")
}

// WorktreeRemove removes the worktree at path from repo, forcing when
// requested.
func WorktreeRemove(repo, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	_, err := Run(repo, append(args, path)...)
	return err
}

// BranchDelete force-deletes branch in repo.
func BranchDelete(repo, branch string) error {
	_, err := Run(repo, "branch", "-D", branch)
	return err
}
