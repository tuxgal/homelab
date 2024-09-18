package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/TwiN/deepmerge"
	"gopkg.in/yaml.v3"
)

// HomelabConfig represents the entire homelab deployment configuration.
type HomelabConfig struct {
	Global     GlobalConfig           `yaml:"global,omitempty"`
	IPAM       IPAMConfig             `yaml:"ipam,omitempty"`
	Hosts      []HostConfig           `yaml:"hosts,omitempty"`
	Groups     []ContainerGroupConfig `yaml:"groups,omitempty"`
	Containers []ContainerConfig      `yaml:"containers,omitempty"`
}

// GlobalConfig represents the configuration that will be applied
// across the entire homelab deployment.
type GlobalConfig struct {
	Env       []ConfigEnv           `yaml:"env,omitempty"`
	MountDefs []MountConfig         `yaml:"mountDefs,omitempty"`
	Container GlobalContainerConfig `yaml:"container,omitempty"`
}

// GlobalContainerConfig represents container related configuration that
// will be applied globally across all containers.
type GlobalContainerConfig struct {
	StopSignal    string                       `yaml:"stopSignal,omitempty"`
	StopTimeout   int                          `yaml:"stopTimeout,omitempty"`
	RestartPolicy ContainerRestartPolicyConfig `yaml:"restartPolicy,omitempty"`
	DomainName    string                       `yaml:"domainName,omitempty"`
	DNSSearch     []string                     `yaml:"dnsSearch,omitempty"`
	Env           []ContainerEnv               `yaml:"env,omitempty"`
	Mounts        []MountConfig                `yaml:"mounts,omitempty"`
	Labels        []LabelConfig                `yaml:"labels,omitempty"`
}

// ConfigEnv is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type ConfigEnv struct {
	Var          string `yaml:"var,omitempty"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// IPAMConfig represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAMConfig struct {
	Networks NetworksConfig `yaml:"networks,omitempty"`
}

// NetworksConfig represents all networks in the homelab configuration.
type NetworksConfig struct {
	BridgeModeNetworks    []BridgeModeNetworkConfig    `yaml:"bridgeModeNetworks,omitempty"`
	ContainerModeNetworks []ContainerModeNetworkConfig `yaml:"containerModeNetworks,omitempty"`
}

// BridgeModeNetworkConfig represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetworkConfig struct {
	Name              string              `yaml:"name,omitempty"`
	HostInterfaceName string              `yaml:"hostInterfaceName,omitempty"`
	CIDR              string              `yaml:"cidr,omitempty"`
	Priority          int                 `yaml:"priority,omitempty"`
	Containers        []ContainerIPConfig `yaml:"containers,omitempty"`
}

// ContainerModeNetworkConfig represents a container network meant to attach a
// container to another container's network stack.
type ContainerModeNetworkConfig struct {
	Name       string               `yaml:"name,omitempty"`
	Priority   int                  `yaml:"priority,omitempty"`
	Containers []ContainerReference `yaml:"containers,omitempty"`
}

// ContainerIP represents the IP information for a container.
type ContainerIPConfig struct {
	IP        string             `yaml:"ip,omitempty"`
	Container ContainerReference `yaml:"container,omitempty"`
}

// HostConfig represents the host specific information.
type HostConfig struct {
	Name              string               `yaml:"name,omitempty"`
	AllowedContainers []ContainerReference `yaml:"allowedContainers,omitempty"`
}

// ContainerReference identifies a specific container part of a group.
type ContainerReference struct {
	Group     string `yaml:"group,omitempty"`
	Container string `yaml:"container,omitempty"`
}

// ContainerGroupConfig represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroupConfig struct {
	Name  string `yaml:"name,omitempty"`
	Order int    `yaml:"order,omitempty"`
}

// ContainerConfig represents a single docker container.
type ContainerConfig struct {
	Info       ContainerReference        `yaml:"info,omitempty"`
	Config     ContainerConfigOptions    `yaml:"config,omitempty"`
	Image      ContainerImageConfig      `yaml:"image,omitempty"`
	Metadata   ContainerMetadataConfig   `yaml:"metadata,omitempty"`
	Lifecycle  ContainerLifecycleConfig  `yaml:"lifecycle,omitempty"`
	User       ContainerUserConfig       `yaml:"user,omitempty"`
	Filesystem ContainerFilesystemConfig `yaml:"fs,omitempty"`
	Network    ContainerNetworkConfig    `yaml:"network,omitempty"`
	Security   ContainerSecurityConfig   `yaml:"security,omitempty"`
	Health     ContainerHealthConfig     `yaml:"health,omitempty"`
	Runtime    ContainerRuntimeConfig    `yaml:"runtime,omitempty"`
}

// ContainerImageConfig respresents the image configuration for the docker
// container.
type ContainerImageConfig struct {
	Image                   string `yaml:"image,omitempty"`
	SkipImagePull           bool   `yaml:"skipImagePull,omitempty"`
	IgnoreImagePullFailures bool   `yaml:"ignoreImagePullFailures,omitempty"`
	PullImageBeforeStop     bool   `yaml:"pullImageBeforeStop,omitempty"`
}

// ContainerConfigOptions represents options that are applied while
// evaluating the config for this docker container.
type ContainerConfigOptions struct {
	Env []ConfigEnv `yaml:"env,omitempty"`
}

// ContainerMetadataConfig represents the metadata for the docker container.
type ContainerMetadataConfig struct {
	Labels []LabelConfig `yaml:"labels,omitempty"`
}

// ContainerLifecycleConfig represents the lifecycle information for the
// docker container.
type ContainerLifecycleConfig struct {
	Order         int                          `yaml:"order,omitempty"`
	StartPreHook  string                       `yaml:"startPreHook,omitempty"`
	RestartPolicy ContainerRestartPolicyConfig `yaml:"restartPolicy,omitempty"`
	AutoRemove    bool                         `yaml:"autoRemove,omitempty"`
	StopSignal    string                       `yaml:"stopSignal,omitempty"`
	StopTimeout   int                          `yaml:"stopTimeout,omitempty"`
}

// ContainerUserConfig represents the user and group information for the
// docker container.
type ContainerUserConfig struct {
	User             string   `yaml:"user,omitempty"`
	PrimaryGroup     string   `yaml:"primaryGroup,omitempty"`
	AdditionalGroups []string `yaml:"additionalGroups,omitempty"`
}

// ContainerFilesystemConfig represents the fileystem information for the
// docker container.
type ContainerFilesystemConfig struct {
	ReadOnlyRootfs bool           `yaml:"readOnlyRootfs,omitempty"`
	Mounts         []MountConfig  `yaml:"mounts,omitempty"`
	Devices        []DeviceConfig `yaml:"devices,omitempty"`
}

// ContainerNetworkConfig represents the networking information for the
// docker container.
type ContainerNetworkConfig struct {
	HostName       string                `yaml:"hostName,omitempty"`
	DomainName     string                `yaml:"domainName,omitempty"`
	DNSServers     []string              `yaml:"dnsServers,omitempty"`
	DNSOptions     []string              `yaml:"dnsOptions,omitempty"`
	DNSSearch      []string              `yaml:"dnsSearch,omitempty"`
	PublishedPorts []PublishedPortConfig `yaml:"publishedPorts,omitempty"`
}

// ContainerSecurityConfig represents the security information for the
// docker container.
type ContainerSecurityConfig struct {
	Privileged bool           `yaml:"privileged,omitempty"`
	Sysctls    []SysctlConfig `yaml:"sysctls,omitempty"`
	CapAdd     []string       `yaml:"capAdd,omitempty"`
	CapDrop    []string       `yaml:"capDrop,omitempty"`
}

// ContainerHealthConfig represents the health check options for the
// docker container.
type ContainerHealthConfig struct {
	Cmd           string `yaml:"cmd,omitempty"`
	Interval      string `yaml:"interval,omitempty"`
	Retries       int    `yaml:"retries,omitempty"`
	StartInterval string `yaml:"startInterval,omitempty"`
	StartPeriod   string `yaml:"startPeriod,omitempty"`
	Timeout       string `yaml:"timeout,omitempty"`
}

// ContainerRuntimeConfig represents the execution and runtime information
// for the docker container.
type ContainerRuntimeConfig struct {
	AttachToTty bool           `yaml:"tty,omitempty"`
	ShmSize     string         `yaml:"shmSize,omitempty"`
	Env         []ContainerEnv `yaml:"env,omitempty"`
	Entrypoint  []string       `yaml:"entrypoint,omitempty"`
	Args        []string       `yaml:"args,omitempty"`
}

// MountConfig represents a filesystem mount.
type MountConfig struct {
	Name     string `yaml:"name,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Src      string `yaml:"src,omitempty"`
	Dst      string `yaml:"dst,omitempty"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
	Options  string `yaml:"options,omitempty"`
}

// DeviceConfig represents a device node that will be exposed to a container.
type DeviceConfig struct {
	Src string `yaml:"src,omitempty"`
	Dst string `yaml:"dst,omitempty"`
}

// SysctlConfig represents a sysctl config to apply to a container.
type SysctlConfig struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// ContainerEnv represents an environment variable and value pair that will be set
// on the specified container.
type ContainerEnv struct {
	Var          string `yaml:"var,omitempty"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// PublishedPortConfig represents a port published from a container.
type PublishedPortConfig struct {
	ContainerPort int    `yaml:"containerPort,omitempty"`
	Proto         string `yaml:"proto,omitempty"`
	HostIP        string `yaml:"hostIp,omitempty"`
	HostPort      int    `yaml:"hostPort,omitempty"`
}

// LabelConfig represents a label set on a container.
type LabelConfig struct {
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type ContainerRestartPolicyConfig struct {
	Mode          string `yaml:"mode,omitempty"`
	MaxRetryCount int    `yaml:"maxRetryCount,omitempty"`
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
