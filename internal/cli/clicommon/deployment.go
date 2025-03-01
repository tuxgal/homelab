package clicommon

import (
	"context"
	"fmt"

	"github.com/tuxgal/homelab/internal/deployment"
)

func BuildDeployment(ctx context.Context, cmd string, opts *GlobalCmdOptions) (*deployment.Deployment, error) {
	path, err := configsPath(ctx, cmd, opts)
	if err != nil {
		return nil, err
	}

	dep, err := deployment.FromConfigsPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%s failed while parsing the configs, reason: %w", cmd, err)
	}

	return dep, nil
}
