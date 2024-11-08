package cmdexec

import (
	"bytes"
	"context"
	"testing"

	l "github.com/tuxgal/homelab/internal/log"
	"github.com/tuxgal/homelab/internal/testhelpers"
	"github.com/tuxgal/tuxlog"
	"github.com/tuxgal/tuxlogi"
)

func TestRetrieveExecutorFromEmptyContext(t *testing.T) {
	t.Parallel()

	tc := "Retrieve Executor - Empty Context"
	want := `Executor not found in context`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctx = l.WithLogger(ctx, newTestLogger())

		defer testhelpers.ExpectPanic(t, "MustExecutor()", tc, want)
		_ = MustExecutor(ctx)
	})
}

func newTestLogger() tuxlogi.Logger {
	config := tuxlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = new(bytes.Buffer)
	return tuxlog.NewLogger(config)
}
