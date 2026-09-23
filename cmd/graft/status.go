package main

import (
	"github.com/spf13/cobra"

	"github.com/skbolton/graft/internal/config"
	"github.com/skbolton/graft/internal/stems"
)

func newStatusCommand() *cobra.Command {
	var porcelain bool

	cmd := &cobra.Command{
		Use:   "status [path]",
		Short: "Report per-repo status of the stem containing the path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			stem, err := stems.FindByPath(cfg.Stems, path)
			if err != nil {
				return err
			}
			report := stems.Status(stem)
			if porcelain {
				stems.WriteStatusPorcelain(cmd.OutOrStdout(), report)
				return nil
			}
			stems.WriteStatus(cmd.OutOrStdout(), report)
			return nil
		},
	}

	cmd.Flags().BoolVar(&porcelain, "porcelain", false, "emit machine-readable records")
	return cmd
}
