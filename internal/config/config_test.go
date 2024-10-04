package config

import (
	"bytes"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/config/env"
	"github.com/tuxdudehomelab/homelab/internal/deepcopy"
	logger "github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
)

var applyConfigEnvToContainerTests = []struct {
	name              string
	container         Container
	globalEnvMap      env.EnvMap
	globalEnvOrder    env.EnvOrder
	containerEnvMap   env.EnvMap
	containerEnvOrder env.EnvOrder
	want              Container
}{
	{
		name: "Container Config - ApplyConfigEnv - Exhaustive",
		container: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order:        1,
				StartPreHook: "$$CONTAINER_SCRIPTS_DIR$$/my-start-prehook.sh",
			},
			User: ContainerUser{
				User:         "$$USER_ID$$",
				PrimaryGroup: "$$USER_PRIMARY_GROUP_ID$$",
				AdditionalGroups: []string{
					"$$ENV_GROUP_FOO$$",
					"$$ENV_GROUP_BAR$$",
				},
			},
			Filesystem: ContainerFilesystem{
				ReadOnlyRootfs: true,
				Mounts: []Mount{
					{
						Name: "mount-1",
						Type: "bind",
						Src:  "$$CONTAINER_CONFIGS_DIR$$/config.yml",
						Dst:  "/data/$$USER_NAME$$/config.yml",
					},
					{
						Name: "some-other-mount",
						Type: "bind",
						Src:  "$$CONTAINER_DATA_DIR$$/my-data",
						Dst:  "$$ENV_MOUNT_DEST_DIR$$/my-data",
					},
					{
						Name: "some-other-mount-2",
						Type: "bind",
						Src:  "$$CONTAINER_BASE_DIR$$/abc/xyz",
						Dst:  "/abc/xyz",
					},
					{
						Name:    "tmpfs-mount",
						Type:    "tmpfs",
						Dst:     "/tmp/cache-$$USER_PRIMARY_GROUP_NAME$$",
						Options: "tmpfs-size=$$ENV_TMPFS_SIZE$$",
					},
				},
				Devices: []Device{
					{
						Src: "$$ENV_SRC_DEV$$",
						Dst: "$$ENV_DST_DEV$$",
					},
				},
			},
			Network: ContainerNetwork{
				HostName:   "MyHost-$$HOST_NAME$$",
				DomainName: "$$ENV_DOMAIN$$",
				DNSServers: []string{
					"1.1.1.1",
					"$$ENV_DNS_SERVER$$",
				},
				DNSOptions: []string{
					"dns-option-1",
					"$$ENV_DNS_OPTION$$",
				},
				DNSSearch: []string{
					"dns-ct-search-1",
					"$$ENV_DNS_SEARCH$$",
				},
				PublishedPorts: []PublishedPort{
					{
						ContainerPort: 12345,
						Protocol:      "tcp",
						HostIP:        "$$HOST_IP$$",
						HostPort:      678,
					},
				},
			},
			Runtime: ContainerRuntime{
				Env: []ContainerEnv{
					{
						Var:   "$$ENV_VAR_1$$",
						Value: "$$ENV_VAR_1_VAL$$",
					},
					{
						Var:          "MY_ENV_2",
						ValueCommand: "$$ENV_VAR_2_VAL_CMD$$",
					},
					{
						Var:   "MY_ENV_3",
						Value: "SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName",
					},
				},
			},
		},
		globalEnvMap: env.EnvMap{
			"ENV_GROUP_FOO":      "groupfoo",
			"ENV_GROUP_BAR":      "groupbar",
			"ENV_MOUNT_DEST_DIR": "/mnt/dst1",
			"ENV_TMPFS_SIZE":     "100000000",
			"ENV_SRC_DEV":        "/dev/src",
			"ENV_DST_DEV":        "/dev/dst",
			"ENV_DOMAIN":         "my-domain",
			"ENV_DNS_SERVER":     "10.11.11.11",
			"ENV_DNS_OPTION":     "dns-option-2",
			"ENV_DNS_SEARCH":     "dns-ct-search-2",
		},
		globalEnvOrder: env.EnvOrder{
			"ENV_GROUP_FOO",
			"ENV_GROUP_BAR",
			"ENV_MOUNT_DEST_DIR",
			"ENV_TMPFS_SIZE",
			"ENV_SRC_DEV",
			"ENV_DST_DEV",
			"ENV_DOMAIN",
			"ENV_DNS_SERVER",
			"ENV_DNS_OPTION",
			"ENV_DNS_SEARCH",
		},
		containerEnvMap: env.EnvMap{
			"ENV_VAR_1":         "MY_ENV_1",
			"ENV_VAR_1_VAL":     "my-env-1-val",
			"ENV_VAR_2_VAL_CMD": "cat /foo/bar.txt",
		},
		containerEnvOrder: env.EnvOrder{
			"ENV_VAR_1",
			"ENV_VAR_1_VAL",
			"ENV_VAR_2_VAL_CMD",
		},
		want: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order:        1,
				StartPreHook: "/tmp/base-dir/g1/c1/scripts/my-start-prehook.sh",
			},
			User: ContainerUser{
				User:         "55555",
				PrimaryGroup: "44444",
				AdditionalGroups: []string{
					"groupfoo",
					"groupbar",
				},
			},
			Filesystem: ContainerFilesystem{
				ReadOnlyRootfs: true,
				Mounts: []Mount{
					{
						Name: "mount-1",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/c1/configs/config.yml",
						Dst:  "/data/fakeuser/config.yml",
					},
					{
						Name: "some-other-mount",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/c1/data/my-data",
						Dst:  "/mnt/dst1/my-data",
					},
					{
						Name: "some-other-mount-2",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/c1/abc/xyz",
						Dst:  "/abc/xyz",
					},
					{
						Name:    "tmpfs-mount",
						Type:    "tmpfs",
						Dst:     "/tmp/cache-fakegroup1",
						Options: "tmpfs-size=100000000",
					},
				},
				Devices: []Device{
					{
						Src: "/dev/src",
						Dst: "/dev/dst",
					},
				},
			},
			Network: ContainerNetwork{
				HostName:   "MyHost-fakehost",
				DomainName: "my-domain",
				DNSServers: []string{
					"1.1.1.1",
					"10.11.11.11",
				},
				DNSOptions: []string{
					"dns-option-1",
					"dns-option-2",
				},
				DNSSearch: []string{
					"dns-ct-search-1",
					"dns-ct-search-2",
				},
				PublishedPorts: []PublishedPort{
					{
						ContainerPort: 12345,
						Protocol:      "tcp",
						HostIP:        "10.76.77.78",
						HostPort:      678,
					},
				},
			},
			Runtime: ContainerRuntime{
				Env: []ContainerEnv{
					{
						Var:   "MY_ENV_1",
						Value: "my-env-1-val",
					},
					{
						Var:          "MY_ENV_2",
						ValueCommand: "cat /foo/bar.txt",
					},
					{
						Var:   "MY_ENV_3",
						Value: "SomeHostName.FakeHost.SomeDomainName",
					},
				},
			},
		},
	},
}

func TestApplyConfigEnvToContainer(t *testing.T) {
	for _, test := range applyConfigEnvToContainerTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			e := env.NewSystemConfigEnvManager(ctx)
			e = e.NewGlobalConfigEnvManager(ctx, "/tmp/base-dir", tc.globalEnvMap, tc.globalEnvOrder)
			e = e.NewContainerConfigEnvManager(ctx, "/tmp/base-dir/g1/c1", tc.containerEnvMap, tc.containerEnvOrder)

			got := deepcopy.MustCopy(tc.container)
			got.ApplyConfigEnv(e)
			if !testhelpers.CmpDiff(t, "Container.ApplyConfigEnv()", tc.name, "apply result", tc.want, got) {
				return
			}
		})
	}
}