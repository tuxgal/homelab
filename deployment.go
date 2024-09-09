package main

import (
	"fmt"
)

type deployment struct {
	config   *HomelabConfig
	groups   containerGroupMap
	networks networkMap
}

type containerGroup struct {
	config     *ContainerGroupConfig
	deployment *deployment
	containers containerMap
}

type container struct {
	config *ContainerConfig
	group  *containerGroup
	ips    networkIPMap
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

type containerGroupMap map[string]*containerGroup
type containerMap map[string]*container
type networkMap map[string]*network
type networkIPMap map[string]*containerIP

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
	networks := make(networkMap)
	for _, n := range config.IPAM.Networks.BridgeModeNetworks {
		nt := newBridgeModeNetwork(&d, &n)
		networks[nt.Name()] = nt
	}
	for _, n := range config.IPAM.Networks.ContainerModeNetworks {
		nt := newContainerModeNetwork(&d, &n)
		networks[nt.Name()] = nt
	}
	d.networks = networks

	groups := make(containerGroupMap)
	for _, g := range config.Groups {
		cg := newContainerGroup(&d, &g, &config.Containers)
		groups[cg.Name()] = cg
	}
	d.groups = groups

	return &d
}

func (d *deployment) queryAllContainers() containerMap {
	result := make(containerMap)
	for _, g := range d.groups {
		for _, c := range g.containers {
			result[c.Name()] = c
		}
	}
	return result
}

func (d *deployment) queryAllContainersInGroup(group string) containerMap {
	result := make(containerMap)
	for _, g := range d.groups {
		if g.Name() == group {
			for _, c := range g.containers {
				result[c.Name()] = c
			}
			break
		}
	}
	return result
}

func (d *deployment) queryContainer(group string, container string) *container {
	cn := containerName(group, container)
	for _, g := range d.groups {
		if g.Name() == group {
			for _, c := range g.containers {
				if c.Name() == cn {
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

	containers := make(containerMap)
	for _, c := range *containerConfigs {
		if c.ParentGroup == g.Name() {
			ct := newContainer(&g, &c)
			containers[ct.Name()] = ct
		}
	}
	g.containers = containers

	return &g
}

func (c *containerGroup) Name() string {
	return c.config.Name
}

func (c *containerGroup) String() string {
	return fmt.Sprintf("Group{Name:%s Containers:%s}", c.Name(), c.containers)
}

func (c containerGroupMap) String() string {
	return stringifyMap(c)
}

func newContainer(group *containerGroup, config *ContainerConfig) *container {
	c := container{group: group, config: config}
	cName := config.Name
	gName := group.Name()

	ips := make(networkIPMap)
	for _, n := range group.deployment.networks {
		if n.mode == networkModeBridge {
			for _, c := range n.bridgeModeConfig.Containers {
				if c.Container.Group == gName && c.Container.Container == cName {
					ips[n.Name()] = newBridgeModeContainerIP(n, c.IP)
					break
				}
			}
		} else if n.mode == networkModeContainer {
			for _, c := range n.containerModeConfig.Containers {
				if c.Group == gName && c.Container == cName {
					ips[n.Name()] = newContainerModeContainerIP(n)
					break
				}
			}
		}
	}
	c.ips = ips
	return &c
}

func (c *container) Name() string {
	return containerName(c.group.Name(), c.config.Name)
}

func (c *container) String() string {
	return fmt.Sprintf("Container{Name:%s}", c.Name())
}

func (c containerMap) String() string {
	return stringifyMap(c)
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

func (n *network) Name() string {
	if n.mode == networkModeBridge {
		return n.bridgeModeConfig.Name
	} else if n.mode == networkModeContainer {
		return n.containerModeConfig.Name
	} else {
		log.Fatalf("unknown network, possibly indicating a bug in the code!")
		return "{Network Unknown}"
	}
}

func (n *network) String() string {
	if n.mode == networkModeBridge {
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.Name())
	} else if n.mode == networkModeContainer {
		return fmt.Sprintf("{Network (Container) Name: %s}", n.Name())
	} else {
		log.Fatalf("unknown network, possibly indicating a bug in the code!")
		return "{Network Unknown}"
	}
}

func (n networkMap) String() string {
	return stringifyMap(n)
}

func containerName(group string, container string) string {
	return fmt.Sprintf("%s-%s", group, container)
}
