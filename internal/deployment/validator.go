package deployment

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/docker/go-units"
	"github.com/tuxgal/homelab/internal/cmdexec"
	"github.com/tuxgal/homelab/internal/config"
	"github.com/tuxgal/homelab/internal/config/env"
	"github.com/tuxgal/homelab/internal/docker"
	"github.com/tuxgal/homelab/internal/host"
	"github.com/tuxgal/homelab/internal/utils"
)

const (
	reservedULAAddr     = "fc00::"
	reservedULAAddrBits = 8
)

var (
	reservedULAPrefix = netip.PrefixFrom(netip.MustParseAddr(reservedULAAddr), reservedULAAddrBits)
)

func validateGlobalConfig(ctx context.Context, parentEnv *env.ConfigEnvManager, conf *config.Global) (*env.ConfigEnvManager, error) {
	if err := validateBaseDir(conf.BaseDir); err != nil {
		return nil, err
	}

	newEnvMap, newEnvOrder, err := validateConfigEnv(conf.Env, "global config")
	if err != nil {
		return nil, err
	}

	// Apply the config env prior to validating other info within the global config.
	env := parentEnv.NewGlobalConfigEnvManager(ctx, conf.BaseDir, newEnvMap, newEnvOrder)
	conf.ApplyConfigEnv(env)

	if err := validateMountsConfig(conf.MountDefs, nil, nil, "global config mount defs"); err != nil {
		return nil, err
	}

	if err := validateGlobalContainerConfig(&conf.Container, conf.MountDefs); err != nil {
		return nil, err
	}

	return env, nil
}

func validateBaseDir(baseDir string) error {
	if len(baseDir) == 0 {
		return fmt.Errorf("homelab base directory cannot be empty")
	}
	pathStat, err := os.Stat(baseDir)
	if err != nil {
		return fmt.Errorf("os.Stat() failed on homelab base directory path, reason: %w", err)
	}
	if !pathStat.IsDir() {
		return fmt.Errorf("homelab base directory path %s must be a directory", baseDir)
	}
	return nil
}

func validateConfigEnv(conf []config.ConfigEnv, location string) (env.EnvMap, env.EnvOrder, error) {
	envs := env.EnvMap{}
	envOrder := env.EnvOrder{}
	for _, e := range conf {
		if len(e.Var) == 0 {
			return nil, nil, fmt.Errorf("empty env var in %s", location)
		}
		if _, found := envs[e.Var]; found {
			return nil, nil, fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}

		if len(e.Value) == 0 && len(e.ValueCommand) == 0 {
			return nil, nil, fmt.Errorf("neither value nor valueCommand specified for env var %s in %s", e.Var, location)
		}
		if len(e.Value) > 0 && len(e.ValueCommand) > 0 {
			return nil, nil, fmt.Errorf("exactly one of value or valueCommand must be specified for env var %s in %s", e.Var, location)
		}

		if len(e.Value) > 0 {
			envs[e.Var] = e.Value
		} else {
			// TODO: Evaluate the value by running the command using the executor.
			envs[e.Var] = "TODO" // e.ValueCommand
		}
		envOrder = append(envOrder, e.Var)
	}
	return envs, envOrder, nil
}

func validateContainerEnv(conf []config.ContainerEnv, location string) error {
	envs := utils.StringSet{}
	for _, e := range conf {
		if len(e.Var) == 0 {
			return fmt.Errorf("empty env var in %s", location)
		}
		if _, found := envs[e.Var]; found {
			return fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}
		envs[e.Var] = struct{}{}

		if len(e.Value) == 0 {
			return fmt.Errorf("value not specified for env var %s in %s", e.Var, location)
		}
	}
	return nil
}

func validateLabelsConfig(conf []config.Label, location string) error {
	labels := utils.StringSet{}
	for _, l := range conf {
		if len(l.Name) == 0 {
			return fmt.Errorf("empty label name in %s", location)
		}
		if _, found := labels[l.Name]; found {
			return fmt.Errorf("label name %s specified more than once in %s", l.Name, location)
		}
		labels[l.Name] = struct{}{}

		if len(l.Value) == 0 {
			return fmt.Errorf("empty label value for label %s in %s", l.Name, location)
		}
	}
	return nil
}

func validateMountsConfig(conf, commonConfig, globalDefs []config.Mount, location string) error {
	// First build a map of the mounts from the globalDefs (which should
	// already have been validated).
	globalMountDefs := utils.StringSet{}
	for _, m := range globalDefs {
		globalMountDefs[m.Name] = struct{}{}
	}

	// Build a map of the mounts from the commonConfig next which acts
	// as the first set of mounts to apply. These should also have been
	// validated prior and hence we don't validate them here again.
	mounts := utils.StringSet{}
	for _, m := range commonConfig {
		mounts[m.Name] = struct{}{}
	}

	// Finally iterate and validate the mounts in the current level config.
	for _, m := range conf {
		if len(m.Name) == 0 {
			return fmt.Errorf("mount name cannot be empty in %s", location)
		}
		if _, found := mounts[m.Name]; found {
			return fmt.Errorf("mount name %s defined more than once in %s", m.Name, location)
		}
		mounts[m.Name] = struct{}{}

		if len(m.Type) == 0 && len(m.Src) == 0 && len(m.Dst) == 0 && !m.ReadOnly {
			// This is a mount with just the name. Match this against the
			// global mount defs.
			if _, found := globalMountDefs[m.Name]; !found {
				return fmt.Errorf("mount specified by just the name %s not found in defs in %s", m.Name, location)
			}
			// No further validation needed for a mount referencing a def.
			continue
		}

		if m.Type != "bind" && m.Type != "tmpfs" {
			return fmt.Errorf("unsupported mount type %s for mount %s in %s", m.Type, m.Name, location)
		}
		if m.Type == "bind" && len(m.Src) == 0 {
			return fmt.Errorf("bind mount name %s cannot have an empty value for src in %s", m.Name, location)
		}
		if m.Type == "tmpfs" && len(m.Src) != 0 {
			return fmt.Errorf("tmpfs mount name %s cannot have a non-empty value for src in %s", m.Name, location)
		}
		if len(m.Dst) == 0 {
			return fmt.Errorf("mount name %s cannot have an empty value for dst in %s", m.Name, location)
		}
		if m.Type == "bind" && m.TmpfsSize != 0 {
			return fmt.Errorf("bind mount name %s cannot specify tmpfs size in %s", m.Name, location)
		}
		if m.Type == "tmpfs" && m.TmpfsSize < 0 {
			return fmt.Errorf("tmpfs mount name %s cannot specify a negative tmpfs size %d in %s", m.Name, m.TmpfsSize, location)
		}
	}
	return nil
}

func validateDevicesConfig(devices []config.Device, location string) error {
	for _, d := range devices {
		if len(d.Src) == 0 {
			return fmt.Errorf("device src cannot be empty in %s", location)
		}
	}
	return nil
}

func validatePublishedPortsConfig(ports []config.PublishedPort, location string) error {
	for _, p := range ports {
		ctPort, err := strconv.ParseInt(p.ContainerPort, 10, 32)
		if err != nil {
			return fmt.Errorf("unable to convert published container port %s to an integer, reason: %w", p.ContainerPort, err)
		}
		if ctPort <= 0 {
			return fmt.Errorf("published container port %d cannot be non-positive in %s", ctPort, location)
		}
		if p.Protocol != "tcp" && p.Protocol != "udp" {
			return fmt.Errorf("published container port %s specifies an invalid protocol %s in %s", p.ContainerPort, p.Protocol, location)
		}
		if len(p.HostIP) == 0 {
			return fmt.Errorf("published host IP cannot be empty for container port %s in %s", p.ContainerPort, location)
		}
		if _, err := netip.ParseAddr(p.HostIP); err != nil {
			return fmt.Errorf("published host IP %s for container port %s is invalid in %s, reason: %w", p.HostIP, p.ContainerPort, location, err)
		}
		hostPort, err := strconv.ParseInt(p.HostPort, 10, 32)
		if err != nil {
			return fmt.Errorf("unable to convert published host port %s to an integer, reason: %w", p.HostPort, err)
		}
		if hostPort <= 0 {
			return fmt.Errorf("published host port %d cannot be non-positive in %s", hostPort, location)
		}
	}
	return nil
}

func validateSysctlsConfig(sysctls []config.Sysctl, location string) error {
	keys := utils.StringSet{}
	for _, s := range sysctls {
		if len(s.Key) == 0 {
			return fmt.Errorf("empty sysctl key in %s", location)
		}
		if _, found := keys[s.Key]; found {
			return fmt.Errorf("sysctl key %s specified more than once in %s", s.Key, location)
		}
		keys[s.Key] = struct{}{}

		if len(s.Value) == 0 {
			return fmt.Errorf("empty sysctl value for sysctl %s in %s", s.Key, location)
		}
	}
	return nil
}

func validateHealthConfig(conf *config.ContainerHealth, location string) error {
	if conf.Retries < 0 {
		return fmt.Errorf("health check retries %d cannot be negative in %s", conf.Retries, location)
	}
	if len(conf.Interval) > 0 {
		if _, err := time.ParseDuration(conf.Interval); err != nil {
			return fmt.Errorf("health check interval %s is invalid in %s, reason: %w", conf.Interval, location, err)
		}
	}
	if len(conf.Timeout) > 0 {
		if _, err := time.ParseDuration(conf.Timeout); err != nil {
			return fmt.Errorf("health check timeout %s is invalid in %s, reason: %w", conf.Timeout, location, err)
		}
	}
	if len(conf.StartPeriod) > 0 {
		if _, err := time.ParseDuration(conf.StartPeriod); err != nil {
			return fmt.Errorf("health check start period %s is invalid in %s, reason: %w", conf.StartPeriod, location, err)
		}
	}
	if len(conf.StartInterval) > 0 {
		if _, err := time.ParseDuration(conf.StartInterval); err != nil {
			return fmt.Errorf("health check start interval %s is invalid in %s, reason: %w", conf.StartInterval, location, err)
		}
	}
	return nil
}

func validateGlobalContainerConfig(conf *config.GlobalContainer, globalMountDefs []config.Mount) error {
	if conf.StopTimeout < 0 {
		return fmt.Errorf("container stop timeout %d cannot be negative in global container config", conf.StopTimeout)
	}
	if err := validateContainerRestartPolicy(&conf.RestartPolicy, "global container config"); err != nil {
		return err
	}
	if err := validateContainerEnv(conf.Env, "global container config"); err != nil {
		return err
	}
	if err := validateMountsConfig(conf.Mounts, nil, globalMountDefs, "global container config mounts"); err != nil {
		return err
	}
	if err := validateLabelsConfig(conf.Labels, "global container config"); err != nil {
		return err
	}
	return nil
}

func validateContainerRestartPolicy(conf *config.ContainerRestartPolicy, location string) error {
	if conf.Mode != "on-failure" && conf.MaxRetryCount != 0 {
		return fmt.Errorf("restart policy max retry count can be set only when the mode is on-failure in %s", location)
	}
	if len(conf.Mode) == 0 {
		return nil
	}
	if _, err := docker.RestartPolicyModeFromString(conf.Mode); err != nil {
		return fmt.Errorf("invalid restart policy mode %s in %s, valid values are %s", conf.Mode, location, docker.RestartPolicyModeValidValues())
	}
	if conf.MaxRetryCount < 0 {
		return fmt.Errorf("restart policy max retry count %d cannot be negative in %s", conf.MaxRetryCount, location)
	}
	return nil
}

func validateIPAMConfig(ctx context.Context, conf *config.IPAM) (NetworkMap, map[config.ContainerReference]networkEndpointList, error) {
	networks := NetworkMap{}
	hostInterfaces := utils.StringSet{}
	bridgeModeNetworks := conf.Networks.BridgeModeNetworks
	v4Prefixes := make(map[netip.Prefix]string)
	v6Prefixes := make(map[netip.Prefix]string)
	containerEndpoints := make(map[config.ContainerReference]networkEndpointList)
	allBridgeModeContainers := make(map[config.ContainerReference]struct{})

	for _, n := range bridgeModeNetworks {
		if len(n.Name) == 0 {
			return nil, nil, fmt.Errorf("network name cannot be empty")
		}
		if _, found := networks[n.Name]; found {
			return nil, nil, fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}

		if len(n.HostInterfaceName) == 0 {
			return nil, nil, fmt.Errorf("host interface name of network %s cannot be empty", n.Name)
		}
		if _, found := hostInterfaces[n.HostInterfaceName]; found {
			return nil, nil, fmt.Errorf("host interface name %s of network %s is already used by another network in the IPAM config", n.HostInterfaceName, n.Name)
		}
		if n.Priority <= 0 {
			return nil, nil, fmt.Errorf("network %s cannot have a non-positive priority %d", n.Name, n.Priority)
		}

		hostInterfaces[n.HostInterfaceName] = struct{}{}
		v4Prefix, err := netip.ParsePrefix(n.CIDR.V4)
		if err != nil {
			return nil, nil, fmt.Errorf("v4 CIDR %s of network %s is invalid, reason: %w", n.CIDR.V4, n.Name, err)
		}
		v4NetAddr := v4Prefix.Addr()
		if !v4NetAddr.Is4() {
			return nil, nil, fmt.Errorf("v4 CIDR %s of network %s is not an IPv4 subnet CIDR", n.CIDR.V4, n.Name)
		}
		if masked := v4Prefix.Masked(); masked.Addr() != v4NetAddr {
			return nil, nil, fmt.Errorf("v4 CIDR %s of network %s is not the same as the network address %s", n.CIDR.V4, n.Name, masked)
		}
		if prefixLen := v4Prefix.Bits(); prefixLen > 30 {
			return nil, nil, fmt.Errorf("v4 CIDR %s of network %s (prefix length: %d) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations", n.CIDR.V4, n.Name, prefixLen)
		}
		if !v4NetAddr.IsPrivate() {
			return nil, nil, fmt.Errorf("v4 CIDR %s of network %s is not within the RFC1918 private address space", n.CIDR.V4, n.Name)
		}
		for pre, preNet := range v4Prefixes {
			if v4Prefix.Overlaps(pre) {
				return nil, nil, fmt.Errorf("v4 CIDR %s of network %s overlaps with v4 CIDR %s of network %s", n.CIDR.V4, n.Name, pre, preNet)
			}
		}
		v4Prefixes[v4Prefix] = n.Name
		v4GatewayAddr := v4NetAddr.Next()

		var v6Prefix netip.Prefix
		var v6GatewayAddr netip.Addr
		var v6NetAddr netip.Addr
		if n.CIDR.V6 != "" {
			var err error
			v6Prefix, err = netip.ParsePrefix(n.CIDR.V6)
			if err != nil {
				return nil, nil, fmt.Errorf("v6 CIDR %s of network %s is invalid, reason: %w", n.CIDR.V6, n.Name, err)
			}
			v6NetAddr = v6Prefix.Addr()
			if !v6NetAddr.Is6() {
				return nil, nil, fmt.Errorf("v6 CIDR %s of network %s is not an IPv6 subnet CIDR", n.CIDR.V6, n.Name)
			}
			if masked := v6Prefix.Masked(); masked.Addr() != v6NetAddr {
				return nil, nil, fmt.Errorf("v6 CIDR %s of network %s is not the same as the network address %s", n.CIDR.V6, n.Name, masked)
			}
			if prefixLen := v6Prefix.Bits(); prefixLen != 64 {
				return nil, nil, fmt.Errorf("v6 CIDR %s of network %s (prefix length: %d) must have a prefix length 64 as per the convention for IPv6 networks", n.CIDR.V6, n.Name, prefixLen)
			}
			if !v6NetAddr.IsPrivate() {
				return nil, nil, fmt.Errorf("v6 CIDR %s of network %s is not within the ULA private address space", n.CIDR.V6, n.Name)
			}
			if v6Prefix.Overlaps(reservedULAPrefix) {
				return nil, nil, fmt.Errorf("v6 CIDR %s of network %s overlaps with the reserved ULA prefix %s", n.CIDR.V6, n.Name, reservedULAPrefix)
			}
			for pre, preNet := range v6Prefixes {
				if v6Prefix.Overlaps(pre) {
					return nil, nil, fmt.Errorf("v6 CIDR %s of network %s overlaps with v6 CIDR %s of network %s", n.CIDR.V6, n.Name, pre, preNet)
				}
			}
			v6Prefixes[v6Prefix] = n.Name
			v6GatewayAddr = v6NetAddr.Next()
		}

		bmn := newBridgeModeNetwork(n.Name, n.Priority, &bridgeModeNetworkInfo{
			priority:          n.Priority,
			hostInterfaceName: n.HostInterfaceName,
			v4CIDR:            v4Prefix,
			v4Gateway:         v4GatewayAddr,
			enableV6:          n.CIDR.V6 != "",
			v6CIDR:            v6Prefix,
			v6Gateway:         v6GatewayAddr,
		})
		networks[n.Name] = bmn

		containers := make(map[config.ContainerReference]struct{})
		containerIPs := make(map[netip.Addr]struct{})
		for _, cip := range n.Containers {
			ct := cip.Container
			if err := validateContainerReference(&ct); err != nil {
				return nil, nil, fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}

			ipv4 := cip.IP.IPv4
			caddrv4, err := netip.ParseAddr(ipv4)
			if err != nil {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid v4 IP %s, reason: %w", ct.Group, ct.Container, n.Name, ipv4, err)
			}
			if !v4Prefix.Contains(caddrv4) {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have a v4 IP %s that does not belong to the network v4 CIDR %s", ct.Group, ct.Container, n.Name, ipv4, v4Prefix)
			}
			if caddrv4.Compare(v4NetAddr) == 0 {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ipv4, v4NetAddr)
			}
			if caddrv4.Compare(v4GatewayAddr) == 0 {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ipv4, v4GatewayAddr)
			}
			if _, found := containerIPs[caddrv4]; found {
				return nil, nil, fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ipv4, ct.Group, ct.Container, n.Name)
			}
			containerIPs[caddrv4] = struct{}{}

			ipv6 := cip.IP.IPv6
			if ipv6 != "" {
				caddrv6, err := netip.ParseAddr(ipv6)
				if err != nil {
					return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid v6 IP %s, reason: %w", ct.Group, ct.Container, n.Name, ipv6, err)
				}
				if n.CIDR.V6 == "" {
					return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s specified a v6 IP address %s when the network has no v6 subnet CIDRs defined", ct.Group, ct.Container, n.Name, ipv6)
				}
				if !v6Prefix.Contains(caddrv6) {
					return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have a v6 IP %s that does not belong to the network v6 CIDR %s", ct.Group, ct.Container, n.Name, ipv6, v6Prefix)
				}
				if caddrv6.Compare(v6NetAddr) == 0 {
					return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ipv6, v6NetAddr)
				}
				if caddrv6.Compare(v6GatewayAddr) == 0 {
					return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ipv6, v6GatewayAddr)
				}
				if _, found := containerIPs[caddrv6]; found {
					return nil, nil, fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ipv6, ct.Group, ct.Container, n.Name)
				}
				containerIPs[caddrv6] = struct{}{}
			}

			if _, found := containers[ct]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} cannot have multiple endpoints in network %s", ct.Group, ct.Container, n.Name)
			}
			containers[ct] = struct{}{}
			allBridgeModeContainers[ct] = struct{}{}
			containerEndpoints[ct] = append(containerEndpoints[ct], newBridgeModeEndpoint(bmn, ipv4, ipv6))
		}
	}

	containerModeNetworks := conf.Networks.ContainerModeNetworks
	allContainerModeContainers := make(map[config.ContainerReference]struct{})
	for _, n := range containerModeNetworks {
		if len(n.Name) == 0 {
			return nil, nil, fmt.Errorf("network name cannot be empty")
		}
		if _, found := networks[n.Name]; found {
			return nil, nil, fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}
		if err := validateContainerReference(&n.Container); err != nil {
			return nil, nil, fmt.Errorf("container reference of container mode network %s is invalid, reason: %w", n.Name, err)
		}
		cmn := newContainerModeNetwork(n.Name, &containerModeNetworkInfo{
			container: n.Container,
		})
		networks[n.Name] = cmn

		for _, ct := range n.AttachingContainers {
			if err := validateContainerReference(&ct); err != nil {
				return nil, nil, fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}
			if _, found := allContainerModeContainers[ct]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} is connected to multiple container mode network stacks", ct.Group, ct.Container)
			}
			if _, found := allBridgeModeContainers[ct]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} is connected to both bridge mode and container mode network stacks", ct.Group, ct.Container)
			}
			allContainerModeContainers[ct] = struct{}{}
			containerEndpoints[ct] = append(containerEndpoints[ct], newContainerModeEndpoint(cmn))
		}
	}

	for ct, endpoints := range containerEndpoints {
		if len(endpoints) <= 1 {
			continue
		}

		priorities := make(map[int]struct{})
		for _, e := range endpoints {
			p := e.network.bridgeModeInfo.priority
			if _, found := priorities[p]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} cannot have multiple bridge mode network endpoints whose networks have the same priority %d", ct.Group, ct.Container, p)
			}
			priorities[p] = struct{}{}
		}

		// Sort the networks for each container by priority (i.e. lowest
		// priority is the primary network interface for the container).
		sort.Slice(endpoints, func(i, j int) bool {
			// These networks are all guaranteed to be bridge mode networks
			// as we have already validated that a given container connects
			// to at most one container mode network and doesn't connect
			// to both bridge and container mode networks at the same time.
			n1 := endpoints[i].network
			n2 := endpoints[j].network

			return n1.bridgeModeInfo.priority < n2.bridgeModeInfo.priority
		})
	}

	return networks, containerEndpoints, nil
}

func validateHostsConfig(ctx context.Context, hosts []config.Host) (containerSet, error) {
	currentHost := host.MustHostInfo(ctx)
	hostNames := utils.StringSet{}
	allowedContainers := containerSet{}
	for _, h := range hosts {
		if len(h.Name) == 0 {
			return nil, fmt.Errorf("host name cannot be empty in the hosts config")
		}
		if _, found := hostNames[h.Name]; found {
			return nil, fmt.Errorf("host %s defined more than once in the hosts config", h.Name)
		}
		hostNames[h.Name] = struct{}{}

		containers := make(map[config.ContainerReference]bool)
		for _, ct := range h.AllowedContainers {
			err := validateContainerReference(&ct)
			if err != nil {
				return nil, fmt.Errorf("allowed container config within host %s has invalid container reference, reason: %w", h.Name, err)
			}
			if containers[ct] {
				return nil, fmt.Errorf("container {Group:%s Container:%s} defined more than once in the hosts config for host %s", ct.Group, ct.Container, h.Name)
			}
			containers[ct] = true
			if h.Name == currentHost.HostName {
				allowedContainers[ct] = true
			}
		}
	}
	return allowedContainers, nil
}

func validateGroupsConfig(groups []config.ContainerGroup) (ContainerGroupMap, error) {
	containerGroups := ContainerGroupMap{}
	for _, g := range groups {
		if len(g.Name) == 0 {
			return nil, fmt.Errorf("group name cannot be empty in the groups config")
		}
		if _, found := containerGroups[g.Name]; found {
			return nil, fmt.Errorf("group %s defined more than once in the groups config", g.Name)
		}
		if g.Order < 1 {
			return nil, fmt.Errorf("group %s cannot have a non-positive order %d", g.Name, g.Order)
		}

		containerGroups[g.Name] = NewContainerGroup(&g)
	}
	return containerGroups, nil
}

func validateContainersConfig(ctx context.Context, parentEnv *env.ConfigEnvManager, containersConfig []config.Container, groups ContainerGroupMap, globalConfig *config.Global, containerEndpoints map[config.ContainerReference]networkEndpointList, allowedContainers containerSet) error {
	exec := cmdexec.MustExecutor(ctx)
	for i, ct := range containersConfig {
		g, found := groups[ct.Info.Group]
		if !found {
			return fmt.Errorf("group definition missing in groups config for the container {Group:%s Container:%s} in the containers config", ct.Info.Group, ct.Info.Container)
		}
		if _, found := g.containers[ct.Info]; found {
			return fmt.Errorf("container {Group:%s Container:%s} defined more than once in the containers config", ct.Info.Group, ct.Info.Container)
		}

		loc := fmt.Sprintf("container {Group: %s Container:%s} config", ct.Info.Group, ct.Info.Container)
		ctConfigEnvMap, ctConfigEnvOrder, err := validateConfigEnv(ct.Config.Env, loc)
		if err != nil {
			return err
		}
		ctEnv := parentEnv.NewContainerConfigEnvManager(ctx, containerGroupBaseDir(globalConfig.BaseDir, ct.Info), containerBaseDir(globalConfig.BaseDir, ct.Info), ctConfigEnvMap, ctConfigEnvOrder)
		ct.ApplyConfigEnv(ctEnv)
		if err := ct.ApplyCmdExecutor(exec); err != nil {
			return err
		}

		if len(ct.Image.Image) == 0 {
			return fmt.Errorf("image cannot be empty in %s", loc)
		}
		if ct.Image.SkipImagePull {
			if ct.Image.IgnoreImagePullFailures {
				return fmt.Errorf("ignoreImagePullFailures cannot be true when skipImagePull is true in %s", loc)
			}
			if ct.Image.PullImageBeforeStop {
				return fmt.Errorf("pullImageBeforeStop cannot be true when skipImagePull is true in %s", loc)
			}
		}

		if err := validateLabelsConfig(ct.Metadata.Labels, loc); err != nil {
			return err
		}

		if ct.Lifecycle.Order <= 0 {
			return fmt.Errorf("container order %d cannot be non-positive in %s", ct.Lifecycle.Order, loc)
		}
		if err := validateContainerRestartPolicy(&ct.Lifecycle.RestartPolicy, loc); err != nil {
			return err
		}
		if ct.Lifecycle.StopTimeout < 0 {
			return fmt.Errorf("container stop timeout %d cannot be negative in %s", ct.Lifecycle.StopTimeout, loc)
		}
		if ct.Lifecycle.WaitAfterStartDelay < 0 {
			return fmt.Errorf("container wait after start delay %d cannot be negative in %s", ct.Lifecycle.WaitAfterStartDelay, loc)
		}

		if len(ct.User.PrimaryGroup) > 0 && len(ct.User.User) == 0 {
			return fmt.Errorf("container user primary group cannot be set without setting the user in %s", loc)
		}

		if err := validateDevicesConfig(ct.Filesystem.Devices.Static, loc); err != nil {
			return err
		}

		if err := validateMountsConfig(ct.Filesystem.Mounts, globalConfig.Container.Mounts, globalConfig.MountDefs, fmt.Sprintf("%s mounts", loc)); err != nil {
			return err
		}

		if err := validatePublishedPortsConfig(ct.Network.PublishedPorts, loc); err != nil {
			return err
		}

		if err := validateSysctlsConfig(ct.Security.Sysctls, loc); err != nil {
			return err
		}

		if err := validateHealthConfig(&ct.Health, loc); err != nil {
			return err
		}

		if len(ct.Runtime.ShmSize) > 0 {
			if _, err := units.RAMInBytes(ct.Runtime.ShmSize); err != nil {
				return fmt.Errorf("invalid shmSize %s in %s, reason: %w", ct.Runtime.ShmSize, loc, err)
			}
		}
		if err := validateContainerEnv(ct.Runtime.Env, loc); err != nil {
			return err
		}

		g.addContainer(&ct, globalConfig, containerEndpoints[ct.Info], allowedContainers[ct.Info])
		// This is needed to store the updated container config after
		// ApplyConfigEnv().
		containersConfig[i] = ct
	}

	return nil
}

func validateContainerReference(ref *config.ContainerReference) error {
	if len(ref.Group) == 0 {
		return fmt.Errorf("container reference cannot have an empty group name")
	}
	if len(ref.Container) == 0 {
		return fmt.Errorf("container reference cannot have an empty container name")
	}
	return nil
}

func newBridgeModeEndpoint(network *Network, ipv4 string, ipv6 string) *containerNetworkEndpoint {
	return &containerNetworkEndpoint{network: network, ipv4: ipv4, ipv6: ipv6}
}

func newContainerModeEndpoint(network *Network) *containerNetworkEndpoint {
	return &containerNetworkEndpoint{network: network}
}

func containerGroupBaseDir(homelabBaseDir string, ct config.ContainerReference) string {
	return fmt.Sprintf("%s/%s", homelabBaseDir, ct.Group)
}

func containerBaseDir(homelabBaseDir string, ct config.ContainerReference) string {
	return fmt.Sprintf("%s/%s/%s", homelabBaseDir, ct.Group, ct.Container)
}
