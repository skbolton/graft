package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/skbolton/graft/internal/hookhelpers"
)

func newHookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Helpers for a repo's own git hooks",
	}
	cmd.AddCommand(newHookOnCreateCommand())
	cmd.AddCommand(newHookCopyCommand())
	return cmd
}

func newHookOnCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "on-create <old-sha>",
		Short: "Succeed only when a post-checkout hook was fired by worktree creation",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: graft hook on-create <old-sha> (forward $1 from the hook)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := hookhelpers.OnCreate(args[0])
			if err != nil {
				return err
			}
			if !ok {
				os.Exit(1)
			}
			return nil
		},
	}
}

func newHookCopyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "copy <path>...",
		Short: "Copy files or directories from the repo's main worktree into this worktree",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return hookhelpers.Copy(args, cmd.ErrOrStderr())
		},
	}
}
