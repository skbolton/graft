package gitops

import (
	"fmt"
	"os"
	"os/exec"
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
