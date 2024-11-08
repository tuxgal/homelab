package testutils

import (
	"io"

	"github.com/tuxgal/tuxlog"
	"github.com/tuxgal/tuxlogi"
)

func NewTestLogger() tuxlogi.Logger {
	config := tuxlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	return tuxlog.NewLogger(config)
}

func NewCapturingTestLogger(lvl tuxlog.Level, w io.Writer) tuxlogi.Logger {
	config := tuxlog.NewConsoleLoggerConfig()
	config.MaxLevel = lvl
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = w
	return tuxlog.NewLogger(config)
}

func NewCapturingVanillaTestLogger(lvl tuxlog.Level, w io.Writer) tuxlogi.Logger {
	config := tuxlog.NewVanillaLoggerConfig()
	config.MaxLevel = lvl
	config.Dest = w
	config.PanicInFatal = true
	return tuxlog.NewLogger(config)
}
