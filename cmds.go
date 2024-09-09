package main

import (
	"flag"
	"fmt"
)

var (
	showConfigCmd = "showconfig"
	startCmd      = "start"
	cmdHandlers   = map[string]cmd{
		showConfigCmd: newShowConfigCmdHandler(),
		startCmd:      newStartCmdHandler(),
	}
)

type cmdOptions struct {
	containerAndGroup containerAndGroupFlags
}

type containerAndGroupFlags struct {
	allGroups bool
	group     string
	container string
}

type cmd interface {
	containerAndGroupFlags() bool
	run(options *cmdOptions) error
}

func handleSubCommand() error {
	args := flag.Args()
	subCmd := args[0]
	handler, ok := cmdHandlers[subCmd]
	if !ok {
		return fmt.Errorf("Invalid command: %s", subCmd)
	}

	// Configure flags for the sub-command and parse the command
	// line flags and args.
	options := cmdOptions{}
	fs := flag.NewFlagSet(subCmd, flag.ExitOnError)
	if handler.containerAndGroupFlags() {
		addContainerAndGroupFlags(fs, &options.containerAndGroup)
	}
	err := fs.Parse(args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse flags for %s command, reason: %w", subCmd, err)
	}

	return handler.run(&options)
}

func addContainerAndGroupFlags(fs *flag.FlagSet, c *containerAndGroupFlags) {
	fs.BoolVar(&c.allGroups, "allGroups", false, "Whether to apply this command across all groups and containers within")
	fs.StringVar(&c.group, "group", "", "The container group to apply this command")
	fs.StringVar(&c.container, "container", "", "The container to apply this command")
}

func validateContainerAndGroupFlags(c *containerAndGroupFlags) error {
	if c.allGroups {
		if c.group != "" {
			return fmt.Errorf("-group flag must not be passed when -allGroups=true")
		}
		if c.container != "" {
			return fmt.Errorf("-container flag must not be passed when -allGroups=true")
		}
	}
	if c.container != "" && c.group == "" {
		return fmt.Errorf("-group flag must also be passed when -container flag is passed")
	}
	return nil
}
