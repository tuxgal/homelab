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
	Global GlobalConfig     `yaml:"global"`
	Ipam   IPAM             `yaml:"ipam"`
	Hosts  []Host           `yaml:"hosts"`
	Groups []ContainerGroup `yaml:"groups"`
}

// GlobalConfig represents the configuration that will be applied
// globally across all containers.
type GlobalConfig struct {
	Env                  []Env    `yaml:"env"`
	Volumes              []Volume `yaml:"volumes"`
	ContainerStopTimeout int      `yaml:"containerStopTimeout"`
	RestartPolicy        string   `yaml:"restartPolicy"`
	TimeZone             string   `yaml:"timeZone"`
	DNSSearch            string   `yaml:"dnsSearch"`
}

// Env is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type Env struct {
	Var   string `yaml:"var"`
	Value string `yaml:"value"`
}

// IPAM represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAM struct {
	Networks    Networks      `yaml:"networks"`
	PrimaryIP   []ContainerIP `yaml:"primaryIp"`
	SecondaryIP []SecondaryIP `yaml:"secondaryIp"`
}

// Networks represents all networks in the homelab configuration.
type Networks struct {
	BridgeModeNetworks []BridgeModeNetwork `yaml:"standardNetworks"`
	ContainerNetworks  []ContainerNetwork  `yaml:"containerNetworks"`
}

// BridgeModeNetwork represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetwork struct {
	Name              string `yaml:"name"`
	HostInterfaceName string `yaml:"hostInterfaceName"`
	Cidr              string `yaml:"cidr"`
}

// ContainerNetwork represents a container network meant to attach a
// container to another container's network stack.
type ContainerNetwork struct {
	Name       string               `yaml:"name"`
	Containers []ContainerWithGroup `yaml:"containers"`
}

// ContainerIP represents the IP information for a container.
type ContainerIP struct {
	IP        string `yaml:"ip"`
	Container string `yaml:"container"`
	Group     string `yaml:"group"`
}

// SecondaryIP represents the secondary IP information for all
// containers in the homelab configuration.
type SecondaryIP struct {
	Network string        `yaml:"network"`
	Ips     []ContainerIP `yaml:"ips"`
}

// Host represents the host specific information.
type Host struct {
	Name              string               `yaml:"name"`
	AllowedContainers []ContainerWithGroup `yaml:"allowedContainers"`
}

// ContainerWithGroup identified a specific container part of a group.
type ContainerWithGroup struct {
	Group     string `yaml:"group"`
	Container string `yaml:"container"`
}

// ContainerGroup represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroup struct {
	Name           string      `yaml:"name"`
	Order          int         `yaml:"order"`
	PrimaryNetwork string      `yaml:"primaryNetwork"`
	Containers     []Container `yaml:"containers"`
}

// Container represents a single docker container.
type Container struct {
	Name                    string          `yaml:"name"`
	Image                   string          `yaml:"image"`
	Order                   int             `yaml:"order"`
	ContainerStopTimeout    int             `yaml:"containerStopTimeout"`
	SkipImagePull           bool            `yaml:"skipImagePull"`
	IgnoreImagePullFailures bool            `yaml:"ignoreImagePullFailures"`
	PullImageAfterStop      bool            `yaml:"pullImageAfterStop"`
	StartPreHook            string          `yaml:"startPreHook"`
	User                    string          `yaml:"user"`
	Group                   string          `yaml:"group"`
	GroupAdd                []string        `yaml:"groupAdd"`
	HostName                string          `yaml:"hostName"`
	DomainName              string          `yaml:"domainName"`
	DynamicFlagsCommand     string          `yaml:"dynamicFlagsCommand"`
	Devices                 []Device        `yaml:"devices"`
	ShmSize                 string          `yaml:"shmSize"`
	CapAdd                  []string        `yaml:"capAdd"`
	Mounts                  []string        `yaml:"mounts"`
	Volumes                 []Volume        `yaml:"volumes"`
	Env                     []ContainerEnv  `yaml:"env"`
	PublishedPorts          []PublishedPort `yaml:"publishedPorts"`
	Labels                  []Label         `yaml:"labels"`
	StopSignal              string          `yaml:"stopSignal"`
	HealthCmd               string          `yaml:"healthCmd"`
	Entrypoint              string          `yaml:"entrypoint"`
	Args                    []string        `yaml:"args"`
}

// Volume represents a bind mounted volume.
type Volume struct {
	Name     string `yaml:"name"`
	Src      string `yaml:"src"`
	Dst      string `yaml:"dst"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
}

// Device represents a device node that will be exposed to a container.
type Device struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// ContainerEnv represents an environment variable and value pair that will be set
// on the specified container.
type ContainerEnv struct {
	Var          string `yaml:"var"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// PublishedPort represents a port published from a container.
type PublishedPort struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// Label represents a label set on a container.
type Label struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func homelabConfigsPath() (string, error) {
	// Use the flag from the command line if present.
	if isFlagPassed(homelabConfigsDirFlag) {
		log.Infof("Using Homelab configs path from command line flag: %s", *homelabConfigsDir)
		return *homelabConfigsDir, nil
	}
	path, err := configsPath()
	if err != nil {
		return "", err
	}

	log.Infof("Using Homelab configs path from CLI config: %s", path)
	return path, nil
}

func mergedConfigReader(path string) (io.Reader, error) {
	var result []byte
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d == nil || d.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		log.Infof("Picked up homelab config: %s", p)
		configFile, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read homelab config file %q, reason: %w", p, err)
		}
		result, err = deepmerge.YAML(result, configFile)
		if err != nil {
			return fmt.Errorf("failed to deep merge config file %q, reason: %w", p, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no homelab configs found in %s", path)
	}

	return bytes.NewReader(result), nil
}

func parseHomelabConfig() (*HomelabConfig, error) {
	path, err := homelabConfigsPath()
	if err != nil {
		return nil, err
	}

	pathStat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("os.Stat() failed on homelab configs path, reason: %w", err)
	}
	if !pathStat.IsDir() {
		return nil, fmt.Errorf("homelab configs path %q must be a directory", path)
	}

	configFile, err := mergedConfigReader(path)
	if err != nil {
		return nil, err
	}

	var config HomelabConfig
	dec := yaml.NewDecoder(configFile)
	dec.KnownFields(true)
	err = dec.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse homelab config, reason: %w", err)
	}

	log.Infof("Homelab Config:\n%v\n", prettyPrintJSON(config))
	return &config, nil
}
