package main

import (
	"flag"
)

type startCmdHandler struct {
	cgFlags containerAndGroupFlags
	dep     *deployment
}

func newStartCmdHandler() *startCmdHandler {
	return &startCmdHandler{}
}

func (s *startCmdHandler) updateFlagSet(fs *flag.FlagSet) {
	addContainerAndGroupFlags(fs, &s.cgFlags)
}

func (s *startCmdHandler) run() error {
	c := HomelabConfig{}
	err := parseHomelabConfig(&c)
	if err != nil {
		return err
	}
	s.dep = newDeployment(&c)

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

	var containers containerList
	if s.cgFlags.allGroups {
		containers = s.dep.queryAllContainers()
	} else if s.cgFlags.group != "" {
		if s.cgFlags.container == "" {
			containers = s.dep.queryAllContainersInGroup(s.cgFlags.group)
		} else {
			containers = append(containers, s.dep.queryContainer(s.cgFlags.group, s.cgFlags.container))
		}
	}

	log.Infof("Result containers =\n%s", containers)

	return nil
}
