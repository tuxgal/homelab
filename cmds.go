package main

import (
	"flag"
	"fmt"
)

var (
	showConfigCmd = "showconfig"
	cmdHandlers   = map[string]cmd{
		showConfigCmd: newShowConfigCmdHandler(),
	}
)

type cmd interface {
	updateFlagSet(fs *flag.FlagSet)
	run() error
}

func handleSubCommand() error {
	args := flag.Args()
	scmd := args[0]
	handler, ok := cmdHandlers[scmd]
	if !ok {
		return fmt.Errorf("Invalid command: %s", scmd)
	}

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	handler.updateFlagSet(fs)
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse flags for %s command, reason: %w", scmd, err)
	}
	return handler.run()
}
