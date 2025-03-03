package fakehost

import (
	"net/netip"

	"github.com/tuxgal/homelab/internal/host"
)

const (
	FakeHostName              = "fakehost"
	FakeHumanFriendlyHostName = "FakeHost"
	FakeHostIPV4              = "10.76.77.78"
	FakeHostNumCPUs           = 32
	FakeHostOS                = "linux"
	FakeHostArch              = "amd64"
	FakeHostDockerPlatform    = "linux/amd64"
)

func NewFakeHostInfo() *host.HostInfo {
	return &host.HostInfo{
		HostName:              FakeHostName,
		HumanFriendlyHostName: FakeHumanFriendlyHostName,
		IPV4:                  netip.MustParseAddr(FakeHostIPV4),
		NumCPUs:               FakeHostNumCPUs,
		OS:                    FakeHostOS,
		Arch:                  FakeHostArch,
		DockerPlatform:        FakeHostDockerPlatform,
	}
}
