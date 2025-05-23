package groups

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxgal/homelab/internal/cli/clicommon"
	"github.com/tuxgal/homelab/internal/cli/clicontext"
	"github.com/tuxgal/homelab/internal/cli/errors"
)

func StopCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [group]",
		Short: "Stops one or more containers in the group",
		Long:  `Stops one or more containers in the requested group as specified in the homelab configuration. Containers can be stopped individually, as a group or all groups (by using 'all' as the group name).`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				//nolint:staticcheck
				return fmt.Errorf("Expected exactly one group name argument to be specified, but found %d instead", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execGroupStopCmd(clicontext.HomelabContext(ctx), args[0], opts)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return clicommon.AutoCompleteGroups(ctx, args, "groups stop autocomplete", opts)
		},
	}
}

func execGroupStopCmd(ctx context.Context, group string, opts *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "groups stop", opts)
	if err != nil {
		return err
	}

	var action string
	if group == clicommon.AllGroups {
		action = "Stopping containers in all groups"
	} else {
		action = fmt.Sprintf("Stopping containers in group %s", group)
	}
	return clicommon.ExecContainerGroupCmd(
		ctx,
		"groups stop",
		action,
		group,
		"",
		dep,
		clicommon.ExecStopContainer,
	)
}
