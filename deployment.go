package main

import (
	"fmt"
	"strings"
)

type deployment struct {
	config   *HomelabConfig
	groups   containerGroupList
	networks networkList
}

type containerGroup struct {
	config     *ContainerGroupConfig
	deployment *deployment
	containers containerList
}

type container struct {
	config           *ContainerConfig
	group            *containerGroup
	networkEndpoints containerIPList
}

type containerIP struct {
	network *network
	IP      string
}

const (
	networkModeUnknown networkMode = iota
	networkModeBridge
	networkModeContainer
)

type networkMode uint8

type network struct {
	deployment          *deployment
	mode                networkMode
	bridgeModeConfig    *BridgeModeNetworkConfig
	containerModeConfig *ContainerModeNetworkConfig
}

type containerGroupList []*containerGroup
type containerList []*container
type networkList []*network
type containerIPList []*containerIP

func newBridgeModeContainerIP(network *network, ip string) *containerIP {
	return &containerIP{network: network, IP: ip}
}

func newContainerModeContainerIP(network *network) *containerIP {
	return &containerIP{network: network}
}

func newDeployment(config *HomelabConfig) *deployment {
	d := deployment{config: config}

	// First build the networks as it will be looked up while building
	// the container groups and containers within.
	networks := make(networkList, 0, len(config.IPAM.Networks.BridgeModeNetworks)+len(config.IPAM.Networks.ContainerModeNetworks))
	for _, n := range config.IPAM.Networks.BridgeModeNetworks {
		networks = append(networks, newBridgeModeNetwork(&d, &n))
	}
	for _, n := range config.IPAM.Networks.ContainerModeNetworks {
		networks = append(networks, newContainerModeNetwork(&d, &n))
	}
	d.networks = networks

	groups := make(containerGroupList, 0, len(config.Groups))
	for _, g := range config.Groups {
		groups = append(groups, newContainerGroup(&d, &g, &config.Containers))
	}
	d.groups = groups

	return &d
}

func (d *deployment) queryAllContainers() containerList {
	result := make(containerList, 0)
	for _, g := range d.groups {
		result = append(result, g.containers...)
	}
	return result
}

func (d *deployment) queryAllContainersInGroup(group string) containerList {
	result := make(containerList, 0)
	for _, g := range d.groups {
		if g.config.Name == group {
			result = append(result, g.containers...)
			break
		}
	}
	return result
}

func (d *deployment) queryContainer(group string, container string) *container {
	for _, g := range d.groups {
		if g.config.Name == group {
			for _, c := range g.containers {
				if c.config.Name == container {
					return c
				}
			}
		}
	}
	return nil
}

func (d *deployment) String() string {
	return fmt.Sprintf("Deployment{Groups:%s, Networks:%s}", d.groups, d.networks)
}

func newContainerGroup(dep *deployment, groupConfig *ContainerGroupConfig, containerConfigs *[]ContainerConfig) *containerGroup {
	g := containerGroup{
		deployment: dep,
		config:     groupConfig,
	}

	containers := make(containerList, 0)
	for _, c := range *containerConfigs {
		if c.ParentGroup == groupConfig.Name {
			containers = append(containers, newContainer(&g, &c))
		}
	}
	g.containers = containers

	return &g
}

func (c *containerGroup) String() string {
	return fmt.Sprintf("Group{Name:%s Containers:%s}", c.config.Name, c.containers)
}

func (c containerGroupList) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	if len(c) > 0 {
		sb.WriteString(c[0].String())
	} else {
		sb.WriteString("empty")
	}
	for i := 1; i < len(c); i++ {
		sb.WriteString(fmt.Sprintf(", %s", c[i]))
	}
	sb.WriteString("]")
	return sb.String()
}

func newContainer(group *containerGroup, config *ContainerConfig) *container {
	c := container{group: group, config: config}
	cName := config.Name
	gName := group.config.Name

	networkEndpoints := make(containerIPList, 0)
	for _, n := range group.deployment.networks {
		if n.mode == networkModeBridge {
			for _, c := range n.bridgeModeConfig.Containers {
				if c.Container.Group == gName && c.Container.Container == cName {
					networkEndpoints = append(networkEndpoints, newBridgeModeContainerIP(n, c.IP))
					break
				}
			}
		} else if n.mode == networkModeContainer {
			for _, c := range n.containerModeConfig.Containers {
				if c.Group == gName && c.Container == cName {
					networkEndpoints = append(networkEndpoints, newContainerModeContainerIP(n))
					break
				}
			}
		}
	}
	c.networkEndpoints = networkEndpoints
	return &c
}

func (c *container) String() string {
	return fmt.Sprintf("Container{Group:%s Name:%s}", c.group.config.Name, c.config.Name)
}

func (c containerList) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	if len(c) > 0 {
		sb.WriteString(c[0].String())
	} else {
		sb.WriteString("empty")
	}
	for i := 1; i < len(c); i++ {
		sb.WriteString(fmt.Sprintf(", %s", c[i]))
	}
	sb.WriteString("]")
	return sb.String()
}

func newBridgeModeNetwork(dep *deployment, config *BridgeModeNetworkConfig) *network {
	n := network{
		deployment:       dep,
		mode:             networkModeBridge,
		bridgeModeConfig: config,
	}
	return &n
}

func newContainerModeNetwork(dep *deployment, config *ContainerModeNetworkConfig) *network {
	n := network{
		deployment:          dep,
		mode:                networkModeContainer,
		containerModeConfig: config,
	}
	return &n
}

func (n *network) String() string {
	if n.mode == networkModeBridge {
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.bridgeModeConfig.Name)
	} else if n.mode == networkModeContainer {
		return fmt.Sprintf("{Network (Container) Name: %s}", n.containerModeConfig.Name)
	} else {
		return "{Network Unknown}"
	}
}

func (n networkList) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	if len(n) > 0 {
		sb.WriteString(n[0].String())
	} else {
		sb.WriteString("empty")
	}
	for i := 1; i < len(n); i++ {
		sb.WriteString(fmt.Sprintf(", %s", n[i]))
	}
	sb.WriteString("]")
	return sb.String()
}
