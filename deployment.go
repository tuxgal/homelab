package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

const (
	osLinux   = "linux"
	archAmd64 = "amd64"
	archArm64 = "arm64"
)

type deployment struct {
	config   *HomelabConfig
	groups   containerGroupMap
	networks networkMap
	host     hostInfo
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

type hostInfo struct {
	numCPUs               int
	os                    string
	arch                  string
	dockerPlatform        string
	hostName              string
	humanFriendlyHostName string
	ip                    net.IP
	allowedContainers     stringSet
	config                *HostConfig
}

type containerGroupMap map[string]*containerGroup
type containerMap map[string]*container
type containerList []*container
type networkMap map[string]*network
type networkIPMap map[string]*containerIP
type stringSet map[string]bool

func buildDeployment() (*deployment, error) {
	c := HomelabConfig{}
	err := c.parse()
	if err != nil {
		return nil, err
	}

	err = c.validate()
	if err != nil {
		return nil, err
	}

	return newDeployment(&c), nil
}

func newDeployment(config *HomelabConfig) *deployment {
	d := deployment{config: config}

	// First build the networks as it will be looked up while building
	// the container groups and containers within.
	d.populateNetworks()
	d.populateGroups()
	d.populateHostInfo()
	return &d
}

func (d *deployment) populateNetworks() {
	networks := make(networkMap)
	for _, n := range d.config.IPAM.Networks.BridgeModeNetworks {
		nt := newBridgeModeNetwork(d, &n)
		networks[nt.Name()] = nt
	}
	for _, n := range d.config.IPAM.Networks.ContainerModeNetworks {
		nt := newContainerModeNetwork(d, &n)
		networks[nt.Name()] = nt
	}
	d.networks = networks
}

func (d *deployment) populateGroups() {
	groups := make(containerGroupMap)
	for _, g := range d.config.Groups {
		cg := newContainerGroup(d, &g, &d.config.Containers)
		groups[cg.Name()] = cg
	}
	d.groups = groups
}

func (d *deployment) populateHostInfo() {
	d.host.numCPUs = runtime.NumCPU()
	d.host.os = runtime.GOOS
	d.host.arch = runtime.GOARCH
	d.host.dockerPlatform = archToDockerPlatform(d.host.arch)
	log.Debugf("Num CPUs = %d", d.host.numCPUs)
	log.Debugf("OS = %s", d.host.os)
	log.Debugf("Arch = %s", d.host.arch)
	log.Debugf("Docker Platform = %s", d.host.dockerPlatform)
	if d.host.os != osLinux {
		log.Fatalf("Only linux OS is supported, found OS: %s", d.host.os)
	}
	if d.host.arch != archAmd64 && d.host.arch != archArm64 {
		log.Fatalf("Only amd64 and arm64 platforms are supported, found Arch: %s", d.host.arch)
	}

	hostName, err := os.Hostname()
	if err != nil {
		log.Fatalf("Unable to determine the current machine's host name, %v", err)
	}
	d.host.humanFriendlyHostName = hostName
	d.host.hostName = strings.ToLower(d.host.humanFriendlyHostName)

	conn, err := net.Dial("udp", "10.1.1.1:1234")
	if err != nil {
		log.Fatalf("Unable to determine the current machine's IP, %v", err)
	}
	defer conn.Close()
	d.host.ip = conn.LocalAddr().(*net.UDPAddr).IP

	d.host.allowedContainers = make(stringSet)
	for _, h := range d.config.Hosts {
		if h.Name == d.host.hostName {
			d.host.config = &h
			for _, c := range d.host.config.AllowedContainers {
				d.host.allowedContainers[containerName(c.Group, c.Container)] = true
			}
			break
		}
	}

	log.Debugf("Host name: %s", d.host.hostName)
	log.Debugf("Human Friendly Host name: %s", d.host.humanFriendlyHostName)
	log.Debugf("Host IP: %s", d.host.ip)
	log.Debugf("Allowed Containers: %s", d.host.allowedContainers)
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

func (c *container) isAllowedOnCurrentHost() bool {
	return c.group.deployment.host.allowedContainers[c.Name()]
}

func (c *container) start() error {
	return fmt.Errorf("container start is not yet supported")
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

func newBridgeModeContainerIP(network *network, ip string) *containerIP {
	return &containerIP{network: network, IP: ip}
}

func newContainerModeContainerIP(network *network) *containerIP {
	return &containerIP{network: network}
}

func containerName(group string, container string) string {
	return fmt.Sprintf("%s-%s", group, container)
}

func archToDockerPlatform(arch string) string {
	switch arch {
	case archAmd64:
		return "linux/amd64"
	case archArm64:
		return "linux/arm64/v8"
	default:
		return fmt.Sprintf("unsupported-docker-arch-%s", arch)
	}
}
