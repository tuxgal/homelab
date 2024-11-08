package testutils

import (
	"context"

	"github.com/tuxgal/homelab/internal/cli/version"
	"github.com/tuxgal/homelab/internal/cmdexec"
	"github.com/tuxgal/homelab/internal/cmdexec/fakecmdexec"
	"github.com/tuxgal/homelab/internal/docker"
	"github.com/tuxgal/homelab/internal/docker/fakedocker"
	"github.com/tuxgal/homelab/internal/host"
	"github.com/tuxgal/homelab/internal/host/fakehost"
	"github.com/tuxgal/homelab/internal/inspect"
	"github.com/tuxgal/homelab/internal/log"
	"github.com/tuxgal/homelab/internal/user"
	"github.com/tuxgal/homelab/internal/user/fakeuser"
	"github.com/tuxgal/tuxlogi"
)

type TestContextInfo struct {
	InspectLevel               inspect.HomelabInspectLevel
	Logger                     tuxlogi.Logger
	Version                    *version.VersionInfo
	Executor                   cmdexec.Executor
	DockerHost                 docker.APIClient
	ContainerPurgeKillAttempts uint32
	UseRealUserInfo            bool
	UseRealHostInfo            bool
	UseRealExecutor            bool
}

func NewVanillaTestContext() context.Context {
	return NewTestContext(
		&TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		})
}

func NewTestContext(info *TestContextInfo) context.Context {
	ctx := context.Background()
	if info.InspectLevel != inspect.HomelabInspectLevelNone {
		ctx = inspect.WithHomelabInspectLevel(ctx, info.InspectLevel)
	}
	if info.Logger != nil {
		ctx = log.WithLogger(ctx, info.Logger)
	} else {
		ctx = log.WithLogger(ctx, NewTestLogger())
	}
	if info.Version != nil {
		ctx = version.WithVersionInfo(ctx, info.Version)
	}
	if !info.UseRealUserInfo {
		ctx = user.WithUserInfo(ctx, fakeuser.NewFakeUserInfo())
	}
	if !info.UseRealHostInfo {
		ctx = host.WithHostInfo(ctx, fakehost.NewFakeHostInfo())
	}
	if info.Executor != nil {
		ctx = cmdexec.WithExecutor(ctx, info.Executor)
	} else if !info.UseRealExecutor {
		ctx = cmdexec.WithExecutor(ctx, fakecmdexec.NewEmptyFakeExecutor())
	}
	if info.DockerHost != nil {
		ctx = docker.WithAPIClient(ctx, info.DockerHost)
	}
	if info.ContainerPurgeKillAttempts != 0 {
		ctx = docker.WithContainerPurgeKillAttempts(ctx, info.ContainerPurgeKillAttempts)
	}
	return ctx
}
