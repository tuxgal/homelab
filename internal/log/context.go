package log

import (
	"context"

	"github.com/tuxgal/tuxlogi"
)

var (
	loggerKey = ctxKeyLogger{}
)

type ctxKeyLogger struct{}

func Log(ctx context.Context) tuxlogi.Logger {
	if logger, found := ctx.Value(loggerKey).(tuxlogi.Logger); found {
		return logger
	}
	panic("Unable to retrieve logger from context")
}

func WithLogger(ctx context.Context, logger tuxlogi.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
