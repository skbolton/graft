package main

import (
	"github.com/spf13/cobra"

	"github.com/skbolton/graft/internal/config"
	"github.com/skbolton/graft/internal/stems"
)

func newDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <stem>",
		Short: "Delete a stem, refusing when work would be lost",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return stems.Delete(cfg.Stems, stems.DeleteOptions{
				Name:   args[0],
				Force:  force,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "delete even when work would be lost")
	return cmd
}
