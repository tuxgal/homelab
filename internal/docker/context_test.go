package docker

import (
	"bytes"
	"context"
	"testing"

	l "github.com/tuxgal/homelab/internal/log"
	"github.com/tuxgal/homelab/internal/testhelpers"
	"github.com/tuxgal/tuxlog"
	"github.com/tuxgal/tuxlogi"
)

func TestRetrieveAPIClientFromEmptyContext(t *testing.T) {
	t.Parallel()

	tc := "Retrieve Docker API Client - Empty Context"
	want := `Docker API Client not found in context`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctx = l.WithLogger(ctx, newTestLogger())

		defer testhelpers.ExpectPanic(t, "MustAPIClient()", tc, want)
		_ = MustAPIClient(ctx)
	})
}

func newTestLogger() tuxlogi.Logger {
	config := tuxlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = new(bytes.Buffer)
	return tuxlog.NewLogger(config)
}
