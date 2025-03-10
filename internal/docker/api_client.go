package docker

import (
	"context"
	"io"

	dcontainer "github.com/docker/docker/api/types/container"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	dclient "github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type APIClient interface {
	Close() error

	ContainerCreate(ctx context.Context, config *dcontainer.Config, hostConfig *dcontainer.HostConfig, networkingConfig *dnetwork.NetworkingConfig, platform *ocispec.Platform, containerName string) (dcontainer.CreateResponse, error)
	ContainerInspect(ctx context.Context, containerName string) (dcontainer.InspectResponse, error)
	ContainerKill(ctx context.Context, containerName, signal string) error
	ContainerRemove(ctx context.Context, containerName string, options dcontainer.RemoveOptions) error
	ContainerStart(ctx context.Context, containerName string, options dcontainer.StartOptions) error
	ContainerStop(ctx context.Context, containerName string, options dcontainer.StopOptions) error

	ImageList(ctx context.Context, options dimage.ListOptions) ([]dimage.Summary, error)
	ImagePull(ctx context.Context, refStr string, options dimage.PullOptions) (io.ReadCloser, error)

	NetworkConnect(ctx context.Context, networkName, containerName string, config *dnetwork.EndpointSettings) error
	NetworkCreate(ctx context.Context, networkName string, options dnetwork.CreateOptions) (dnetwork.CreateResponse, error)
	NetworkDisconnect(ctx context.Context, networkName, containerName string, force bool) error
	NetworkList(ctx context.Context, options dnetwork.ListOptions) ([]dnetwork.Summary, error)
	NetworkRemove(ctx context.Context, networkName string) error
}

func MustRealAPIClient(ctx context.Context) APIClient {
	d, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		log(ctx).Fatalf("Failed to create a new docker API client, reason: %v", err)
	}
	return d
}
