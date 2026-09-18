package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Context struct {
	Workspace  string
	Path       string
	Branch     string
	Collection string
	Repos      []string
}

// Env builds the GRAFT_* environment variables passed to lifecycle hooks.
func (c Context) Env() []string {
	return append(os.Environ(),
		"GRAFT_WORKSPACE="+c.Workspace,
		"GRAFT_PATH="+c.Path,
		"GRAFT_BRANCH="+c.Branch,
		"GRAFT_COLLECTION="+c.Collection,
		"GRAFT_REPOS="+strings.Join(c.Repos, "\n"),
	)
}

// RunPostcreate executes the postcreate hook command with the stem directory
// as cwd and the context delivered as environment variables.
func RunPostcreate(command string, ctx Context) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = ctx.Path
	cmd.Env = ctx.Env()
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("postcreate hook failed: %w\n%s", err, strings.TrimSuffix(out.String(), "\n"))
	}
	return nil
}
