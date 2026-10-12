package hookhelpers

import (
	"fmt"
	"strings"
)

// createMarker is the all-zeros old-HEAD argument git passes to a
// post-checkout hook when the checkout is a worktree creation.
const createMarker = "0000000000000000000000000000000000000000"

// OnCreate reports whether oldSha is git's worktree-creation marker. An
// empty oldSha is a usage error: the caller must forward the hook's $1.
func OnCreate(oldSha string) (bool, error) {
	if strings.TrimSpace(oldSha) == "" {
		return false, fmt.Errorf("missing <old-sha> argument: forward $1 from the post-checkout hook")
	}
	return oldSha == createMarker, nil
}
