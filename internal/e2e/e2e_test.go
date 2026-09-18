package e2e_test

import (
	"strings"
	"testing"

	"github.com/skbolton/graft/internal/e2e"
)

func TestVersion(t *testing.T) {
	e2e.TempHOME(t)
	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t), "--version")
	if code != 0 {
		t.Fatalf("graft --version: exit %d", code)
	}
	if !strings.HasPrefix(stdout, "graft ") {
		t.Errorf("version output %q does not start with %q", stdout, "graft ")
	}
}

func TestNoArgsPrintsHelp(t *testing.T) {
	e2e.TempHOME(t)
	stdout, _, code := e2e.RunGraft(t, e2e.GraftBinary(t))
	if code != 0 {
		t.Fatalf("graft (no args): exit %d", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("help output missing usage:\n%s", stdout)
	}
}
