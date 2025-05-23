package env

import (
	"context"
	"fmt"

	"github.com/tuxgal/homelab/internal/host"
	"github.com/tuxgal/homelab/internal/user"
)

var (
	configEnvHostIPV4              = "HOST_IPV4"
	configEnvHostName              = "HOST_NAME"
	configEnvHumanFriendlyHostName = "HUMAN_FRIENDLY_HOST_NAME"
	configEnvUserName              = "USER_NAME"
	configEnvUserID                = "USER_ID"
	configEnvUserPrimaryGroupName  = "USER_PRIMARY_GROUP_NAME"
	configEnvUserPrimaryGroupID    = "USER_PRIMARY_GROUP_ID"
	configEnvHomelabBaseDir        = "HOMELAB_BASE_DIR"
	configEnvContainerGroupBaseDir = "CONTAINER_GROUP_BASE_DIR"
	configEnvContainerBaseDir      = "CONTAINER_BASE_DIR"
	configEnvContainerConfigsDir   = "CONTAINER_CONFIGS_DIR"
	configEnvContainerDatasDir     = "CONTAINER_DATA_DIR"
	configEnvContainerScriptsDir   = "CONTAINER_SCRIPTS_DIR"
)

type ConfigEnvManager struct {
	env *configEnv
}

type EnvMap map[string]string
type EnvOrder []string

func NewSystemConfigEnvManager(ctx context.Context) *ConfigEnvManager {
	envMap, envOrder := defaultEnv(ctx)
	return &ConfigEnvManager{
		env: newConfigEnv(ctx, envMap, envOrder),
	}
}

func (c *ConfigEnvManager) NewGlobalConfigEnvManager(ctx context.Context, homelabBaseDir string, env EnvMap, order EnvOrder) *ConfigEnvManager {
	// Apply standard env variables specific to the config that are
	// relevant globally within the config.
	newEnv := c.env.override(
		ctx,
		EnvMap{
			configEnvHomelabBaseDir: homelabBaseDir,
		},
		EnvOrder{
			configEnvHomelabBaseDir,
		},
	)
	// Apply other env variables which were read from the global config.
	return &ConfigEnvManager{
		env: newEnv.override(ctx, env, order),
	}
}

func (c *ConfigEnvManager) NewContainerConfigEnvManager(ctx context.Context, containerGroupBaseDir, containerBaseDir string, env EnvMap, order EnvOrder) *ConfigEnvManager {
	// Apply env variables specific to this container that are relevant
	// within the container config.
	newEnv := c.env.override(
		ctx,
		EnvMap{
			configEnvContainerGroupBaseDir: containerGroupBaseDir,
			configEnvContainerBaseDir:      containerBaseDir,
			configEnvContainerConfigsDir:   containerConfigsDir(containerBaseDir),
			configEnvContainerDatasDir:     containerDataDir(containerBaseDir),
			configEnvContainerScriptsDir:   containerScriptsDir(containerBaseDir),
		},
		EnvOrder{
			configEnvContainerGroupBaseDir,
			configEnvContainerBaseDir,
			configEnvContainerConfigsDir,
			configEnvContainerDatasDir,
			configEnvContainerScriptsDir,
		},
	)
	return &ConfigEnvManager{
		env: newEnv.override(ctx, env, order),
	}
}

func (c *ConfigEnvManager) Apply(input string) string {
	return c.env.apply(input)
}

func defaultEnv(ctx context.Context) (EnvMap, EnvOrder) {
	h := host.MustHostInfo(ctx)
	u := user.MustUserInfo(ctx)
	return EnvMap{
			configEnvHostIPV4:              h.IPV4.String(),
			configEnvHostName:              h.HostName,
			configEnvHumanFriendlyHostName: h.HumanFriendlyHostName,
			configEnvUserName:              u.User.Username,
			configEnvUserID:                u.User.Uid,
			configEnvUserPrimaryGroupName:  u.PrimaryGroup.Name,
			configEnvUserPrimaryGroupID:    u.PrimaryGroup.Gid,
		}, EnvOrder{
			configEnvHostIPV4,
			configEnvHostName,
			configEnvHumanFriendlyHostName,
			configEnvUserName,
			configEnvUserID,
			configEnvUserPrimaryGroupName,
			configEnvUserPrimaryGroupID,
		}
}

func containerConfigsDir(containerBaseDir string) string {
	return fmt.Sprintf("%s/configs", containerBaseDir)
}

func containerDataDir(containerBaseDir string) string {
	return fmt.Sprintf("%s/data", containerBaseDir)
}

func containerScriptsDir(containerBaseDir string) string {
	return fmt.Sprintf("%s/scripts", containerBaseDir)
}
