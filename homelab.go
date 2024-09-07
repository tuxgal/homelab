// Command homelab is used to manage the configuration, deployment, and
// orchestration of a group of docker containers on a given host.
package main

import (
	"flag"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

var (
	log = buildLogger()
)

func buildLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.MaxLevel = zzzlog.LvlInfo
	return zzzlog.NewLogger(config)
}

func run() int {
	log.Infof("Hello from homelab!")
	return 0
}

func main() {
	flag.Parse()
	os.Exit(run())
}
