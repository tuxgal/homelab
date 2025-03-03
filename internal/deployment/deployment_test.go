package deployment

import (
	"fmt"
	"strings"
	"testing"

	dcontainer "github.com/docker/docker/api/types/container"
	dmount "github.com/docker/docker/api/types/mount"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/tuxgal/homelab/internal/cmdexec/fakecmdexec"
	"github.com/tuxgal/homelab/internal/config"
	"github.com/tuxgal/homelab/internal/docker/fakedocker"
	"github.com/tuxgal/homelab/internal/testhelpers"
	"github.com/tuxgal/homelab/internal/testutils"
	"github.com/tuxgal/homelab/internal/utils"
)

var buildDeploymentUsingReaderTests = []struct {
	name              string
	config            string
	ctxInfo           *testutils.TestContextInfo
	want              *config.Homelab
	wantDockerConfigs containerDockerConfigMap
}{
	{
		name: "Valid extensive config",
		config: `
global:
  baseDir: testdata/dummy-base-dir
  env:
    - var: MY_CONFIG_VAR_1
      value: MY_CONFIG_VAR_1_VALUE
    - var: MY_CONFIG_VAR_2
      value: MY_CONFIG_VAR_2_VALUE
    - var: MY_CONFIG_VAR_3
      valueCommand:
        - /foo/bar/some-env-var-cmd
    - var: ENV_GROUP_1
      value: fakegroup1
    - var: ENV_GROUP_2
      value: fakegroup2
    - var: ENV_DEST_DIR
      value: /foo123/bar123
    - var: ENV_DOMAIN
      value: my.custom.domain
    - var: ENV_DNS_SERVER
      value: 10.53.53.53
    - var: ENV_DNS_OPTION
      value: dns-opt-123
    - var: ENV_DNS_SEARCH
      value: custom-dns-search
    - var: MY_HOST_PORT_1
      value: 2122
    - var: MY_CT_PORT_1
      value: 4321
  mountDefs:
    - name: mount-def-1
      type: bind
      src: /abc/def/ghi
      dst: /pqr/stu/vwx
      readOnly: true
    - name: mount-def-2
      type: bind
      src: /abc1/def1
      dst: /pqr2/stu2/vwx2
    - name: homelab-self-signed-tls-cert
      type: bind
      src: /path/to/my/self/signed/cert/on/host
      dst: /path/to/my/self/signed/cert/on/container
  container:
    stopSignal: SIGTERM
    stopTimeout: 8
    restartPolicy:
      mode: unless-stopped
    domainName: example.tld
    dnsSearch:
      - dns-search-1
      - dns-search-2
    env:
      - var: MY_CONTAINER_ENV_VAR_1
        value: MY_CONTAINER_ENV_VAR_1_VALUE
      - var: MY_CONTAINER_ENV_VAR_2
        value: MY_CONTAINER_ENV_VAR_2_VALUE
      - var: MY_CONTAINER_ENV_VAR_3
        value: /foo2/bar2/some-other-env-var-cmd
    mounts:
      - name: mount-def-1
      - name: mount-def-2
      - name: mount-def-3
        type: bind
        src: /foo
        dst: /bar
        readOnly: true
    labels:
      - name: my-label-1
        value: my-label-1-value
      - name: my-label-2
        value: my-label-2-value
ipam:
  networks:
    bridgeModeNetworks:
      - name: group1-bridge
        hostInterfaceName: docker-grp1
        cidr:
          v4: 172.18.18.0/24
          v6: fd99:172:18:18::/64
        priority: 1
        containers:
          - ip:
              v4: 172.18.18.11
              v6: fd99:172:18:18::11
            container:
              group: group1
              container: ct1
          - ip:
              v4: 172.18.18.12
            container:
              group: group1
              container: ct2
      - name: group2-bridge
        hostInterfaceName: docker-grp2
        cidr:
          v4: 172.18.19.0/24
        priority: 1
        containers:
          - ip:
              v4: 172.18.19.11
            container:
              group: group2
              container: ct3
      - name: group3-bridge
        hostInterfaceName: docker-grp3
        cidr:
          v4: 172.18.20.0/24
        priority: 1
        containers:
          - ip:
              v4: 172.18.20.11
            container:
              group: group3
              container: ct4
      - name: common-bridge
        hostInterfaceName: docker-cmn
        cidr:
          v4: 172.18.30.0/24
          v6: fd99:172:18:30::/64
        priority: 2
        containers:
          - ip:
              v4: 172.18.30.11
              v6: fd99:172:18:30::11
            container:
              group: group1
              container: ct1
          - ip:
              v4: 172.18.30.12
            container:
              group: group1
              container: ct2
          - ip:
              v4: 172.18.30.13
              v6: fd99:172:18:30::13
            container:
              group: group2
              container: ct3
    containerModeNetworks:
      - name: group3-ct4
        container:
          group: group3
          container: ct4
        attachingContainers:
          - group: group3
            container: ct5
          - group: group3
            container: ct6
hosts:
  - name: fakehost
    allowedContainers:
      - group: group1
        container: ct1
      - group: group2
        container: ct3
      - group: group4
        container: ct7
  - name: host2
    allowedContainers:
      - group: group1
        container: ct2
  - name: host3
groups:
  - name: group1
    order: 1
  - name: group2
    order: 2
  - name: group3
    order: 3
  - name: group4
    order: 4
containers:
  - info:
      group: group1
      container: ct1
    config:
      env:
        - var: ENV_SRC_DEV
          value: /dev/src123
        - var: ENV_DST_DEV
          value: /dev/dst123
        - var: ENV_MY_ENV_VAL
          value: my-env-val
        - var: ENV_MY_ENV_2_VAL_CMD
          value: cat /foo/bar/baz.txt
        - var: MY_HOST_PORT_2
          value: 6789
        - var: MY_CT_PORT_2
          value: 8765
    image:
      image: tuxgal/homelab-base:master
      skipImagePull: false
      ignoreImagePullFailures: true
      pullImageBeforeStop: true
    metadata:
      labels:
        - name: my.ct1.label.name.1
          value: my.ct1.label.value.1
        - name: my.ct1.label.name.2
          value: my.ct1.label.value.2
    lifecycle:
      order: 1
      startPreHook:
        - $$CONTAINER_SCRIPTS_DIR$$/my-start-prehook.sh
      restartPolicy:
        mode: always
      autoRemove: true
      stopSignal: SIGHUP
      stopTimeout: 10
    user:
      user: $$USER_ID$$
      primaryGroup: $$USER_PRIMARY_GROUP_ID$$
      additionalGroups:
        - $$ENV_GROUP_1$$
        - $$ENV_GROUP_2$$
    fs:
      readOnlyRootfs: true
      mounts:
        - name: some-other-mount-1
          type: bind
          src: $$HOMELAB_BASE_DIR$$/abc
          dst: /abc
        - name: tmpfs-mount-1
          type: tmpfs
          dst: /tmp/cache-$$HUMAN_FRIENDLY_HOST_NAME$$
        - name: some-other-mount-2
          type: bind
          src: $$CONTAINER_BASE_DIR$$/some/random/dir
          dst: /xyz
        - name: blocky-config-mount
          type: bind
          src: $$CONTAINER_CONFIGS_DIR$$/generated/config.yml
          dst: /data/blocky/config/config.yml
          readOnly: true
        - name: tmpfs-mount-2
          type: tmpfs
          dst: /tmp/cache-$$USER_NAME$$
          tmpfsSize: 111111
        - name: homelab-self-signed-tls-cert
        - name: some-other-mount-3
          type: bind
          src: $$CONTAINER_DATA_DIR$$/my-data
          dst: $$ENV_DEST_DIR$$/my-data
      devices:
        static:
          - src: /dev/foostat1
            dst: /dev/barstat1
            disallowRead: false
            disallowWrite: true
            disallowMknod: true
          - src: $$ENV_SRC_DEV$$
            dst: $$ENV_DST_DEV$$
          - src: /dev/foostat2
            disallowRead: true
            disallowWrite: true
            disallowMknod: false
        dynamic:
          - $$CONTAINER_SCRIPTS_DIR$$/devices.sh
    network:
      hostName: Special-$$HOST_NAME$$-$$USER_PRIMARY_GROUP_NAME$$
      domainName: $$ENV_DOMAIN$$
      dnsServers:
        - 1.1.1.1
        - $$ENV_DNS_SERVER$$
      dnsOptions:
        - dns-option-1
        - $$ENV_DNS_OPTION$$
      dnsSearch:
        - dns-ct-search-1
        - $$ENV_DNS_SEARCH$$
      extraHosts:
        - my-extra-host-1
        - $$HOST_NAME$$-extra
      publishedPorts:
        - containerPort: $$MY_CT_PORT_1$$
          proto: tcp
          hostIp: $$HOST_IPV4$$
          hostPort: $$MY_HOST_PORT_1$$
        - containerPort: $$MY_CT_PORT_2$$
          proto: udp
          hostIp: 127.0.0.1
          hostPort: $$MY_HOST_PORT_2$$
        - containerPort: 5353
          proto: tcp
          hostIp: 0.0.0.0
          hostPort: 53
        - containerPort: 5353
          proto: tcp
          hostIp: '::'
          hostPort: 53
        - containerPort: 5353
          proto: udp
          hostIp: 0.0.0.0
          hostPort: 53
        - containerPort: 5353
          proto: udp
          hostIp: '::'
          hostPort: 53
    security:
      privileged: true
      sysctls:
        - key: net.ipv4.ip_forward
          value: 1
        - key: net.ipv4.conf.all.src_valid_mark
          value: 1
      capAdd:
        - SYS_RAWIO
        - SYS_ADMIN
      capDrop:
        - NET_ADMIN
        - SYS_MODULE
    health:
      cmd:
        - my-health-cmd
        - health-arg-1
        - health-arg-2
      retries: 3
      interval: 60s
      timeout: 10s
      startPeriod: 3m
      startInterval: 10s
    runtime:
      tty: true
      shmSize: 1g
      env:
        - var: MY_ENV
          value: $$ENV_MY_ENV_VAL$$
        - var: MY_ENV_2
          value: $$ENV_MY_ENV_2_VAL_CMD$$
        - var: MY_ENV_3
          value: SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName
      entrypoint:
        - my-custom-entrypoint
        - ep-arg1
        - ep-arg2
      args:
        - foo
        - bar-$$HUMAN_FRIENDLY_HOST_NAME$$
        - baz
  - info:
      group: group1
      container: ct2
    image:
      image: abc123/xyz123
    lifecycle:
      order: 2
      restartPolicy:
        mode: no
  - info:
      group: group2
      container: ct3
    image:
      image: abc123/xyz124
    lifecycle:
      order: 3
  - info:
      group: group3
      container: ct4
    image:
      image: abc123/xyz125
    lifecycle:
      order: 1
  - info:
      group: group3
      container: ct5
    image:
      image: abc123/xyz126
    lifecycle:
      order: 2
  - info:
      group: group3
      container: ct6
    image:
      image: abc123/xyz127
    lifecycle:
      order: 3
  - info:
      group: group4
      container: ct7
    image:
      image: abc123/xyz128
    lifecycle:
      order: 1
ignore:
  - foo
  - 4567
  - abc:
      - xyz
      - pqrs
  - random: 123
  - mymap:
      key1: val1
      key2: val2`,
		ctxInfo: &testutils.TestContextInfo{
			Executor: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
				ValidCmds: []fakecmdexec.FakeValidCmdInfo{
					{
						Cmd: []string{
							"testdata/dummy-base-dir/group1/ct1/scripts/devices.sh",
						},
						Output: "/dev/foodyn1:/dev/bardyn1:rw,/dev/foodyn2:/dev/bardyn2:rwm,/dev/foodyn3:/dev/bardyn3:m",
					},
				},
			}),
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: &config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Env: []config.ConfigEnv{
					{
						Var:   "MY_CONFIG_VAR_1",
						Value: "MY_CONFIG_VAR_1_VALUE",
					},
					{
						Var:   "MY_CONFIG_VAR_2",
						Value: "MY_CONFIG_VAR_2_VALUE",
					},
					{
						Var: "MY_CONFIG_VAR_3",
						ValueCommand: []string{
							"/foo/bar/some-env-var-cmd",
						},
					},
					{
						Var:   "ENV_GROUP_1",
						Value: "fakegroup1",
					},
					{
						Var:   "ENV_GROUP_2",
						Value: "fakegroup2",
					},
					{
						Var:   "ENV_DEST_DIR",
						Value: "/foo123/bar123",
					},
					{
						Var:   "ENV_DOMAIN",
						Value: "my.custom.domain",
					},
					{
						Var:   "ENV_DNS_SERVER",
						Value: "10.53.53.53",
					},
					{
						Var:   "ENV_DNS_OPTION",
						Value: "dns-opt-123",
					},
					{
						Var:   "ENV_DNS_SEARCH",
						Value: "custom-dns-search",
					},
					{
						Var:   "MY_HOST_PORT_1",
						Value: "2122",
					},
					{
						Var:   "MY_CT_PORT_1",
						Value: "4321",
					},
				},
				MountDefs: []config.Mount{
					{
						Name:     "mount-def-1",
						Type:     "bind",
						Src:      "/abc/def/ghi",
						Dst:      "/pqr/stu/vwx",
						ReadOnly: true,
					},
					{
						Name: "mount-def-2",
						Type: "bind",
						Src:  "/abc1/def1",
						Dst:  "/pqr2/stu2/vwx2",
					},
					{
						Name: "homelab-self-signed-tls-cert",
						Type: "bind",
						Src:  "/path/to/my/self/signed/cert/on/host",
						Dst:  "/path/to/my/self/signed/cert/on/container",
					},
				},
				Container: config.GlobalContainer{
					StopSignal:  "SIGTERM",
					StopTimeout: 8,
					RestartPolicy: config.ContainerRestartPolicy{
						Mode: "unless-stopped",
					},
					DomainName: "example.tld",
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
					Env: []config.ContainerEnv{
						{
							Var:   "MY_CONTAINER_ENV_VAR_1",
							Value: "MY_CONTAINER_ENV_VAR_1_VALUE",
						},
						{
							Var:   "MY_CONTAINER_ENV_VAR_2",
							Value: "MY_CONTAINER_ENV_VAR_2_VALUE",
						},
						{
							Var:   "MY_CONTAINER_ENV_VAR_3",
							Value: "/foo2/bar2/some-other-env-var-cmd",
						},
					},
					Mounts: []config.Mount{
						{
							Name: "mount-def-1",
						},
						{
							Name: "mount-def-2",
						},
						{
							Name:     "mount-def-3",
							Type:     "bind",
							Src:      "/foo",
							Dst:      "/bar",
							ReadOnly: true,
						},
					},
					Labels: []config.Label{
						{
							Name:  "my-label-1",
							Value: "my-label-1-value",
						},
						{
							Name:  "my-label-2",
							Value: "my-label-2-value",
						},
					},
				},
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "group1-bridge",
							HostInterfaceName: "docker-grp1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.18.0/24",
								V6: "fd99:172:18:18::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.18.11",
										IPv6: "fd99:172:18:18::11",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.18.12",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
							},
						},
						{
							Name:              "group2-bridge",
							HostInterfaceName: "docker-grp2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.19.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.19.11",
									},
									Container: config.ContainerReference{
										Group:     "group2",
										Container: "ct3",
									},
								},
							},
						},
						{
							Name:              "group3-bridge",
							HostInterfaceName: "docker-grp3",
							CIDR: config.NetworkCIDR{
								V4: "172.18.20.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.20.11",
									},
									Container: config.ContainerReference{
										Group:     "group3",
										Container: "ct4",
									},
								},
							},
						},
						{
							Name:              "common-bridge",
							HostInterfaceName: "docker-cmn",
							CIDR: config.NetworkCIDR{
								V4: "172.18.30.0/24",
								V6: "fd99:172:18:30::/64",
							},
							Priority: 2,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.30.11",
										IPv6: "fd99:172:18:30::11",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.30.12",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.30.13",
										IPv6: "fd99:172:18:30::13",
									},
									Container: config.ContainerReference{
										Group:     "group2",
										Container: "ct3",
									},
								},
							},
						},
					},
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "group3-ct4",
							Container: config.ContainerReference{
								Group:     "group3",
								Container: "ct4",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group:     "group3",
									Container: "ct5",
								},
								{
									Group:     "group3",
									Container: "ct6",
								},
							},
						},
					},
				},
			},
			Hosts: []config.Host{
				{
					Name: "fakehost",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "group1",
							Container: "ct1",
						},
						{
							Group:     "group2",
							Container: "ct3",
						},
						{
							Group:     "group4",
							Container: "ct7",
						},
					},
				},
				{
					Name: "host2",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "group1",
							Container: "ct2",
						},
					},
				},
				{
					Name: "host3",
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "group1",
					Order: 1,
				},
				{
					Name:  "group2",
					Order: 2,
				},
				{
					Name:  "group3",
					Order: 3,
				},
				{
					Name:  "group4",
					Order: 4,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "group1",
						Container: "ct1",
					},
					Config: config.ContainerConfigOptions{
						Env: []config.ConfigEnv{
							{
								Var:   "ENV_SRC_DEV",
								Value: "/dev/src123",
							},
							{
								Var:   "ENV_DST_DEV",
								Value: "/dev/dst123",
							},
							{
								Var:   "ENV_MY_ENV_VAL",
								Value: "my-env-val",
							},
							{
								Var:   "ENV_MY_ENV_2_VAL_CMD",
								Value: "cat /foo/bar/baz.txt",
							},
							{
								Var:   "MY_HOST_PORT_2",
								Value: "6789",
							},
							{
								Var:   "MY_CT_PORT_2",
								Value: "8765",
							},
						},
					},
					Image: config.ContainerImage{
						Image:                   "tuxgal/homelab-base:master",
						SkipImagePull:           false,
						IgnoreImagePullFailures: true,
						PullImageBeforeStop:     true,
					},
					Metadata: config.ContainerMetadata{
						Labels: []config.Label{
							{
								Name:  "my.ct1.label.name.1",
								Value: "my.ct1.label.value.1",
							},
							{
								Name:  "my.ct1.label.name.2",
								Value: "my.ct1.label.value.2",
							},
						},
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
						StartPreHook: []string{
							"testdata/dummy-base-dir/group1/ct1/scripts/my-start-prehook.sh",
						},
						RestartPolicy: config.ContainerRestartPolicy{
							Mode: "always",
						},
						AutoRemove:  true,
						StopSignal:  "SIGHUP",
						StopTimeout: 10,
					},
					User: config.ContainerUser{
						User:         "55555",
						PrimaryGroup: "44444",
						AdditionalGroups: []string{
							"fakegroup1",
							"fakegroup2",
						},
					},
					Filesystem: config.ContainerFilesystem{
						ReadOnlyRootfs: true,
						Mounts: []config.Mount{
							{
								Name: "some-other-mount-1",
								Type: "bind",
								Src:  "testdata/dummy-base-dir/abc",
								Dst:  "/abc",
							},
							{
								Name: "tmpfs-mount-1",
								Type: "tmpfs",
								Dst:  "/tmp/cache-FakeHost",
							},
							{
								Name: "some-other-mount-2",
								Type: "bind",
								Src:  "testdata/dummy-base-dir/group1/ct1/some/random/dir",
								Dst:  "/xyz",
							},
							{
								Name:     "blocky-config-mount",
								Type:     "bind",
								Src:      "testdata/dummy-base-dir/group1/ct1/configs/generated/config.yml",
								Dst:      "/data/blocky/config/config.yml",
								ReadOnly: true,
							},
							{
								Name:      "tmpfs-mount-2",
								Type:      "tmpfs",
								Dst:       "/tmp/cache-fakeuser",
								TmpfsSize: 111111,
							},
							{
								Name: "homelab-self-signed-tls-cert",
							},
							{
								Name: "some-other-mount-3",
								Type: "bind",
								Src:  "testdata/dummy-base-dir/group1/ct1/data/my-data",
								Dst:  "/foo123/bar123/my-data",
							},
						},
						Devices: config.ContainerDevice{
							Static: []config.Device{
								{
									Src:           "/dev/foostat1",
									Dst:           "/dev/barstat1",
									DisallowWrite: true,
									DisallowMknod: true,
								},
								{
									Src: "/dev/src123",
									Dst: "/dev/dst123",
								},
								{
									Src:           "/dev/foostat2",
									DisallowRead:  true,
									DisallowWrite: true,
								},
							},
							DynamicCommand: []string{
								"testdata/dummy-base-dir/group1/ct1/scripts/devices.sh",
							},
							Dynamic: []config.Device{
								{
									Src:           "/dev/foodyn1",
									Dst:           "/dev/bardyn1",
									DisallowMknod: true,
								},
								{
									Src: "/dev/foodyn2",
									Dst: "/dev/bardyn2",
								},
								{
									Src:           "/dev/foodyn3",
									Dst:           "/dev/bardyn3",
									DisallowRead:  true,
									DisallowWrite: true,
								},
							},
						},
					},
					Network: config.ContainerNetwork{
						HostName:   "Special-fakehost-fakegroup1",
						DomainName: "my.custom.domain",
						DNSServers: []string{
							"1.1.1.1",
							"10.53.53.53",
						},
						DNSOptions: []string{
							"dns-option-1",
							"dns-opt-123",
						},
						DNSSearch: []string{
							"dns-ct-search-1",
							"custom-dns-search",
						},
						ExtraHosts: []string{
							"my-extra-host-1",
							"fakehost-extra",
						},
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "4321",
								Protocol:      "tcp",
								HostIP:        "10.76.77.78",
								HostPort:      "2122",
							},
							{
								ContainerPort: "8765",
								Protocol:      "udp",
								HostIP:        "127.0.0.1",
								HostPort:      "6789",
							},
							{
								ContainerPort: "5353",
								Protocol:      "tcp",
								HostIP:        "0.0.0.0",
								HostPort:      "53",
							},
							{
								ContainerPort: "5353",
								Protocol:      "tcp",
								HostIP:        "::",
								HostPort:      "53",
							},
							{
								ContainerPort: "5353",
								Protocol:      "udp",
								HostIP:        "0.0.0.0",
								HostPort:      "53",
							},
							{
								ContainerPort: "5353",
								Protocol:      "udp",
								HostIP:        "::",
								HostPort:      "53",
							},
						},
					},
					Security: config.ContainerSecurity{
						Privileged: true,
						Sysctls: []config.Sysctl{
							{
								Key:   "net.ipv4.ip_forward",
								Value: "1",
							},
							{
								Key:   "net.ipv4.conf.all.src_valid_mark",
								Value: "1",
							},
						},
						CapAdd: []string{
							"SYS_RAWIO",
							"SYS_ADMIN",
						},
						CapDrop: []string{
							"NET_ADMIN",
							"SYS_MODULE",
						},
					},
					Health: config.ContainerHealth{
						Cmd: []string{
							"my-health-cmd",
							"health-arg-1",
							"health-arg-2",
						},
						Retries:       3,
						Interval:      "60s",
						Timeout:       "10s",
						StartPeriod:   "3m",
						StartInterval: "10s",
					},
					Runtime: config.ContainerRuntime{
						AttachToTty: true,
						ShmSize:     "1g",
						Env: []config.ContainerEnv{
							{
								Var:   "MY_ENV",
								Value: "my-env-val",
							},
							{
								Var:   "MY_ENV_2",
								Value: "cat /foo/bar/baz.txt",
							},
							{
								Var:   "MY_ENV_3",
								Value: "SomeHostName.FakeHost.SomeDomainName",
							},
						},
						Entrypoint: []string{
							"my-custom-entrypoint",
							"ep-arg1",
							"ep-arg2",
						},
						Args: []string{
							"foo",
							"bar-FakeHost",
							"baz",
						},
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "group1",
						Container: "ct2",
					},
					Image: config.ContainerImage{
						Image: "abc123/xyz123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 2,
						RestartPolicy: config.ContainerRestartPolicy{
							Mode: "no",
						},
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "group2",
						Container: "ct3",
					},
					Image: config.ContainerImage{
						Image: "abc123/xyz124",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 3,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "group3",
						Container: "ct4",
					},
					Image: config.ContainerImage{
						Image: "abc123/xyz125",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "group3",
						Container: "ct5",
					},
					Image: config.ContainerImage{
						Image: "abc123/xyz126",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 2,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "group3",
						Container: "ct6",
					},
					Image: config.ContainerImage{
						Image: "abc123/xyz127",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 3,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "group4",
						Container: "ct7",
					},
					Image: config.ContainerImage{
						Image: "abc123/xyz128",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		wantDockerConfigs: containerDockerConfigMap{
			config.ContainerReference{
				Group:     "group1",
				Container: "ct1",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Hostname:   "Special-fakehost-fakegroup1",
					Domainname: "my.custom.domain",
					User:       "55555:44444",
					ExposedPorts: nat.PortSet{
						"4321/tcp": struct{}{},
						"5353/tcp": struct{}{},
						"5353/udp": struct{}{},
						"8765/udp": struct{}{},
					},
					Tty: true,
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
						"MY_ENV=my-env-val",
						"MY_ENV_2=cat /foo/bar/baz.txt",
						"MY_ENV_3=SomeHostName.FakeHost.SomeDomainName",
					},
					Cmd: []string{
						"foo",
						"bar-FakeHost",
						"baz",
					},
					Healthcheck: &dcontainer.HealthConfig{
						Test: []string{
							"CMD",
							"my-health-cmd",
							"health-arg-1",
							"health-arg-2",
						},
						Interval:      utils.MustParseDuration("1m"),
						Timeout:       utils.MustParseDuration("10s"),
						StartPeriod:   utils.MustParseDuration("3m"),
						StartInterval: utils.MustParseDuration("10s"),
						Retries:       3,
					},
					Image: "tuxgal/homelab-base:master",
					Entrypoint: []string{
						"my-custom-entrypoint",
						"ep-arg1",
						"ep-arg2",
					},
					Labels: map[string]string{
						"my.ct1.label.name.1": "my.ct1.label.value.1",
						"my.ct1.label.name.2": "my.ct1.label.value.2",
					},
					StopSignal:  "SIGHUP",
					StopTimeout: testhelpers.NewInt(10),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
						"testdata/dummy-base-dir/abc:/abc",
						"testdata/dummy-base-dir/group1/ct1/some/random/dir:/xyz",
						"testdata/dummy-base-dir/group1/ct1/configs/generated/config.yml:/data/blocky/config/config.yml:ro",
						"/path/to/my/self/signed/cert/on/host:/path/to/my/self/signed/cert/on/container",
						"testdata/dummy-base-dir/group1/ct1/data/my-data:/foo123/bar123/my-data",
					},
					NetworkMode: "group1-bridge",
					PortBindings: nat.PortMap{
						"4321/tcp": []nat.PortBinding{
							{
								HostIP:   "10.76.77.78",
								HostPort: "2122",
							},
						},
						"5353/tcp": []nat.PortBinding{
							{
								HostIP:   "0.0.0.0",
								HostPort: "53",
							},
							{
								HostIP:   "::",
								HostPort: "53",
							},
						},
						"5353/udp": []nat.PortBinding{
							{
								HostIP:   "0.0.0.0",
								HostPort: "53",
							},
							{
								HostIP:   "::",
								HostPort: "53",
							},
						},
						"8765/udp": []nat.PortBinding{
							{
								HostIP:   "127.0.0.1",
								HostPort: "6789",
							},
						},
					},
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "always",
					},
					AutoRemove: true,
					CapAdd: []string{
						"SYS_RAWIO",
						"SYS_ADMIN",
					},
					CapDrop: []string{
						"NET_ADMIN",
						"SYS_MODULE",
					},
					DNS: []string{
						"1.1.1.1",
						"10.53.53.53",
					},
					DNSOptions: []string{
						"dns-option-1",
						"dns-opt-123",
					},
					DNSSearch: []string{
						"dns-ct-search-1",
						"custom-dns-search",
					},
					ExtraHosts: []string{
						"my-extra-host-1",
						"fakehost-extra",
					},
					GroupAdd: []string{
						"fakegroup1",
						"fakegroup2",
					},
					Privileged:     true,
					ReadonlyRootfs: true,
					ShmSize:        1073741824,
					Sysctls: map[string]string{
						"net.ipv4.conf.all.src_valid_mark": "1",
						"net.ipv4.ip_forward":              "1",
					},
					Resources: dcontainer.Resources{
						Devices: []dcontainer.DeviceMapping{
							{
								PathOnHost:        "/dev/foostat1",
								PathInContainer:   "/dev/barstat1",
								CgroupPermissions: "r",
							},
							{
								PathOnHost:        "/dev/src123",
								PathInContainer:   "/dev/dst123",
								CgroupPermissions: "rwm",
							},
							{
								PathOnHost:        "/dev/foostat2",
								PathInContainer:   "/dev/foostat2",
								CgroupPermissions: "m",
							},
							{
								PathOnHost:        "/dev/foodyn1",
								PathInContainer:   "/dev/bardyn1",
								CgroupPermissions: "rw",
							},
							{
								PathOnHost:        "/dev/foodyn2",
								PathInContainer:   "/dev/bardyn2",
								CgroupPermissions: "rwm",
							},
							{
								PathOnHost:        "/dev/foodyn3",
								PathInContainer:   "/dev/bardyn3",
								CgroupPermissions: "m",
							},
						},
					},
					Mounts: []dmount.Mount{
						{
							Type:   "tmpfs",
							Target: "/tmp/cache-FakeHost",
						},
						{
							Type:   "tmpfs",
							Target: "/tmp/cache-fakeuser",
							TmpfsOptions: &dmount.TmpfsOptions{
								SizeBytes: 111111,
							},
						},
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"group1-bridge": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.18.11",
								IPv6Address: "fd99:172:18:18::11",
							},
							Gateway:             "172.18.18.1",
							IPAddress:           "172.18.18.11",
							IPPrefixLen:         24,
							IPv6Gateway:         "fd99:172:18:18::1",
							GlobalIPv6Address:   "fd99:172:18:18::11",
							GlobalIPv6PrefixLen: 64,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "group1",
				Container: "ct2",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname: "example.tld",
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
					},
					Image:       "abc123/xyz123",
					StopTimeout: testhelpers.NewInt(8),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
					},
					NetworkMode: "group1-bridge",
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "no",
					},
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"group1-bridge": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.18.12",
							},
							Gateway:     "172.18.18.1",
							IPAddress:   "172.18.18.12",
							IPPrefixLen: 24,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "group2",
				Container: "ct3",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname: "example.tld",
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
					},
					Image:       "abc123/xyz124",
					StopTimeout: testhelpers.NewInt(8),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
					},
					NetworkMode: "group2-bridge",
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "unless-stopped",
					},
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"group2-bridge": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.19.11",
							},
							Gateway:     "172.18.19.1",
							IPAddress:   "172.18.19.11",
							IPPrefixLen: 24,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "group3",
				Container: "ct4",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname: "example.tld",
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
					},
					Image:       "abc123/xyz125",
					StopTimeout: testhelpers.NewInt(8),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
					},
					NetworkMode: "group3-bridge",
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "unless-stopped",
					},
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"group3-bridge": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.20.11",
							},
							Gateway:     "172.18.20.1",
							IPAddress:   "172.18.20.11",
							IPPrefixLen: 24,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "group3",
				Container: "ct5",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname: "example.tld",
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
					},
					Image:       "abc123/xyz126",
					StopTimeout: testhelpers.NewInt(8),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
					},
					NetworkMode: "container:group3-ct4",
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "unless-stopped",
					},
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
				},
			},
			config.ContainerReference{
				Group:     "group3",
				Container: "ct6",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname: "example.tld",
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
					},
					Image:       "abc123/xyz127",
					StopTimeout: testhelpers.NewInt(8),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
					},
					NetworkMode: "container:group3-ct4",
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "unless-stopped",
					},
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
				},
			},
			config.ContainerReference{
				Group:     "group4",
				Container: "ct7",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname: "example.tld",
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=/foo2/bar2/some-other-env-var-cmd",
					},
					Image:       "abc123/xyz128",
					StopTimeout: testhelpers.NewInt(8),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/foo:/bar:ro",
					},
					NetworkMode: "none",
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "unless-stopped",
					},
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
				},
			},
		},
	},
	{
		name: "Valid Groups Only config",
		config: `
global:
  baseDir: testdata/dummy-base-dir
groups:
  - name: group1
    order: 1
  - name: group2
    order: 3
  - name: group3
    order: 2`,
		want: &config.Homelab{
			Global: config.Global{
				BaseDir: "testdata/dummy-base-dir",
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "group1",
					Order: 1,
				},
				{
					Name:  "group2",
					Order: 3,
				},
				{
					Name:  "group3",
					Order: 2,
				},
			},
		},
		wantDockerConfigs: containerDockerConfigMap{},
	},
}

func TestBuildDeploymentUsingReader(t *testing.T) {
	t.Parallel()

	for _, test := range buildDeploymentUsingReaderTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := strings.NewReader(tc.config)
			ctx := testutils.NewVanillaTestContext()
			if tc.ctxInfo != nil {
				ctx = testutils.NewTestContext(tc.ctxInfo)
			}
			got, gotErr := FromReader(ctx, input)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromReader()", tc.name, gotErr)
				return
			}

			if !testhelpers.CmpDiff(t, "FromReader()", tc.name, "config", tc.want, got.Config) {
				return
			}

			if !testhelpers.CmpDiff(t, "FromReader()", tc.name, "docker configs", tc.wantDockerConfigs, got.dockerConfigs) {
				return
			}
		})
	}
}

var buildDeploymentFromConfigsPathTests = []struct {
	name              string
	configsPath       string
	want              *config.Homelab
	wantDockerConfigs containerDockerConfigMap
}{
	{
		name:        "Valid Multi File Config",
		configsPath: "parse-configs-valid",
		want: &config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Env: []config.ConfigEnv{
					{
						Var:   "MY_GLOBAL_FOO",
						Value: "MY_GLOBAL_BAR",
					},
				},
				Container: config.GlobalContainer{
					StopSignal:  "SIGTERM",
					StopTimeout: 5,
					RestartPolicy: config.ContainerRestartPolicy{
						Mode:          "on-failure",
						MaxRetryCount: 5,
					},
					DomainName: "somedomain",
				},
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100::11",
									},
									Container: config.ContainerReference{
										Group:     "g1",
										Container: "c1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.12",
									},
									Container: config.ContainerReference{
										Group:     "g1",
										Container: "c2",
									},
								},
							},
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.101.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.101.21",
									},
									Container: config.ContainerReference{
										Group:     "g2",
										Container: "c3",
									},
								},
							},
						},
						{
							Name:              "net-common",
							HostInterfaceName: "docker-cmn",
							CIDR: config.NetworkCIDR{
								V4: "172.19.200.0/24",
								V6: "fd99:172:19:200::/64",
							},
							Priority: 2,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.19.200.201",
										IPv6: "fd99:172:19:200::201",
									},
									Container: config.ContainerReference{
										Group:     "g1",
										Container: "c1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.19.200.202",
									},
									Container: config.ContainerReference{
										Group:     "g1",
										Container: "c2",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.19.200.203",
										IPv6: "fd99:172:19:200::203",
									},
									Container: config.ContainerReference{
										Group:     "g2",
										Container: "c3",
									},
								},
							},
						},
					},
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "g3-c4",
							Container: config.ContainerReference{
								Group:     "g3",
								Container: "c4",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group:     "g3",
									Container: "c5",
								},
								{
									Group:     "g3",
									Container: "c6",
								},
								{
									Group:     "g3",
									Container: "c7",
								},
							},
						},
					},
				},
			},
			Hosts: []config.Host{
				{
					Name: "fakehost",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "g1",
							Container: "c1",
						},
						{
							Group:     "g3",
							Container: "c4",
						},
					},
				},
				{
					Name: "host2",
				},
				{
					Name: "host3",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "g2",
							Container: "c3",
						},
					},
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
				{
					Name:  "g2",
					Order: 2,
				},
				{
					Name:  "g3",
					Order: 3,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "abc/xyz",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c2",
					},
					Image: config.ContainerImage{
						Image: "abc/xyz2",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 2,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g2",
						Container: "c3",
					},
					Image: config.ContainerImage{
						Image: "abc/xyz3",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g3",
						Container: "c4",
					},
					Image: config.ContainerImage{
						Image: "abc/xyz4",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		wantDockerConfigs: containerDockerConfigMap{
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz",
					StopTimeout: testhelpers.NewInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "net1",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"net1": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.100.11",
								IPv6Address: "fd99:172:18:100::11",
							},
							Gateway:             "172.18.100.1",
							IPAddress:           "172.18.100.11",
							IPPrefixLen:         24,
							IPv6Gateway:         "fd99:172:18:100::1",
							GlobalIPv6Address:   "fd99:172:18:100::11",
							GlobalIPv6PrefixLen: 64,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "g1",
				Container: "c2",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz2",
					StopTimeout: testhelpers.NewInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "net1",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"net1": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.100.12",
							},
							Gateway:     "172.18.100.1",
							IPAddress:   "172.18.100.12",
							IPPrefixLen: 24,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "g2",
				Container: "c3",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz3",
					StopTimeout: testhelpers.NewInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "net2",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"net2": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.101.21",
							},
							Gateway:     "172.18.101.1",
							IPAddress:   "172.18.101.21",
							IPPrefixLen: 24,
						},
					},
				},
			},
			config.ContainerReference{
				Group:     "g3",
				Container: "c4",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz4",
					StopTimeout: testhelpers.NewInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "none",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
			},
		},
	},
}

func TestBuildDeploymentFromConfigsPath(t *testing.T) {
	t.Parallel()

	for _, test := range buildDeploymentFromConfigsPathTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := fmt.Sprintf("%s/testdata/%s", testhelpers.Pwd(), tc.configsPath)
			got, gotErr := FromConfigsPath(testutils.NewVanillaTestContext(), p)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfigsPath()", tc.name, gotErr)
				return
			}

			if !testhelpers.CmpDiff(t, "FromConfigsPath()", tc.name, "config", tc.want, got.Config) {
				return
			}

			if !testhelpers.CmpDiff(t, "FromConfigsPath()", tc.name, "docker configs", tc.wantDockerConfigs, got.dockerConfigs) {
				return
			}
		})
	}
}

var buildDeploymentFromConfigsPathErrorTests = []struct {
	name        string
	configsPath string
	want        string
}{
	{
		name:        "Non-existing configs dir path",
		configsPath: "foo-bar",
		want:        `os.Stat\(\) failed on homelab configs path, reason: stat [^ ]+foo-bar: no such file or directory`,
	},
	{
		name:        "No configs",
		configsPath: "parse-configs-invalid-empty-dir",
		want:        `no homelab configs found in [^ ]+/testdata/parse-configs-invalid-empty-dir`,
	},
	{
		name:        "File configs dir path",
		configsPath: "parse-configs-invalid-with-empty-file/.empty",
		want:        `homelab configs path [^ ]+/testdata/parse-configs-invalid-with-empty-file/.empty must be a directory`,
	},
	{
		name:        "Unreadable config dir",
		configsPath: "/root",
		want:        `failed to read contents of directory /root, reason: open /root: permission denied`,
	},
	{
		name:        "Unreadable config file",
		configsPath: "parse-configs-invalid-unreadable-config",
		want:        `failed to read homelab config file [^ ]+/testdata/parse-configs-invalid-unreadable-config/invalid-symlink.yaml, reason: open [^ ]+/testdata/parse-configs-invalid-unreadable-config/invalid-symlink.yaml: no such file or directory`,
	},
	{
		name:        "Deep merge configs fail",
		configsPath: "parse-configs-invalid-deepmerge-fail",
		want:        `failed to deep merge config file [^ ]+parse-configs-invalid-deepmerge-fail/config2.yaml, reason: error due to parameter with value of primitive type: only maps and slices/arrays can be merged, which means you cannot have define the same key twice for parameters that are not maps or slices/arrays`,
	},
	{
		name:        "Invalid config key",
		configsPath: "parse-configs-invalid-config-key",
		want:        `(?s)failed to parse homelab config, reason: yaml: unmarshal errors:.+: field someInvalidKey not found in type config\.GlobalContainer`,
	},
}

func TestBuildDeploymentFromConfigsPathErrors(t *testing.T) {
	t.Parallel()

	for _, test := range buildDeploymentFromConfigsPathErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := tc.configsPath
			if !strings.HasPrefix(tc.configsPath, "/") {
				p = fmt.Sprintf("%s/testdata/%s", testhelpers.Pwd(), tc.configsPath)
			}

			_, gotErr := FromConfigsPath(testutils.NewVanillaTestContext(), p)
			if gotErr == nil {
				testhelpers.LogErrorNil(t, "FromConfigsPath()", tc.name, tc.want)
				return
			}

			if !testhelpers.RegexMatch(t, "FromConfigsPath()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var buildDeploymentFromConfigStringerTests = []struct {
	name   string
	config config.Homelab
	want   string
}{
	{
		name: "Valid Empty Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
		},
		want: `Deployment{Groups:\[empty\], Networks:\[empty\]}`,
	},
	{
		name: "Valid IPAM Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.101.0/24",
							},
							Priority: 1,
						},
					},
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net3",
							Container: config.ContainerReference{
								Group:     "g5",
								Container: "ct101",
							},
						},
						{
							Name: "net4",
							Container: config.ContainerReference{
								Group:     "g6",
								Container: "ct201",
							},
						},
					},
				},
			},
		},
		want: `Deployment{Groups:\[empty\], Networks:\[{Network \(Bridge\) Name: net1}, {Network \(Bridge\) Name: net2}, {Network \(Container\) Name: net3}, {Network \(Container\) Name: net4}\]}`,
	},
	{
		name: "Valid Containers Without Hosts And IPAM Configs",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
				{
					Name:  "g2",
					Order: 2,
				},
				{
					Name:  "g3",
					Order: 3,
				},
				{
					Name:  "g4",
					Order: 4,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c2",
					},
					Image: config.ContainerImage{
						Image: "foo/bar2:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g2",
						Container: "c3",
					},
					Image: config.ContainerImage{
						Image: "foo/bar3:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g4",
						Container: "c4",
					},
					Image: config.ContainerImage{
						Image: "foo/bar4:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g4",
						Container: "c5",
					},
					Image: config.ContainerImage{
						Image: "foo/bar5:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `Deployment{Groups:\[Group{Name:g1 Containers:\[Container{Name:g1-c1}, Container{Name:g1-c2}\]}, Group{Name:g2 Containers:\[Container{Name:g2-c3}\]}, Group{Name:g3 Containers:\[empty\]}, Group{Name:g4 Containers:\[Container{Name:g4-c4}, Container{Name:g4-c5}\]}\], Networks:\[empty\]}`,
	},
}

func TestBuildDeploymentFromConfigStringer(t *testing.T) {
	t.Parallel()

	for _, test := range buildDeploymentFromConfigStringerTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dep, gotErr := FromConfig(testutils.NewVanillaTestContext(), &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			got := dep.String()
			if !testhelpers.RegexMatch(t, "FromConfig()", tc.name, "deployment string representation", tc.want, got) {
				return
			}

		})
	}
}

var buildDeploymentFromConfigErrorTests = []struct {
	name   string
	config config.Homelab
	want   string
}{
	{
		name:   "Empty Base Dir",
		config: config.Homelab{},
		want:   `homelab base directory cannot be empty`,
	},
	{
		name: "Non-Existing Base Dir Path",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: "/foo/bar",
			},
		},
		want: `os.Stat\(\) failed on homelab base directory path, reason: stat /foo/bar: no such file or directory`,
	},
	{
		name: "Base Dir Path Points To A File",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: "testdata/dummy-base-dir/.empty",
			},
		},
		want: `homelab base directory path testdata/dummy-base-dir/\.empty must be a directory`,
	},
	{
		name: "Empty Global Config Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Env: []config.ConfigEnv{
					{
						Value: "foo-bar",
					},
				},
			},
		},
		want: `empty env var in global config`,
	},
	{
		name: "Duplicate Global Config Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Env: []config.ConfigEnv{
					{
						Var:   "FOO",
						Value: "foo-bar",
					},
					{
						Var:   "FOO2",
						Value: "foo-bar-2",
					},
					{
						Var:   "FOO",
						Value: "foo-bar-3",
					},
				},
			},
		},
		want: `env var FOO specified more than once in global config`,
	},
	{
		name: "Global Config Env Var Without Value And ValueCommand",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Env: []config.ConfigEnv{
					{
						Var: "FOO",
					},
				},
			},
		},
		want: `neither value nor valueCommand specified for env var FOO in global config`,
	},
	{
		name: "Global Config Env Var With Both Value And ValueCommand",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Env: []config.ConfigEnv{
					{
						Var:   "FOO",
						Value: "my-foo-bar",
						ValueCommand: []string{
							"/foo/bar/baz",
						},
					},
				},
			},
		},
		want: `exactly one of value or valueCommand must be specified for env var FOO in global config`,
	},
	{
		name: "Global Config Empty Mount Def Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Type:     "bind",
						Src:      "/foo",
						Dst:      "/bar",
						ReadOnly: true,
					},
				},
			},
		},
		want: `mount name cannot be empty in global config mount defs`,
	},
	{
		name: "Global Config Duplicate Mount Defs",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "mount-foo1",
						Type: "bind",
						Src:  "/foo1",
						Dst:  "/bar1",
					},
					{
						Name: "mount-foo2",
						Type: "bind",
						Src:  "/foo2",
						Dst:  "/bar2",
					},
					{
						Name: "mount-foo1",
						Type: "bind",
						Src:  "/foo3",
						Dst:  "/bar3",
					},
				},
			},
		},
		want: `mount name mount-foo1 defined more than once in global config mount defs`,
	},
	{
		name: "Global Config Mount Def With Invalid Mount Type",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "garbage",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
			},
		},
		want: `unsupported mount type garbage for mount foo in global config mount defs`,
	},
	{
		name: "Global Config Mount Def With Empty Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "bind",
						Dst:  "/bar",
					},
				},
			},
		},
		want: `bind mount name foo cannot have an empty value for src in global config mount defs`,
	},
	{
		name: "Global Config Tmpfs Mount Def With Non-Empty Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "tmpfs",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
			},
		},
		want: `tmpfs mount name foo cannot have a non-empty value for src in global config mount defs`,
	},
	{
		name: "Global Config Mount Def With Empty Dst",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for dst in global config mount defs`,
	},
	{
		name: "Global Config Bind Mount Def With Tmpfs Size",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name:      "foo",
						Type:      "bind",
						Src:       "/foo",
						Dst:       "/bar",
						TmpfsSize: 1000,
					},
				},
			},
		},
		want: `bind mount name foo cannot specify tmpfs size in global config mount defs`,
	},
	{
		name: "Global Config Tmpfs Mount Def With Negative Tmpfs Size",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name:      "foo",
						Type:      "tmpfs",
						Dst:       "/bar",
						TmpfsSize: -1000,
					},
				},
			},
		},
		want: `tmpfs mount name foo cannot specify a negative tmpfs size -1000 in global config mount defs`,
	},
	{
		name: "Global Container Config Negative Stop Timeout",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					StopTimeout: -1,
				},
			},
		},
		want: `container stop timeout -1 cannot be negative in global container config`,
	},
	{
		name: "Global Container Config Restart Policy MaxRetryCount Set With Non-On-Failure Mode",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					RestartPolicy: config.ContainerRestartPolicy{
						Mode:          "always",
						MaxRetryCount: 5,
					},
				},
			},
		},
		want: `restart policy max retry count can be set only when the mode is on-failure in global container config`,
	},
	{
		name: "Global Container Config Invalid Restart Policy Mode",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					RestartPolicy: config.ContainerRestartPolicy{
						Mode: "garbage",
					},
				},
			},
		},
		want: `invalid restart policy mode garbage in global container config, valid values are \[ 'no', 'always', 'on-failure', 'unless-stopped' \]`,
	},
	{
		name: "Global Container Config Negative Restart Policy MaxRetryCount",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					RestartPolicy: config.ContainerRestartPolicy{
						Mode:          "on-failure",
						MaxRetryCount: -1,
					},
				},
			},
		},
		want: `restart policy max retry count -1 cannot be negative in global container config`,
	},
	{
		name: "Empty Global Container Config Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Env: []config.ContainerEnv{
						{
							Value: "foo-bar",
						},
					},
				},
			},
		},
		want: `empty env var in global container config`,
	},
	{
		name: "Duplicate Global Container Config Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Env: []config.ContainerEnv{
						{
							Var:   "FOO",
							Value: "foo-bar",
						},
						{
							Var:   "FOO2",
							Value: "foo-bar-2",
						},
						{
							Var:   "FOO",
							Value: "foo-bar-3",
						},
					},
				},
			},
		},
		want: `env var FOO specified more than once in global container config`,
	},
	{
		name: "Global Container Config Env Var Without Value",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Env: []config.ContainerEnv{
						{
							Var: "FOO",
						},
					},
				},
			},
		},
		want: `value not specified for env var FOO in global container config`,
	},
	{
		name: "Global Container Config Empty Mount Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Type:     "bind",
							Src:      "/foo",
							Dst:      "/bar",
							ReadOnly: true,
						},
					},
				},
			},
		},
		want: `mount name cannot be empty in global container config mounts`,
	},
	{
		name: "Global Container Config Duplicate Mounts",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "mount-foo1",
							Type: "bind",
							Src:  "/foo1",
							Dst:  "/bar1",
						},
						{
							Name: "mount-foo2",
							Type: "bind",
							Src:  "/foo2",
							Dst:  "/bar2",
						},
						{
							Name: "mount-foo1",
							Type: "bind",
							Src:  "/foo3",
							Dst:  "/bar3",
						},
					},
				},
			},
		},
		want: `mount name mount-foo1 defined more than once in global container config mounts`,
	},
	{
		name: "Global Container Config Mount With Invalid Mount Type",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo",
							Type: "garbage",
							Src:  "/foo",
							Dst:  "/bar",
						},
					},
				},
			},
		},
		want: `unsupported mount type garbage for mount foo in global container config mounts`,
	},
	{
		name: "Global Container Config Mount With Empty Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo",
							Type: "bind",
							Dst:  "/bar",
						},
					},
				},
			},
		},
		want: `bind mount name foo cannot have an empty value for src in global container config mounts`,
	},
	{
		name: "Global Container Config Tmpfs Mount With Non-Empty Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo",
							Type: "tmpfs",
							Src:  "/foo",
							Dst:  "/bar",
						},
					},
				},
			},
		},
		want: `tmpfs mount name foo cannot have a non-empty value for src in global container config mounts`,
	},
	{
		name: "Global Container Config Mount With Empty Dst",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo",
							Type: "bind",
							Src:  "/foo",
						},
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for dst in global container config mounts`,
	},
	{
		name: "Global Container Config Bind Mount With Tmpfs Size",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name:      "foo",
							Type:      "bind",
							Src:       "/foo",
							Dst:       "/bar",
							TmpfsSize: 10000,
						},
					},
				},
			},
		},
		want: `bind mount name foo cannot specify tmpfs size in global container config mounts`,
	},
	{
		name: "Global Container Config Tmpfs Mount With Negative Tmpfs Size",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name:      "foo",
							Type:      "tmpfs",
							Dst:       "/bar",
							TmpfsSize: -10000,
						},
					},
				},
			},
		},
		want: `tmpfs mount name foo cannot specify a negative tmpfs size -10000 in global container config mounts`,
	},
	{
		name: "Global Container Config Mount Def Reference Not Found",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo2",
						},
					},
				},
			},
		},
		want: `mount specified by just the name foo2 not found in defs in global container config mounts`,
	},
	{
		name: "Empty Global Container Config Label Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Labels: []config.Label{
						{
							Value: "foo-bar",
						},
					},
				},
			},
		},
		want: `empty label name in global container config`,
	},
	{
		name: "Duplicate Global Container Config Label Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Labels: []config.Label{
						{
							Name:  "FOO",
							Value: "foo-bar",
						},
						{
							Name:  "FOO2",
							Value: "foo-bar-2",
						},
						{
							Name:  "FOO",
							Value: "foo-bar-3",
						},
					},
				},
			},
		},
		want: `label name FOO specified more than once in global container config`,
	},
	{
		name: "Global Container Config Empty Label Value",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Labels: []config.Label{
						{
							Name: "FOO",
						},
					},
				},
			},
		},
		want: `empty label value for label FOO in global container config`,
	},
	{
		name: "Empty Bridge Mode Network Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `network name cannot be empty`,
	},
	{
		name: "Duplicate Bridge Mode Network",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1-2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.101.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `network net1 defined more than once in the IPAM config`,
	},
	{
		name: "Empty Host Interface Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name: "net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `host interface name of network net1 cannot be empty`,
	},
	{
		name: "Duplicate Network Host Interface Names",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.101.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `host interface name docker-net1 of network net2 is already used by another network in the IPAM config`,
	},
	{
		name: "Empty Bridge Mode Network Priority",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
						},
					},
				},
			},
		},
		want: `network net1 cannot have a non-positive priority 0`,
	},
	{
		name: "Invalid v4 CIDR - Empty",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `v4 CIDR  of network net1 is invalid, reason: netip\.ParsePrefix\(""\): no '/'`,
	},
	{
		name: "Invalid v4 CIDR - Unparsable",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "garbage-cidr",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR garbage-cidr of network net1 is invalid, reason: netip\.ParsePrefix\("garbage-cidr"\): no '/'`,
	},
	{
		name: "Invalid v4 CIDR - Missing /",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.16",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.100\.16 of network net1 is invalid, reason: netip\.ParsePrefix\("172\.18\.100\.16"\): no '/'`,
	},
	{
		name: "Invalid v4 CIDR - Wrong Prefix Length",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/33",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.100\.0/33 of network net1 is invalid, reason: netip\.ParsePrefix\("172\.18\.100\.0/33"\): prefix length out of range`,
	},
	{
		name: "Invalid v4 CIDR - IPv6",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "2002::1234:abcd:ffff:c0a8:101/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 2002::1234:abcd:ffff:c0a8:101/64 of network net1 is not an IPv4 subnet CIDR`,
	},
	{
		name: "Invalid v4 CIDR - Octets Out Of Range",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.260.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.260\.0/24 of network net1 is invalid, reason: netip\.ParsePrefix\("172\.18\.260\.0/24"\): ParseAddr\("172\.18\.260\.0"\): IPv4 field has value >255`,
	},
	{
		name: "Invalid v4 CIDR - Not A Network Address",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.1/25",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.100\.1/25 of network net1 is not the same as the network address 172\.18\.100\.0/25`,
	},
	{
		name: "Invalid v4 CIDR - Long Prefix 31",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/31",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.100\.0/31 of network net1 \(prefix length: 31\) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations`,
	},
	{
		name: "Invalid v4 CIDR - Long Prefix 32",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/32",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.100\.0/32 of network net1 \(prefix length: 32\) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations`,
	},
	{
		name: "Non-RFC1918 v4 CIDR - Public IPv4",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "11.12.13.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 11\.12\.13\.0/24 of network net1 is not within the RFC1918 private address space`,
	},
	{
		name: "Non-RFC1918 v4 CIDR - Link Local",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "169.254.10.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 169\.254\.10\.0/24 of network net1 is not within the RFC1918 private address space`,
	},
	{
		name: "Non-RFC1918 v4 CIDR - Multicast",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "224.0.0.0/26",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 224\.0\.0\.0/26 of network net1 is not within the RFC1918 private address space`,
	},
	{
		name: "Overlapping v4 CIDR",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.0.0/16",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v4 CIDR 172\.18\.0\.0/16 of network net2 overlaps with v4 CIDR 172\.18\.100\.0/24 of network net1`,
	},
	{
		name: "Invalid v6 CIDR - Unparsable",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "garbage-cidr",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR garbage-cidr of network net1 is invalid, reason: netip\.ParsePrefix\("garbage-cidr"\): no '/'`,
	},
	{
		name: "Invalid v6 CIDR - Missing /",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100:16::",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fd99:172:18:100:16:: of network net1 is invalid, reason: netip\.ParsePrefix\("fd99:172:18:100:16::"\): no '/'`,
	},
	{
		name: "Invalid v6 CIDR - Wrong Prefix Length - Larger",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/63",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fd99:172:18:100::/63 of network net1 \(prefix length: 63\) must have a prefix length 64 as per the convention for IPv6 networks`,
	},
	{
		name: "Invalid v6 CIDR - Wrong Prefix Length - Smaller",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/65",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fd99:172:18:100::/65 of network net1 \(prefix length: 65\) must have a prefix length 64 as per the convention for IPv6 networks`,
	},
	{
		name: "Invalid v6 CIDR - IPv4",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "172.18.101.0/24",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR 172\.18\.101\.0/24 of network net1 is not an IPv6 subnet CIDR`,
	},
	{
		name: "Invalid v6 CIDR - Hextets Out Of Range",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:f100a::/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fd99:172:18:f100a::/64 of network net1 is invalid, reason: netip\.ParsePrefix\("fd99:172:18:f100a::/64"\): ParseAddr\("fd99:172:18:f100a::"\): each group must have 4 or less digits \(at "f100a::"\)`,
	},
	{
		name: "Invalid v6 CIDR - Not A Network Address",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::1/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fd99:172:18:100::1/64 of network net1 is not the same as the network address fd99:172:18:100::/64`,
	},
	{
		name: "Non-ULA v6 CIDR - GUA IPv6",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "2001:1234:5678:90ab::/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR 2001:1234:5678:90ab::/64 of network net1 is not within the ULA private address space`,
	},
	{
		name: "Non-ULA v6 CIDR - Link Local",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fe80:1234:5678:90ab::/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fe80:1234:5678:90ab::/64 of network net1 is not within the ULA private address space`,
	},
	{
		name: "Non-ULA v6 CIDR - Multicast",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "ff00:1234:5678:90ab::/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR ff00:1234:5678:90ab::/64 of network net1 is not within the ULA private address space`,
	},
	{
		name: "ULA v6 CIDR - Reserved ULA Prefix IPv6",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fcab:1234:5678:90ab::/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fcab:1234:5678:90ab::/64 of network net1 overlaps with the reserved ULA prefix fc00::/8`,
	},
	{
		name: "Overlapping v6 CIDR",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:1234:5678:90ab::/64",
							},
							Priority: 1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.101.0/24",
								V6: "fd99:1234:5678:90ab::/64",
							},
							Priority: 1,
						},
					},
				},
			},
		},
		want: `v6 CIDR fd99:1234:5678:90ab::/64 of network net2 overlaps with v6 CIDR fd99:1234:5678:90ab::/64 of network net1`,
	},
	{
		name: "Bridge Mode Network v4 IP Invalid Container Reference - Empty Group",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
									},
									Container: config.ContainerReference{
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Bridge Mode Network v4 IP Invalid Container Reference - Empty Container",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
									},
									Container: config.ContainerReference{
										Group: "g1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Invalid Container v4 IP - Unparsable",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "garbage-ip",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid v4 IP garbage-ip, reason: ParseAddr\("garbage-ip"\): unable to parse IP`,
	},
	{
		name: "Invalid Container v4 IP - Too Short",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid v4 IP 172\.18\.100, reason: ParseAddr\("172\.18\.100"\): IPv4 address too short`,
	},
	{
		name: "Invalid Container v4 IP - Too Long",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.1.2.3.4",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid v4 IP 172\.18\.100\.1\.2\.3\.4, reason: ParseAddr\("172\.18\.100\.1\.2\.3\.4"\): IPv4 address too long`,
	},
	{
		name: "Container v4 IP Not Within Network CIDR",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.101.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have a v4 IP 172\.18\.101\.2 that does not belong to the network v4 CIDR 172\.18\.100\.0/24`,
	},
	{
		name: "Container v4 IP same as Network Address",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.0",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP 172\.18\.100\.0 matching the network address 172\.18\.100\.0`,
	},
	{
		name: "Container v4 IP same as Gateway Address",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.1",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP 172\.18\.100\.1 matching the gateway address 172\.18\.100\.1`,
	},
	{
		name: "Multiple v4 Endpoints For Same Container Within A Bridge Mode Network",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.3",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} cannot have multiple endpoints in network net1`,
	},
	{
		name: "Duplicate v4 Container IPs",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.3",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.4",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct3",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct4",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `IP 172\.18\.100\.2 of container {Group:group1 Container:ct4} is already in use by another container in network net1`,
	},
	{
		name: "Bridge Mode Network v6 IP Invalid Container Reference - Empty Group",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100::11",
									},
									Container: config.ContainerReference{
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Bridge Mode Network v6 IP Invalid Container Reference - Empty Container",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100::11",
									},
									Container: config.ContainerReference{
										Group: "g1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Invalid Container v6 IP - Unparsable",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "garbage-ip",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid v6 IP garbage-ip, reason: ParseAddr\("garbage-ip"\): unable to parse IP`,
	},
	{
		name: "Container With v6 IP But Subnet with Only v4",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100::11",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 specified a v6 IP address fd99:172:18:100::11 when the network has no v6 subnet CIDRs defined`,
	},

	{
		name: "Invalid Container v6 IP - Too Short",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100:1111:2222:3333",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid v6 IP fd99:172:18:100:1111:2222:3333, reason: ParseAddr\("fd99:172:18:100:1111:2222:3333"\): address string too short`,
	},
	{
		name: "Invalid Container v6 IP - Too Long",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100:1111:2222:3333:4444:5555",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid v6 IP fd99:172:18:100:1111:2222:3333:4444:5555, reason: ParseAddr\("fd99:172:18:100:1111:2222:3333:4444:5555"\): trailing garbage after address \(at \"5555\"\)`,
	},
	{
		name: "Container v6 IP Not Within Network CIDR",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:101::11",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have a v6 IP fd99:172:18:101::11 that does not belong to the network v6 CIDR fd99:172:18:100::/64`,
	},
	{
		name: "Container v6 IP same as Network Address",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100::",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP fd99:172:18:100:: matching the network address fd99:172:18:100::`,
	},
	{
		name: "Container v6 IP same as Gateway Address",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.11",
										IPv6: "fd99:172:18:100::1",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP fd99:172:18:100::1 matching the gateway address fd99:172:18:100::1`,
	},
	{
		name: "Multiple v6 Endpoints For Same Container Within A Bridge Mode Network",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
										IPv6: "fd99:172:18:100::2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.3",
										IPv6: "fd99:172:18:100::3",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} cannot have multiple endpoints in network net1`,
	},
	{
		name: "Duplicate v6 Container IPs",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
								V6: "fd99:172:18:100::/64",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
										IPv6: "fd99:172:18:100::2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.3",
										IPv6: "fd99:172:18:100::3",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.4",
										IPv6: "fd99:172:18:100::4",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct3",
									},
								},
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.5",
										IPv6: "fd99:172:18:100::2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct4",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `IP fd99:172:18:100::2 of container {Group:group1 Container:ct4} is already in use by another container in network net1`,
	},
	{
		name: "Multiple Same Priority Bridge Mode Network Endpoints For Same Container",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR: config.NetworkCIDR{
								V4: "172.18.101.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.101.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} cannot have multiple bridge mode network endpoints whose networks have the same priority 1`,
	},
	{
		name: "Empty Container Mode Network Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Container: config.ContainerReference{
								Group:     "some-group",
								Container: "some-container",
							},
						},
					},
				},
			},
		},
		want: `network name cannot be empty`,
	},
	{
		name: "Duplicate Container Mode Network Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group-1",
								Container: "some-container-1",
							},
						},
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group-2",
								Container: "some-container-2",
							},
						},
					},
				},
			},
		},
		want: `network net1 defined more than once in the IPAM config`,
	},
	{
		name: "Duplicate Bridge/Container Mode Networks",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
						},
					},
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group",
								Container: "some-container",
							},
						},
					},
				},
			},
		},
		want: `network net1 defined more than once in the IPAM config`,
	},
	{
		name: "Container Mode Network Invalid Container Reference - Empty Group",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Container: "some-container",
							},
						},
					},
				},
			},
		},
		want: `container reference of container mode network net1 is invalid, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Container Mode Network Invalid Container Reference - Empty Container",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group: "some-group",
							},
						},
					},
				},
			},
		},
		want: `container reference of container mode network net1 is invalid, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Container Mode Network Invalid Attaching Container Reference - Empty Group",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group",
								Container: "some-container",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Container: "ct1",
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Container Mode Network Invalid Attaching Container Reference - Empty Container",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group",
								Container: "some-container",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group: "g1",
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Multiple Container Mode Network Stacks For Same Container Within Same Network Stack",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group",
								Container: "some-container",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group:     "group1",
									Container: "ct1",
								},
								{
									Group:     "group1",
									Container: "ct2",
								},
								{
									Group:     "group2",
									Container: "ct3",
								},
								{
									Group:     "group1",
									Container: "ct2",
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct2} is connected to multiple container mode network stacks`,
	},
	{
		name: "Multiple Container Mode Network Stacks For Same Container Across Network Stacks",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net1",
							Container: config.ContainerReference{
								Group:     "some-group-1",
								Container: "some-container-1",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group:     "group1",
									Container: "ct1",
								},
							},
						},
						{
							Name: "net2",
							Container: config.ContainerReference{
								Group:     "some-group-2",
								Container: "some-container-2",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group:     "group2",
									Container: "ct2",
								},
								{
									Group:     "group1",
									Container: "ct1",
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} is connected to multiple container mode network stacks`,
	},
	{
		name: "Container Connected To Both Bridge Mode and Container Mode Networks",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			IPAM: config.IPAM{
				Networks: config.Networks{
					BridgeModeNetworks: []config.BridgeModeNetwork{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR: config.NetworkCIDR{
								V4: "172.18.100.0/24",
							},
							Priority: 1,
							Containers: []config.ContainerIPInfo{
								{
									IP: config.ContainerIP{
										IPv4: "172.18.100.2",
									},
									Container: config.ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
					ContainerModeNetworks: []config.ContainerModeNetwork{
						{
							Name: "net2",
							Container: config.ContainerReference{
								Group:     "some-group-1",
								Container: "some-container-1",
							},
							AttachingContainers: []config.ContainerReference{
								{
									Group:     "group1",
									Container: "ct1",
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} is connected to both bridge mode and container mode network stacks`,
	},
	{
		name: "Empty Host Name In Hosts Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Hosts: []config.Host{
				{
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "g1",
							Container: "ct1",
						},
					},
				},
			},
		},
		want: `host name cannot be empty in the hosts config`,
	},
	{
		name: "Duplicate Host Name In Hosts Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Hosts: []config.Host{
				{
					Name: "h1",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "g1",
							Container: "ct1",
						},
						{
							Group:     "g2",
							Container: "ct2",
						},
					},
				},
				{
					Name: "h1",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "g3",
							Container: "ct3",
						},
						{
							Group:     "g4",
							Container: "ct4",
						},
					},
				},
			},
		},
		want: `host h1 defined more than once in the hosts config`,
	},
	{
		name: "Invalid Container Reference Within Host Config - Empty Group",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Hosts: []config.Host{
				{
					Name: "h1",
					AllowedContainers: []config.ContainerReference{
						{
							Container: "ct1",
						},
					},
				},
			},
		},
		want: `allowed container config within host h1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Invalid Container Reference Within Host Config - Empty Container",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Hosts: []config.Host{
				{
					Name: "h1",
					AllowedContainers: []config.ContainerReference{
						{
							Group: "g1",
						},
					},
				},
			},
		},
		want: `allowed container config within host h1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Duplicate Container Within Host Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Hosts: []config.Host{
				{
					Name: "h1",
					AllowedContainers: []config.ContainerReference{
						{
							Group:     "g1",
							Container: "ct1",
						},
						{
							Group:     "g2",
							Container: "ct2",
						},
						{
							Group:     "g3",
							Container: "ct3",
						},
						{
							Group:     "g2",
							Container: "ct2",
						},
					},
				},
			},
		},
		want: `container {Group:g2 Container:ct2} defined more than once in the hosts config for host h1`,
	},
	{
		name: "Empty Group Name In Groups Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Order: 1,
				},
			},
		},
		want: `group name cannot be empty in the groups config`,
	},
	{
		name: "Duplicate Container Group Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
				{
					Name:  "g2",
					Order: 1,
				},
				{
					Name:  "g3",
					Order: 1,
				},
				{
					Name:  "g1",
					Order: 2,
				},
			},
		},
		want: `group g1 defined more than once in the groups config`,
	},
	{
		name: "Container Group Without Order",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name: "g1",
				},
			},
		},
		want: `group g1 cannot have a non-positive order 0`,
	},
	{
		name: "Container Group With Zero Order",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 0,
				},
			},
		},
		want: `group g1 cannot have a non-positive order 0`,
	},
	{
		name: "Container Group With Negative Order",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: -1,
				},
			},
		},
		want: `group g1 cannot have a non-positive order -1`,
	},
	{
		name: "Container Group Definition Missing",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
				},
			},
		},
		want: `group definition missing in groups config for the container {Group:g1 Container:c1} in the containers config`,
	},
	{
		name: "Duplicate Container Definition",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c2",
					},
					Image: config.ContainerImage{
						Image: "foo/bar2:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar3:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `container {Group:g1 Container:c1} defined more than once in the containers config`,
	},
	{
		name: "Empty Container Config Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Config: config.ContainerConfigOptions{
						Env: []config.ConfigEnv{
							{
								Value: "foo-bar",
							},
						},
					},
				},
			},
		},
		want: `empty env var in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Config Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Config: config.ContainerConfigOptions{
						Env: []config.ConfigEnv{
							{
								Var:   "FOO",
								Value: "foo-bar",
							},
							{
								Var:   "FOO2",
								Value: "foo-bar-2",
							},
							{
								Var:   "FOO",
								Value: "foo-bar-3",
							},
						},
					},
				},
			},
		},
		want: `env var FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Env Var Without Value And ValueCommand",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Config: config.ContainerConfigOptions{
						Env: []config.ConfigEnv{
							{
								Var: "FOO",
							},
						},
					},
				},
			},
		},
		want: `neither value nor valueCommand specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Env Var With Both Value And ValueCommand",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Config: config.ContainerConfigOptions{
						Env: []config.ConfigEnv{
							{
								Var:   "FOO",
								Value: "my-foo-bar",
								ValueCommand: []string{
									"/foo/bar/baz",
								},
							},
						},
					},
				},
			},
		},
		want: `exactly one of value or valueCommand must be specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Empty Container Config Image",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `image cannot be empty in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config SkipImagePull And IgnoreImagePullFailures Both Set To True",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image:                   "foo/bar:123",
						SkipImagePull:           true,
						IgnoreImagePullFailures: true,
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `ignoreImagePullFailures cannot be true when skipImagePull is true in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config SkipImagePull And PullImageBeforeStop Both Set To True",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image:               "foo/bar:123",
						SkipImagePull:       true,
						PullImageBeforeStop: true,
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `pullImageBeforeStop cannot be true when skipImagePull is true in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Empty Container Config Label Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Metadata: config.ContainerMetadata{
						Labels: []config.Label{
							{
								Value: "foo-bar",
							},
						},
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `empty label name in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Config Label Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Metadata: config.ContainerMetadata{
						Labels: []config.Label{
							{
								Name:  "FOO",
								Value: "foo-bar",
							},
							{
								Name:  "FOO2",
								Value: "foo-bar-2",
							},
							{
								Name:  "FOO",
								Value: "foo-bar-3",
							},
						},
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `label name FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Label Value",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Metadata: config.ContainerMetadata{
						Labels: []config.Label{
							{
								Name: "FOO",
							},
						},
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
				},
			},
		},
		want: `empty label value for label FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Order",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
				},
			},
		},
		want: `container order 0 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Negative Order",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: -1,
					},
				},
			},
		},
		want: `container order -1 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Restart Policy MaxRetryCount Set With Non-On-Failure Mode",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
						RestartPolicy: config.ContainerRestartPolicy{
							Mode:          "always",
							MaxRetryCount: 5,
						},
					},
				},
			},
		},
		want: `restart policy max retry count can be set only when the mode is on-failure in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Invalid Restart Policy Mode",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
						RestartPolicy: config.ContainerRestartPolicy{
							Mode: "garbage",
						},
					},
				},
			},
		},
		want: `invalid restart policy mode garbage in container {Group: g1 Container:c1} config, valid values are \[ 'no', 'always', 'on-failure', 'unless-stopped' \]`,
	},
	{
		name: "Container Config Negative Restart Policy MaxRetryCount",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
						RestartPolicy: config.ContainerRestartPolicy{
							Mode:          "on-failure",
							MaxRetryCount: -1,
						},
					},
				},
			},
		},
		want: `restart policy max retry count -1 cannot be negative in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Negative StopTimeout",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order:       1,
						StopTimeout: -1,
					},
				},
			},
		},
		want: `container stop timeout -1 cannot be negative in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config PrimaryUserGroup Without User",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					User: config.ContainerUser{
						PrimaryGroup: "my-user-group",
					},
				},
			},
		},
		want: `container user primary group cannot be set without setting the user in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Device Missing Src And Dst",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Devices: config.ContainerDevice{
							Static: []config.Device{
								{},
							},
						},
					},
				},
			},
		},
		want: `device src cannot be empty in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Device Missing Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Devices: config.ContainerDevice{
							Static: []config.Device{
								{
									Dst: "/dev/my-target-dev",
								},
							},
						},
					},
				},
			},
		},
		want: `device src cannot be empty in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Dynamic Device Invalid Command",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Devices: config.ContainerDevice{
							DynamicCommand: []string{
								"garbage-command",
								"garbage-arg1",
								"garbage-arg2",
								"garbage-arg3",
							},
						},
					},
				},
			},
		},
		want: `invalid fake executor command garbage-command \["garbage-arg1" "garbage-arg2" "garbage-arg3"\]`,
	},
	{
		name: "Container Config Empty Mount Name",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Type:     "bind",
								Src:      "/foo",
								Dst:      "/bar",
								ReadOnly: true,
							},
						},
					},
				},
			},
		},
		want: `mount name cannot be empty in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Duplicate Mounts Within Container Config",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "mount-foo1",
								Type: "bind",
								Src:  "/foo1",
								Dst:  "/bar1",
							},
							{
								Name: "mount-foo2",
								Type: "bind",
								Src:  "/foo2",
								Dst:  "/bar2",
							},
							{
								Name: "mount-foo1",
								Type: "bind",
								Src:  "/foo3",
								Dst:  "/bar3",
							},
						},
					},
				},
			},
		},
		want: `mount name mount-foo1 defined more than once in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Duplicate Mounts Within Container And Global Configs Combined",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "mount-foo2",
							Type: "bind",
							Src:  "/foo2-global",
							Dst:  "/bar2-global",
						},
						{
							Name: "mount-foo3",
							Type: "bind",
							Src:  "/foo3",
							Dst:  "/bar3",
						},
					},
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "mount-foo1",
								Type: "bind",
								Src:  "/foo1",
								Dst:  "/bar1",
							},
							{
								Name: "mount-foo2",
								Type: "bind",
								Src:  "/foo2",
								Dst:  "/bar2",
							},
						},
					},
				},
			},
		},
		want: `mount name mount-foo2 defined more than once in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount With Invalid Mount Type",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo",
								Type: "garbage",
								Src:  "/foo",
								Dst:  "/bar",
							},
						},
					},
				},
			},
		},
		want: `unsupported mount type garbage for mount foo in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount With Empty Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo",
								Type: "bind",
								Dst:  "/bar",
							},
						},
					},
				},
			},
		},
		want: `bind mount name foo cannot have an empty value for src in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Tmpfs Mount With Non-Empty Src",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo",
								Type: "tmpfs",
								Src:  "/foo",
								Dst:  "/bar",
							},
						},
					},
				},
			},
		},
		want: `tmpfs mount name foo cannot have a non-empty value for src in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount With Empty Dst",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo",
								Type: "bind",
								Src:  "/foo",
							},
						},
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for dst in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Bind Mount With Tmpfs Size",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name:      "foo",
								Type:      "bind",
								Src:       "/foo",
								Dst:       "/bar",
								TmpfsSize: 10000,
							},
						},
					},
				},
			},
		},
		want: `bind mount name foo cannot specify tmpfs size in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Tmpfs Mount With Negative Tmpfs Size",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name:      "foo",
								Type:      "tmpfs",
								Dst:       "/bar",
								TmpfsSize: -10000,
							},
						},
					},
				},
			},
		},
		want: `tmpfs mount name foo cannot specify a negative tmpfs size -10000 in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount Def Reference Not Found",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo2",
							},
						},
					},
				},
			},
		},
		want: `mount specified by just the name foo2 not found in defs in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Duplicate Mount Name Only With Global Container Config Mount",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo1",
							Type: "bind",
							Src:  "/foo1",
							Dst:  "/bar1",
						},
						{
							Name: "foo2",
							Type: "bind",
							Src:  "/foo2",
							Dst:  "/bar2",
						},
					},
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo2",
							},
						},
					},
				},
			},
		},
		want: `mount name foo2 defined more than once in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Duplicate Mount Name And Def With Global Container Config Mount",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "foo1",
							Type: "bind",
							Src:  "/foo1",
							Dst:  "/bar1",
						},
						{
							Name: "foo2",
							Type: "bind",
							Src:  "/foo2",
							Dst:  "/bar2",
						},
					},
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "foo2",
								Type: "bind",
								Src:  "/foo-something-else",
								Dst:  "/bar-something-else",
							},
						},
					},
				},
			},
		},
		want: `mount name foo2 defined more than once in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Published Port - Container Port Empty",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								Protocol: "tcp",
								HostIP:   "127.0.0.1",
								HostPort: "5001",
							},
						},
					},
				},
			},
		},
		want: `unable to convert published container port  to an integer, reason: strconv.ParseInt: parsing "": invalid syntax`,
	},
	{
		name: "Container Config Published Port - Container Port Non Integer",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "-6365foobar9563",
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      "5001",
							},
						},
					},
				},
			},
		},
		want: `unable to convert published container port -6365foobar9563 to an integer, reason: strconv.ParseInt: parsing "-6365foobar9563": invalid syntax`,
	},
	{
		name: "Container Config Published Port - Container Port Negative",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "-1",
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      "5001",
							},
						},
					},
				},
			},
		},
		want: `published container port -1 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Protocol Empty",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								HostIP:        "127.0.0.1",
								HostPort:      "5001",
							},
						},
					},
				},
			},
		},
		want: `published container port 10001 specifies an invalid protocol  in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Protocol Invalid",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								Protocol:      "garbage",
								HostIP:        "127.0.0.1",
								HostPort:      "5001",
							},
						},
					},
				},
			},
		},
		want: `published container port 10001 specifies an invalid protocol garbage in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Host IP Empty",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								Protocol:      "tcp",
								HostPort:      "5001",
							},
						},
					},
				},
			},
		},
		want: `published host IP cannot be empty for container port 10001 in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Host IP Invalid",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								Protocol:      "tcp",
								HostIP:        "abc.def.ghi.jkl",
								HostPort:      "5001",
							},
						},
					},
				},
			},
		},
		want: `published host IP abc\.def\.ghi\.jkl for container port 10001 is invalid in container {Group: g1 Container:c1} config, reason: ParseAddr\("abc\.def\.ghi\.jkl"\): unexpected character \(at "abc\.def\.ghi\.jkl"\)`,
	},
	{
		name: "Container Config Published Port - Host Port Empty",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
							},
						},
					},
				},
			},
		},
		want: `unable to convert published host port  to an integer, reason: strconv.ParseInt: parsing "": invalid syntax`,
	},
	{
		name: "Container Config Published Port - Host Port Non Integer",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      "9876garbage0123",
							},
						},
					},
				},
			},
		},
		want: `unable to convert published host port 9876garbage0123 to an integer, reason: strconv.ParseInt: parsing "9876garbage0123": invalid syntax`,
	},
	{
		name: "Container Config Published Port - Host Port Negative",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Network: config.ContainerNetwork{
						PublishedPorts: []config.PublishedPort{
							{
								ContainerPort: "10001",
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      "-1",
							},
						},
					},
				},
			},
		},
		want: `published host port -1 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Empty Container Config Sysctl Key",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Security: config.ContainerSecurity{
						Sysctls: []config.Sysctl{
							{
								Value: "foo-bar",
							},
						},
					},
				},
			},
		},
		want: `empty sysctl key in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Config Sysctl Key",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Security: config.ContainerSecurity{
						Sysctls: []config.Sysctl{
							{
								Key:   "FOO",
								Value: "foo-bar",
							},
							{
								Key:   "FOO2",
								Value: "foo-bar-2",
							},
							{
								Key:   "FOO",
								Value: "foo-bar-3",
							},
						},
					},
				},
			},
		},
		want: `sysctl key FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Sysctl Value",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Security: config.ContainerSecurity{
						Sysctls: []config.Sysctl{
							{
								Key: "FOO",
							},
						},
					},
				},
			},
		},
		want: `empty sysctl value for sysctl FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Health Config - Negative Retries",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Health: config.ContainerHealth{
						Retries: -1,
					},
				},
			},
		},
		want: `health check retries -1 cannot be negative in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Health Config - Invalid Interval",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Health: config.ContainerHealth{
						Interval: "garbage",
					},
				},
			},
		},
		want: `health check interval garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container Health Config - Invalid Timeout",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Health: config.ContainerHealth{
						Timeout: "garbage",
					},
				},
			},
		},
		want: `health check timeout garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container Health Config - Invalid StartPeriod",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Health: config.ContainerHealth{
						StartPeriod: "garbage",
					},
				},
			},
		},
		want: `health check start period garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container Health Config - Invalid StartInterval",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Health: config.ContainerHealth{
						StartInterval: "garbage",
					},
				},
			},
		},
		want: `health check start interval garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container ShmSize Invalid Unit",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Runtime: config.ContainerRuntime{
						ShmSize: "1foobar",
					},
				},
			},
		},
		want: `invalid shmSize 1foobar in container {Group: g1 Container:c1} config, reason: invalid suffix: 'foobar'`,
	},
	{
		name: "Container ShmSize Invalid Value",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Runtime: config.ContainerRuntime{
						ShmSize: "garbage",
					},
				},
			},
		},
		want: `invalid shmSize garbage in container {Group: g1 Container:c1} config, reason: invalid size: 'garbage'`,
	},
	{
		name: "Empty Container Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Runtime: config.ContainerRuntime{
						Env: []config.ContainerEnv{
							{
								Value: "foo-bar",
							},
						},
					},
				},
			},
		},
		want: `empty env var in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Env Var",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Runtime: config.ContainerRuntime{
						Env: []config.ContainerEnv{
							{
								Var:   "FOO",
								Value: "foo-bar",
							},
							{
								Var:   "FOO2",
								Value: "foo-bar-2",
							},
							{
								Var:   "FOO",
								Value: "foo-bar-3",
							},
						},
					},
				},
			},
		},
		want: `env var FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Env Var Without Value",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "foo/bar:123",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Runtime: config.ContainerRuntime{
						Env: []config.ContainerEnv{
							{
								Var: "FOO",
							},
						},
					},
				},
			},
		},
		want: `value not specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
}

func TestBuildDeploymentFromConfigErrors(t *testing.T) {
	t.Parallel()

	for _, test := range buildDeploymentFromConfigErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, gotErr := FromConfig(testutils.NewVanillaTestContext(), &tc.config)
			if gotErr == nil {
				testhelpers.LogErrorNil(t, "Homelab.validate()", tc.name, tc.want)
				return
			}

			if !testhelpers.RegexMatch(t, "Homelab.validate()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}
