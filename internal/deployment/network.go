package deployment

import (
	"context"
	"fmt"
	"net/netip"

	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/tuxgal/homelab/internal/config"
	"github.com/tuxgal/homelab/internal/docker"
	"github.com/tuxgal/homelab/internal/newutils"
)

type Network struct {
	networkName       string
	mode              NetworkMode
	bridgeModeInfo    *bridgeModeNetworkInfo
	containerModeInfo *containerModeNetworkInfo
}

type bridgeModeNetworkInfo struct {
	priority          int
	hostInterfaceName string
	v4CIDR            netip.Prefix
	v4Gateway         netip.Addr
	enableV6          bool
	v6CIDR            netip.Prefix
	v6Gateway         netip.Addr
}

type containerModeNetworkInfo struct {
	container config.ContainerReference
}

type NetworkMap map[string]*Network
type NetworkList []*Network

const (
	NetworkModeUnknown NetworkMode = iota
	NetworkModeBridge
	NetworkModeContainer
)

type NetworkMode uint8

func newBridgeModeNetwork(name string, priority int, info *bridgeModeNetworkInfo) *Network {
	n := Network{
		networkName:    name,
		mode:           NetworkModeBridge,
		bridgeModeInfo: info,
	}
	return &n
}

func newContainerModeNetwork(name string, info *containerModeNetworkInfo) *Network {
	n := Network{
		networkName:       name,
		mode:              NetworkModeContainer,
		containerModeInfo: info,
	}
	return &n
}

func (n *Network) Create(ctx context.Context, dc *docker.Client) (bool, error) {
	if n.mode == NetworkModeContainer {
		return false, fmt.Errorf("container mode network %s cannot be created", n.Name())
	}

	// TODO: Validate that the existing network and the new network have
	// exactly the same properties if we choose to reuse the existing
	// network, and display a warning when they differ.
	if !dc.NetworkExists(ctx, n.Name()) {
		err := dc.CreateNetwork(ctx, n.Name(), n.createOptions())
		if err != nil {
			return false, err
		}
		log(ctx).Infof("Created network %s", n.Name())
		log(ctx).InfoEmpty()
		return true, nil
	}
	log(ctx).Debugf("Not re-creating existing network %s", n.Name())
	return false, nil
}

func (n *Network) Delete(ctx context.Context, dc *docker.Client) (bool, error) {
	if n.mode == NetworkModeContainer {
		return false, fmt.Errorf("container mode network %s cannot be deleted", n.Name())
	}

	if !dc.NetworkExists(ctx, n.Name()) {
		return false, nil
	}

	err := dc.RemoveNetwork(ctx, n.Name())
	if err != nil {
		return false, err
	}

	log(ctx).Infof("Deleted network %s", n.Name())
	log(ctx).InfoEmpty()
	return true, nil
}

func (n *Network) connectContainer(ctx context.Context, dc *docker.Client, containerName, ipv4 string, ipv6 string) error {
	return dc.ConnectContainerToBridgeModeNetwork(ctx, containerName, n.Name(), ipv4, ipv6)
}

//nolint:nolintlint,unused // TODO: Remove this after this function is used.
func (n *Network) disconnectContainer(ctx context.Context, dc *docker.Client, containerName string) error {
	return dc.DisconnectContainerFromNetwork(ctx, containerName, n.Name())
}

func (n *Network) createOptions() dnetwork.CreateOptions {
	if n.mode != NetworkModeBridge {
		panic("Only bridge mode network creation is possible")
	}

	ipamConfigs := []dnetwork.IPAMConfig{
		{
			Subnet:  n.bridgeModeInfo.v4CIDR.String(),
			Gateway: n.bridgeModeInfo.v4Gateway.String(),
		},
	}
	if n.bridgeModeInfo.enableV6 {
		ipamConfigs = append(ipamConfigs, dnetwork.IPAMConfig{
			Subnet:  n.bridgeModeInfo.v6CIDR.String(),
			Gateway: n.bridgeModeInfo.v6Gateway.String(),
		})
	}

	return dnetwork.CreateOptions{
		Driver:     "bridge",
		Scope:      "local",
		EnableIPv6: newutils.NewBool(false),
		IPAM: &dnetwork.IPAM{
			Driver: "default",
			Config: ipamConfigs,
		},
		Internal:   false,
		Attachable: false,
		Ingress:    false,
		ConfigOnly: false,
		Options: map[string]string{
			"com.docker.network.bridge.enable_icc":           "true",
			"com.docker.network.bridge.enable_ip_masquerade": "true",
			// There is no equivalent option right now for IPv6 as per
			// docker documentation.
			"com.docker.network.bridge.host_binding_ipv4": n.bridgeModeInfo.v4Gateway.String(),
			"com.docker.network.bridge.name":              n.bridgeModeInfo.hostInterfaceName,
			"com.docker.network.bridge.mtu":               "1500",
		},
	}
}

func (n *Network) Name() string {
	return n.networkName
}

func (n *Network) Mode() NetworkMode {
	return n.mode
}

func (n *Network) String() string {
	switch n.mode {
	case NetworkModeBridge:
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.Name())
	case NetworkModeContainer:
		return fmt.Sprintf("{Network (Container) Name: %s}", n.Name())
	default:
		panic("unknown network mode, possibly indicating a bug in the code!")
	}
}
