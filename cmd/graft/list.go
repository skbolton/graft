package main

import (
	"github.com/spf13/cobra"

	"github.com/skbolton/graft/internal/config"
	"github.com/skbolton/graft/internal/stems"
)

func newListCommand() *cobra.Command {
	var porcelain bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stems and their member repos",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			found, err := stems.List(cfg.Stems)
			if err != nil {
				return err
			}
			if porcelain {
				return stems.WritePorcelain(cmd.OutOrStdout(), cfg, found)
			}
			stems.WriteHuman(cmd.OutOrStdout(), found)
			return nil
		},
	}

	cmd.Flags().BoolVar(&porcelain, "porcelain", false, "emit machine-readable records")
	return cmd
}
