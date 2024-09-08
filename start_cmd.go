package main

import (
	"flag"
)

type startCmdHandler struct {
	cgFlags containerAndGroupFlags
	config  HomelabConfig
}

func newStartCmdHandler() *startCmdHandler {
	return &startCmdHandler{}
}

func (s *startCmdHandler) updateFlagSet(fs *flag.FlagSet) {
	addContainerAndGroupFlags(fs, &s.cgFlags)
}

func (s *startCmdHandler) run() error {
	err := parseHomelabConfig(&s.config)
	if err != nil {
		return err
	}

	log.Infof("allGroups: %t", s.cgFlags.allGroups)
	log.Infof("group: %s", s.cgFlags.group)
	log.Infof("container: %s", s.cgFlags.container)

	// Identify the containers that are in scope for this command invocation.
	// Run start() against each of those containers.

	// start() for a single container involves these steps:
	// 1. Validate the container is allowed to run on the current host.
	// 2. Create the network for the container if it doesn't exist already.
	// 3. Execute any pre-start commands.
	// 4. Pull the container image.
	// 5. Purge (i.e. stop and remove) any previously existing containers
	// under the same name.
	// 6. Create the container.
	// 7. Start the container.

	_, _ = containersForCmd(&s.config, s.cgFlags.allGroups, s.cgFlags.group, s.cgFlags.container)

	return nil
}

func containersForCmd(config *HomelabConfig, allGroups bool, group string, container string) ([]ContainerConfig, error) {
	return nil, nil
}
