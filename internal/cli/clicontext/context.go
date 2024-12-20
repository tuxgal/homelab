package clicontext

import (
	"context"

	"github.com/tuxgal/homelab/internal/cmdexec"
	"github.com/tuxgal/homelab/internal/docker"
	"github.com/tuxgal/homelab/internal/host"
	"github.com/tuxgal/homelab/internal/user"
)

func HomelabContext(ctx context.Context) context.Context {
	if _, found := user.UserInfoFromContext(ctx); !found {
		ctx = user.WithUserInfo(ctx, user.NewUserInfo(ctx))
	}
	if _, found := host.HostInfoFromContext(ctx); !found {
		ctx = host.WithHostInfo(ctx, host.NewHostInfo(ctx))
	}
	if _, found := cmdexec.ExecutorFromContext(ctx); !found {
		ctx = cmdexec.WithExecutor(ctx, cmdexec.NewExecutor())
	}
	if _, found := docker.APIClientFromContext(ctx); !found {
		ctx = docker.WithAPIClient(ctx, docker.MustRealAPIClient(ctx))
	}
	return ctx
}
