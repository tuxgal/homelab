package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/netip"
	"os"
	"path/filepath"

	"github.com/TwiN/deepmerge"
	"gopkg.in/yaml.v3"
)

// HomelabConfig represents the entire homelab deployment configuration.
type HomelabConfig struct {
	Global     GlobalConfig           `yaml:"global"`
	IPAM       IPAMConfig             `yaml:"ipam"`
	Hosts      []HostConfig           `yaml:"hosts"`
	Groups     []ContainerGroupConfig `yaml:"groups"`
	Containers []ContainerConfig      `yaml:"containers"`
}

// GlobalConfig represents the configuration that will be applied
// across the entire homelab deployment.
type GlobalConfig struct {
	Env       []GlobalEnvConfig     `yaml:"env"`
	MountDefs []MountConfig         `yaml:"mountDefs"`
	Container GlobalContainerConfig `yaml:"container"`
}

// GlobalContainerConfig represents container related configuration that
// will be applied globally across all containers.
type GlobalContainerConfig struct {
	StopSignal    string               `yaml:"stopSignal"`
	StopTimeout   int                  `yaml:"stopTimeout"`
	RestartPolicy string               `yaml:"restartPolicy"`
	DomainName    string               `yaml:"domainName"`
	DNSSearch     []string             `yaml:"dnsSearch"`
	Env           []ContainerEnvConfig `yaml:"env"`
	Mounts        []MountConfig        `yaml:"mounts"`
	Labels        []LabelConfig        `yaml:"labels"`
}

// GlobalEnvConfig is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type GlobalEnvConfig struct {
	Var          string `yaml:"var"`
	Value        string `yaml:"value"`
	ValueCommand string `yaml:"valueCommand"`
}

// IPAMConfig represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAMConfig struct {
	Networks NetworksConfig `yaml:"networks"`
}

// NetworksConfig represents all networks in the homelab configuration.
type NetworksConfig struct {
	BridgeModeNetworks    []BridgeModeNetworkConfig    `yaml:"bridgeModeNetworks"`
	ContainerModeNetworks []ContainerModeNetworkConfig `yaml:"containerModeNetworks"`
}

// BridgeModeNetworkConfig represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetworkConfig struct {
	Name              string              `yaml:"name"`
	HostInterfaceName string              `yaml:"hostInterfaceName"`
	CIDR              string              `yaml:"cidr"`
	Priority          int                 `yaml:"priority"`
	Containers        []ContainerIPConfig `yaml:"containers"`
}

// ContainerModeNetworkConfig represents a container network meant to attach a
// container to another container's network stack.
type ContainerModeNetworkConfig struct {
	Name       string               `yaml:"name"`
	Priority   int                  `yaml:"priority"`
	Containers []ContainerReference `yaml:"containers"`
}

// ContainerIP represents the IP information for a container.
type ContainerIPConfig struct {
	IP        string             `yaml:"ip"`
	Container ContainerReference `yaml:"container"`
}

// HostConfig represents the host specific information.
type HostConfig struct {
	Name              string               `yaml:"name"`
	AllowedContainers []ContainerReference `yaml:"allowedContainers"`
}

// ContainerReference identifies a specific container part of a group.
type ContainerReference struct {
	Group     string `yaml:"group"`
	Container string `yaml:"container"`
}

// ContainerGroupConfig represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroupConfig struct {
	Name  string `yaml:"name"`
	Order int    `yaml:"order"`
}

// ContainerConfig represents a single docker container.
type ContainerConfig struct {
	Info       ContainerReference        `yaml:"info"`
	Image      ContainerImageConfig      `yaml:"image"`
	Metadata   ContainerMetadataConfig   `yaml:"metadata"`
	Lifecycle  ContainerLifecycleConfig  `yaml:"lifecycle"`
	User       ContainerUserConfig       `yaml:"user"`
	Filesystem ContainerFilesystemConfig `yaml:"fs"`
	Network    ContainerNetworkConfig    `yaml:"network"`
	Security   ContainerSecurityConfig   `yaml:"security"`
	Runtime    ContainerRuntimeConfig    `yaml:"runtime"`
}

// ContainerImageConfig respresents the image configuration for the docker
// container.
type ContainerImageConfig struct {
	Image                   string `yaml:"image"`
	SkipImagePull           bool   `yaml:"skipImagePull"`
	IgnoreImagePullFailures bool   `yaml:"ignoreImagePullFailures"`
	PullImageBeforeStop     bool   `yaml:"pullImageBeforeStop"`
}

// ContainerMetadataConfig represents the metadata for the docker container.
type ContainerMetadataConfig struct {
	Labels []LabelConfig `yaml:"labels"`
}

// ContainerLifecycleConfig represents the lifecycle information for the
// docker container.
type ContainerLifecycleConfig struct {
	Order         int    `yaml:"order"`
	StartPreHook  string `yaml:"startPreHook"`
	RestartPolicy string `yaml:"restartPolicy"`
	AutoRemove    bool   `yaml:"autoRemove"`
	StopSignal    string `yaml:"stopSignal"`
	StopTimeout   int    `yaml:"stopTimeout"`
}

// ContainerUserConfig represents the user and group information for the
// docker container.
type ContainerUserConfig struct {
	User             string   `yaml:"user"`
	PrimaryGroup     string   `yaml:"primaryGroup"`
	AdditionalGroups []string `yaml:"additionalGroups"`
}

// ContainerFilesystemConfig represents the fileystem information for the
// docker container.
type ContainerFilesystemConfig struct {
	ReadOnlyRootfs bool           `yaml:"readOnlyRootfs"`
	Mounts         []MountConfig  `yaml:"mounts"`
	Devices        []DeviceConfig `yaml:"devices"`
}

// ContainerNetworkConfig represents the networking information for the
// docker container.
type ContainerNetworkConfig struct {
	HostName       string                `yaml:"hostName"`
	DomainName     string                `yaml:"domainName"`
	DNSServers     []string              `yaml:"dnsServers"`
	DNSOptions     []string              `yaml:"dnsOptions"`
	DNSSearch      []string              `yaml:"dnsSearch"`
	PublishedPorts []PublishedPortConfig `yaml:"publishedPorts"`
}

// ContainerSecurityConfig represents the security information for the
// docker container.
type ContainerSecurityConfig struct {
	Privileged bool           `yaml:"privileged"`
	Sysctls    []SysctlConfig `yaml:"sysctls"`
	CapAdd     []string       `yaml:"capAdd"`
	CapDrop    []string       `yaml:"capDrop"`
}

// ContainerRuntimeConfig represents the execution and runtime information
// for the docker container.
type ContainerRuntimeConfig struct {
	AttachToTty bool                 `yaml:"tty"`
	ShmSize     string               `yaml:"shmSize"`
	HealthCmd   string               `yaml:"healthCmd"`
	Env         []ContainerEnvConfig `yaml:"env"`
	Entrypoint  []string             `yaml:"entrypoint"`
	Args        []string             `yaml:"args"`
}

// MountConfig represents a filesystem mount.
type MountConfig struct {
	Type     string `yaml:"type"`
	Name     string `yaml:"name"`
	Src      string `yaml:"src"`
	Dst      string `yaml:"dst"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
	Options  string `yaml:"options"`
}

// DeviceConfig represents a device node that will be exposed to a container.
type DeviceConfig struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// SysctlConfig represents a sysctl config to apply to a container.
type SysctlConfig struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// ContainerEnvConfig represents an environment variable and value pair that will be set
// on the specified container.
type ContainerEnvConfig struct {
	Var          string `yaml:"var"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// PublishedPortConfig represents a port published from a container.
type PublishedPortConfig struct {
	ContainerPort int    `yaml:"containerPort"`
	Proto         string `yaml:"proto"`
	HostIP        string `yaml:"hostIp"`
	HostPort      int    `yaml:"hostPort"`
}

// LabelConfig represents a label set on a container.
type LabelConfig struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func mergedConfigReader(path string) (io.Reader, error) {
	var result []byte
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read contents of directory %s, reason: %w", path, err)
		} else if d == nil || d.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		log.Debugf("Picked up homelab config: %s", p)
		configFile, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read homelab config file %s, reason: %w", p, err)
		}
		result, err = deepmerge.YAML(result, configFile)
		if err != nil {
			return fmt.Errorf("failed to deep merge config file %s, reason: %w", p, err)
		}
		return nil
	})
	log.DebugEmpty()

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no homelab configs found in %s", path)
	}

	return bytes.NewReader(result), nil
}

func (h *HomelabConfig) parse(r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab config, reason: %w", err)
	}

	log.Tracef("Homelab Config:\n%v\n", prettyPrintJSON(h))
	return nil
}

func (h *HomelabConfig) parseConfigs(configsPath string) error {
	pathStat, err := os.Stat(configsPath)
	if err != nil {
		return fmt.Errorf("os.Stat() failed on homelab configs path, reason: %w", err)
	}
	if !pathStat.IsDir() {
		return fmt.Errorf("homelab configs path %s must be a directory", configsPath)
	}

	m, err := mergedConfigReader(configsPath)
	if err != nil {
		return err
	}

	return h.parse(m)
}

func (h *HomelabConfig) validate() error {
	err := validateGlobalConfig(&h.Global)
	if err != nil {
		return err
	}

	err = validateHostsConfig(h.Hosts)
	if err != nil {
		return err
	}

	err = validateIPAMConfig(&h.IPAM)
	if err != nil {
		return err
	}

	err = validateGroupsConfig(h.Groups)
	if err != nil {
		return err
	}

	err = validateContainersConfig(h.Containers)
	if err != nil {
		return err
	}

	return nil
}

func validateGlobalConfig(config *GlobalConfig) error {
	err := validateGlobalEnvConfig(config.Env)
	if err != nil {
		return err
	}

	err = validateMountDefs(config.MountDefs)
	if err != nil {
		return err
	}

	err = validateGlobalContainerConfig(&config.Container)
	if err != nil {
		return err
	}

	return nil
}

func validateGlobalEnvConfig(config []GlobalEnvConfig) error {
	// TODO: Perform the following (and more) validations:
	//     a. No duplicate global config env names.
	//     b. Validate mandatory properties of every global config env.
	//     c. Every global config env specifies exactly one of value or
	//        valueCommand, but not both.
	return nil
}

func validateMountDefs(config []MountConfig) error {
	// TODO: Perform the following (and more) validations:
	//     a. Validate mandatory properties of every global config mount.
	//     b. No duplicate global config mount names.
	return nil
}

func validateGlobalContainerConfig(config *GlobalContainerConfig) error {
	// TODO: Perform the following (and more) validations:
	return nil
}

func validateIPAMConfig(config *IPAMConfig) error {
	networks := make(map[string]bool)
	hostInterfaces := make(map[string]bool)
	bridgeModeNetworks := config.Networks.BridgeModeNetworks
	prefixes := make(map[netip.Prefix]string)
	for _, n := range bridgeModeNetworks {
		if len(n.Name) == 0 {
			return fmt.Errorf("network name cannot be empty")
		}
		if networks[n.Name] {
			return fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}

		if len(n.HostInterfaceName) == 0 {
			return fmt.Errorf("host interface name of network %s cannot be empty", n.Name)
		}
		if hostInterfaces[n.HostInterfaceName] {
			return fmt.Errorf("host interface name %s of network %s is already used by another network in the IPAM config", n.HostInterfaceName, n.Name)
		}
		if n.Priority <= 0 {
			return fmt.Errorf("network %s has a non-positive priority %d", n.Name, n.Priority)
		}

		networks[n.Name] = true
		hostInterfaces[n.HostInterfaceName] = true
		prefix, err := netip.ParsePrefix(n.CIDR)
		if err != nil {
			return fmt.Errorf("CIDR %s of network %s is invalid, reason: %w", n.CIDR, n.Name, err)
		}
		netAddr := prefix.Addr()
		if !netAddr.Is4() {
			return fmt.Errorf("CIDR %s of network %s is not an IPv4 subnet CIDR", n.CIDR, n.Name)
		}
		masked := prefix.Masked()
		if masked.Addr() != netAddr {
			return fmt.Errorf("CIDR %s of network %s is not the same as the network address %s", n.CIDR, n.Name, masked)
		}
		prefixLen := prefix.Bits()
		if prefixLen > 30 {
			return fmt.Errorf("CIDR %s of network %s (prefix length: %d) has a prefix length more than 30 which makes the network unusable for container IP address allocations", n.CIDR, n.Name, prefixLen)
		}
		if !netAddr.IsPrivate() {
			return fmt.Errorf("CIDR %s of network %s is not within the RFC1918 private address space", n.CIDR, n.Name)
		}
		for pre, preNet := range prefixes {
			if prefix.Overlaps(pre) {
				return fmt.Errorf("CIDR %s of network %s overlaps with CIDR %s of network %s", n.CIDR, n.Name, pre, preNet)
			}
		}
		prefixes[prefix] = n.Name

		gatewayAddr := netAddr.Next()
		containers := make(map[ContainerReference]bool)
		containerIPs := make(map[netip.Addr]bool)
		for _, cip := range n.Containers {
			ip := cip.IP
			ct := cip.Container
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}

			caddr, err := netip.ParseAddr(ip)
			if err != nil {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid IP %s, reason: %w", ct.Group, ct.Container, n.Name, ip, err)
			}
			if !prefix.Contains(caddr) {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has IP %s that does not belong to the network CIDR %s", ct.Group, ct.Container, n.Name, ip, prefix)
			}
			if caddr.Compare(netAddr) == 0 {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ip, netAddr)
			}
			if caddr.Compare(gatewayAddr) == 0 {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ip, gatewayAddr)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} has multiple endpoints in network %s", ct.Group, ct.Container, n.Name)
			}
			if containerIPs[caddr] {
				return fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ip, ct.Group, ct.Container, n.Name)
			}

			containers[ct] = true
			containerIPs[caddr] = true
		}
	}
	containerModeNetworks := config.Networks.ContainerModeNetworks
	for _, n := range containerModeNetworks {
		if len(n.Name) == 0 {
			return fmt.Errorf("network name cannot be empty")
		}
		if networks[n.Name] {
			return fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}
		if n.Priority <= 0 {
			return fmt.Errorf("network %s has a non-positive priority %d", n.Name, n.Priority)
		}
		networks[n.Name] = true

		containers := make(map[ContainerReference]bool)
		for _, ct := range n.Containers {
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} is connected to multiple container mode network stacks", ct.Group, ct.Container)
			}
			containers[ct] = true
		}
	}
	return nil
}

func validateHostsConfig(hosts []HostConfig) error {
	hostNames := make(map[string]bool)
	for _, h := range hosts {
		if len(h.Name) == 0 {
			return fmt.Errorf("host name cannot be empty in the hosts config")
		}
		if hostNames[h.Name] {
			return fmt.Errorf("host %s defined more than once in the hosts config", h.Name)
		}
		hostNames[h.Name] = true

		containers := make(map[ContainerReference]bool)
		for _, ct := range h.AllowedContainers {
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("allowed container config within host %s has invalid container reference, reason: %w", h.Name, err)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} defined more than once in the hosts config for host %s", ct.Group, ct.Container, h.Name)
			}
			containers[ct] = true
		}
	}
	return nil
}

func validateGroupsConfig(groups []ContainerGroupConfig) error {
	groupNames := make(map[string]bool)
	for _, g := range groups {
		if len(g.Name) == 0 {
			return fmt.Errorf("group name cannot be empty in the groups config")
		}
		if groupNames[g.Name] {
			return fmt.Errorf("group %s defined more than once in the groups config", g.Name)
		}
		if g.Order < 1 {
			return fmt.Errorf("group %s has a non-positive order %d", g.Name, g.Order)
		}

		groupNames[g.Name] = true
	}
	return nil
}

func validateContainersConfig(containers []ContainerConfig) error {
	// TODO: Perform the following (and more) validations:
	// Container configs:
	//     a. Parent group name is a valid group defined under group config.
	//     b. No duplicate container names within the same group.
	//     c. Order defined for all the containers.
	//     d. Image defined for all the containers.
	//     e. Validate mandatory properties of every device config.
	//     f. Validate manadatory properties of every container config mount.
	//     g. Mount pure name references are valid global config mount references.
	//     h. Validate manadatory properties of every container config env.
	//     i. Every container config env specifies exactly one of value or
	//        valueCommand, but not both.
	//     j. Validate mandatory properties of every published port config.
	//     k. Validate mandatory properties of every label config.
	return nil
}

func validateContainerReference(ref *ContainerReference) error {
	if len(ref.Group) == 0 {
		return fmt.Errorf("container reference cannot have an empty group name")
	}
	if len(ref.Container) == 0 {
		return fmt.Errorf("container reference cannot have an empty container name")
	}
	return nil
}
