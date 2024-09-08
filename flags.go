package main

import (
	"flag"
)

var (
	cliConfig     = flag.String("cliConfig", "", "The path to the Homelab CLI config")
	cliConfigFlag = flag.Lookup("cliConfig")

	homelabConfigsDir     = flag.String("configsDir", "", "The path to the directory containing the homelab configs")
	homelabConfigsDirFlag = flag.Lookup("configsDir")
)

// Returns true if a flag was passed in the command line invocation.
func isFlagPassed(target *flag.Flag) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == target.Name {
			found = true
		}
	})
	return found
}

func validateFlags() bool {
	if isFlagPassed(cliConfigFlag) && isFlagPassed(homelabConfigsDirFlag) {
		log.Fatalf("Both -cliConfig and -configsDir flags cannot be specified at the same time")
		return false
	}
	return true
}

type containerAndGroupFlags struct {
	allGroups bool
	group     string
	container string
}

func addContainerAndGroupFlags(fs *flag.FlagSet, c *containerAndGroupFlags) {
	fs.BoolVar(&c.allGroups, "allGroups", false, "Whether to apply this command across all groups and containers within")
	fs.StringVar(&c.group, "group", "", "The container group to apply this command")
	fs.StringVar(&c.container, "container", "", "The container to apply this command")
}
