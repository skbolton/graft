package main

import (
	"github.com/spf13/cobra"

	"github.com/skbolton/graft/internal/config"
	"github.com/skbolton/graft/internal/stems"
)

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <stem>",
		Short: "Report per-repo health of a stem",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			stem, err := stems.Find(cfg.Stems, args[0])
			if err != nil {
				return err
			}
			stems.WriteStatus(cmd.OutOrStdout(), stems.Status(stem))
			return nil
		},
	}
	return cmd
}
