package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags "-X main.version=<ver>".
var version = "dev"

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "graft",
		Short:         "Multi-repo git worktree workspaces",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.Version = version
	root.SetVersionTemplate("graft {{.Version}}\n")
	return root
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
