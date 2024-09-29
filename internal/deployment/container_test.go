package deployment

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/docker/fakedocker"
	"github.com/tuxdudehomelab/homelab/internal/inspect"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

var containerStartTests = []struct {
	name           string
	config         config.HomelabConfig
	cRef           config.ContainerReference
	ctxInfo        *testutils.TestContextInfo
	preExec        func(context.Context)
	wantNotStarted bool
}{
	{
		name: "Container Start - Doesn't Exist Already",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - Existing Image",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ExistingImages: utils.StringSet{
					"abc/xyz": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - No Network Endpoints",
		config: buildSingleContainerNoNetworkConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Created State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateCreated,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Paused State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Restarting State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Removing State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRemoving,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		preExec: func(ctx context.Context) {
			go func() {
				time.Sleep(200 * time.Millisecond)
				d := fakedocker.FakeDockerHostFromContext(ctx)
				err := d.ForceRemoveContainer("g1-c1")
				if err != nil {
					panic(err)
				}
			}()
		},
	},
	{
		name: "Container Start - Exists Already In Exited State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateExited,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Dead State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateDead,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State Requiring Multiple Stops",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 5,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State Requiring Multiple Kills",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
						RequiredExtraKills: 4,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - Container Mode Network",
		config: buildSingleContainerWithContainerModeNetworkConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			config.ContainerReference{
				Group:     "g1",
				Container: "c2",
			}),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
}

func TestContainerStart(t *testing.T) {
	// Reduce this delay to keep the tests executing quickly.
	stopAndRemoveKillDelay = 100 * time.Millisecond
	for _, test := range containerStartTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			// Enable debug inspect level while running the container start tests
			// for extra code coverage.
			if tc.ctxInfo.InspectLevel == inspect.HomelabInspectLevelNone {
				tc.ctxInfo.InspectLevel = inspect.HomelabInspectLevelDebug
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dockerClient, gotErr := docker.NewDockerClient(ctx, dep.Host.DockerPlatform, dep.Host.Arch)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "NewDockerClient()", tc.name, gotErr)
			}
			defer dockerClient.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			if tc.preExec != nil {
				tc.preExec(ctx)
			}

			gotStarted, gotErr := ct.Start(ctx, dockerClient)
			if gotErr != nil {
				testhelpers.LogErrorNotNilWithOutput(t, "container.start()", tc.name, buf, gotErr)
				return
			}
			wantStarted := !tc.wantNotStarted
			if gotStarted != wantStarted {
				testhelpers.LogCustomWithOutput(t, "container.start()", tc.name, buf, fmt.Sprintf("gotStarted (%t) != wantStarted (%t)", gotStarted, wantStarted))
			}

			d := fakedocker.FakeDockerHostFromContext(ctx)
			gotState := d.GetContainerState(fmt.Sprintf("%s-%s", tc.cRef.Group, tc.cRef.Container))
			if gotState != docker.ContainerStateRunning {
				testhelpers.LogCustomWithOutput(t, "Container state after container.start()", tc.name, buf, fmt.Sprintf("gotState (%s) != ContainerStateRunning", gotState))
			}
		})
	}
}

var containerStartErrorTests = []struct {
	name      string
	config    config.HomelabConfig
	cRef      config.ContainerReference
	ctxInfo   *testutils.TestContextInfo
	wantPanic bool
	want      string
}{
	{
		name: "Container Start - Image Not Available",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Failed to start container g1-c1, reason:failed to pull the image abc/xyz, reason: image abc/xyz not found or invalid and cannot be pulled by the fake docker host`,
	},
	{
		name: "Container Start - Image Pull Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed while pulling the image abc/xyz, reason: failed to pull image abc/xyz on the fake docker host`,
	},
	{
		name: "Container Start - No Local Image After Pull",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				NoImageAfterPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:image abc/xyz not available locally after a successful pull, possibly indicating a bug or a system failure!`,
	},
	{
		name: "Container Start - Kill Existing Container Fails",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
					},
				},
				FailContainerKill: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to stop and remove container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Start - Unkillable Existing Container",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
						RequiredExtraKills: 5,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to stop and remove container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Start - Container State Unknown",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateUnknown,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		wantPanic: true,
		want:      `container g1-c1 is in an unsupported state Unknown, possibly indicating a bug in the code`,
	},
	{
		name: "Container Start - Inspect Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
				},
				FailContainerInspect: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to retrieve the container state, reason: failed to inspect container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Stop Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
				},
				FailContainerStop: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Remove Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
				FailContainerRemove: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to remove the container, reason: failed to remove container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Create Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateDead,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailContainerCreate: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the container, reason: failed to create container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Start Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailContainerStart: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to start the container, reason: failed to start container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Primary Network Create Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailNetworkCreate: utils.StringSet{
					"g1-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the network, reason: failed to create network g1-bridge on the fake docker host`,
	},
	{
		name: "Container Start - Secondary Network Create Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailNetworkCreate: utils.StringSet{
					"proxy-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the network, reason: failed to create network proxy-bridge on the fake docker host`,
	},
	{
		name: "Container Start - Secondary Network Connect Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailNetworkConnect: utils.StringSet{
					"proxy-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to connect container g1-c1 to network proxy-bridge, reason: failed to connect container g1-c1 to network proxy-bridge on the fake docker host`,
	},
}

func TestContainerStartErrors(t *testing.T) {
	for _, test := range containerStartErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			ctx := testutils.NewTestContext(tc.ctxInfo)

			if tc.wantPanic {
				defer testhelpers.ExpectPanicWithOutput(t, "container.start()", tc.name, buf, tc.want)
			}

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dockerClient, gotErr := docker.NewDockerClient(ctx, dep.Host.DockerPlatform, dep.Host.Arch)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "NewDockerClient()", tc.name, gotErr)
				return
			}
			defer dockerClient.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			gotStarted, gotErr := ct.Start(ctx, dockerClient)
			if gotErr == nil {
				testhelpers.LogErrorNilWithOutput(t, "container.start()", tc.name, buf, tc.want)
				return
			}
			if gotStarted {
				testhelpers.LogCustomWithOutput(t, "container.start()", tc.name, buf, "gotStarted (true) != wantStarted (false)")
				return
			}

			if !testhelpers.RegexMatchWithOutput(t, "container.start()", tc.name, buf, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

func buildSingleContainerConfig(ct config.ContainerReference, image string) config.HomelabConfig {
	conf := buildSingleContainerNoNetworkConfig(ct, image)
	conf.IPAM = config.IPAMConfig{
		Networks: config.NetworksConfig{
			BridgeModeNetworks: []config.BridgeModeNetworkConfig{
				{
					Name:              fmt.Sprintf("%s-bridge", ct.Group),
					HostInterfaceName: fmt.Sprintf("docker-%s", ct.Group),
					CIDR:              "172.18.101.0/24",
					Priority:          1,
					Containers: []config.ContainerIPConfig{
						{
							IP:        "172.18.101.11",
							Container: ct,
						},
					},
				},
				{
					Name:              "proxy-bridge",
					HostInterfaceName: "docker-prx",
					CIDR:              "172.18.201.0/24",
					Priority:          2,
					Containers: []config.ContainerIPConfig{
						{
							IP:        "172.18.201.11",
							Container: ct,
						},
					},
				},
			},
		},
	}
	return conf
}

func buildSingleContainerWithContainerModeNetworkConfig(ct config.ContainerReference, image string, connectTo config.ContainerReference) config.HomelabConfig {
	conf := buildSingleContainerNoNetworkConfig(ct, image)
	conf.IPAM = config.IPAMConfig{
		Networks: config.NetworksConfig{
			ContainerModeNetworks: []config.ContainerModeNetworkConfig{
				{
					Name:      "net1",
					Container: connectTo,
					AttachingContainers: []config.ContainerReference{
						ct,
					},
				},
			},
		},
	}
	return conf
}

func buildSingleContainerNoNetworkConfig(ct config.ContainerReference, image string) config.HomelabConfig {
	return config.HomelabConfig{
		Global: config.GlobalConfig{
			BaseDir: testhelpers.HomelabBaseDir(),
		},
		Hosts: []config.HostConfig{
			{
				Name: "fakehost",
				AllowedContainers: []config.ContainerReference{
					ct,
				},
			},
		},
		Groups: []config.ContainerGroupConfig{
			{
				Name:  ct.Group,
				Order: 1,
			},
		},
		Containers: []config.ContainerConfig{
			{
				Info: ct,
				Image: config.ContainerImageConfig{
					Image: image,
				},
				Lifecycle: config.ContainerLifecycleConfig{
					Order: 1,
				},
			},
		},
	}
}