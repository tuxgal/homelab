// Command homelab is used to manage the configuration, deployment, and
// orchestration of a group of docker containers on a given host.
package main

import (
	"flag"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

const (
	defaultCLIConfigPathFormat = "%s/.homelab/config.yaml"
	logLevelEnvVar             = "HOMELAB_LOG_LEVEL"
)

var (
	log = buildLogger()
)

func buildLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	envLogLevel, isEnvVarSet := os.LookupEnv(logLevelEnvVar)
	if isEnvVarSet && envLogLevel == "trace" {
		config.MaxLevel = zzzlog.LvlTrace
	} else if isEnvVarSet && envLogLevel == "debug" {
		config.MaxLevel = zzzlog.LvlDebug
	} else {
		config.MaxLevel = zzzlog.LvlInfo
	}
	return zzzlog.NewLogger(config)
}

func run() int {
	if !validateFlags() {
		return 1
	}
	err := handleSubCommand()
	if err != nil {
		log.Errorf("%s", err)
		return 1
	}
	return 0
}

func main() {
	flag.Parse()
	os.Exit(run())
}
