package main

import (
	"flag"
)

type startCmdHandler struct {
	cgFlags containerAndGroupFlags
}

func newStartCmdHandler() *startCmdHandler {
	return &startCmdHandler{}
}

func (s *startCmdHandler) updateFlagSet(fs *flag.FlagSet) {
	addContainerAndGroupFlags(fs, &s.cgFlags)
}

func (s *startCmdHandler) run() error {
	_, err := parseHomelabConfig()
	if err != nil {
		return err
	}

	log.Infof("allGroups: %t", s.cgFlags.allGroups)
	log.Infof("group: %s", s.cgFlags.group)
	log.Infof("container: %s", s.cgFlags.container)
	return nil
}
