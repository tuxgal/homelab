package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxgal/homelab/internal/cli/clicommon"
	"github.com/tuxgal/homelab/internal/cli/cmds/config"
)

func ConfigCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildConfigCmd(ctx)
	cmd.AddCommand(config.ShowConfigCmd(ctx, opts))
	return cmd
}

func buildConfigCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "config",
		GroupID: clicommon.ConfigCmdGroupID,
		Short:   "Homelab config related commands",
		Long:    `Homelab configuration related commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab config sub-command is required")
		},
	}
}
