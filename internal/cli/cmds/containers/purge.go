package containers

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxgal/homelab/internal/cli/clicommon"
	"github.com/tuxgal/homelab/internal/cli/clicontext"
	"github.com/tuxgal/homelab/internal/cli/errors"
)

func PurgeCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "purge [container]",
		Short: "Purges the container",
		Long:  `Purges the requested container as specified in the homelab configuration. The name is specified in the group/container format.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				//nolint:staticcheck
				return fmt.Errorf("Expected exactly one container name argument to be specified, but found %d instead", len(args))
			}
			_, _, err := validateContainerName(args[0])
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execContainerPurgeCmd(clicontext.HomelabContext(ctx), args[0], opts)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return clicommon.AutoCompleteContainers(ctx, args, "containers purge autocomplete", opts)
		},
	}
}

func execContainerPurgeCmd(ctx context.Context, containerArg string, opts *clicommon.GlobalCmdOptions) error {
	g, ct := mustContainerName(containerArg)
	dep, err := clicommon.BuildDeployment(ctx, "containers purge", opts)
	if err != nil {
		return err
	}

	return clicommon.ExecContainerGroupCmd(
		ctx,
		"containers purge",
		fmt.Sprintf("Purging container %s in group %s", ct, g),
		g,
		ct,
		dep,
		clicommon.ExecPurgeContainer,
	)
}
