package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skbolton/graft/internal/grow"
)

func newGrowCommand() *cobra.Command {
	var collection string
	var sources string
	var noHooks bool

	cmd := &cobra.Command{
		Use:   "grow <branch>",
		Short: "Grow a stem of worktrees across selected source repos",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			var names []string
			if sources != "" {
				names = strings.Split(sources, ",")
			}
			return grow.Run(grow.Options{
				Branch:     args[0],
				Collection: collection,
				Sources:    names,
				NoHooks:    noHooks,
				Dir:        cwd,
				Stdout:     cmd.OutOrStdout(),
				Stderr:     cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().StringVarP(&collection, "collection", "c", "", "use a configured collection of repos")
	cmd.Flags().StringVarP(&sources, "sources", "s", "", "comma-separated source repo names")
	cmd.Flags().BoolVar(&noHooks, "no-hooks", false, "skip the postcreate hook")
	return cmd
}
