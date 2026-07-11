package inspect

import "context"

type HomelabInspectLevel uint8

const (
	HomelabInspectLevelNone HomelabInspectLevel = iota
	HomelabInspectLevelDebug
	HomelabInspectLevelTrace
)

var (
	homelabInspectLevelKey = ctxKeyHomelabInspectLevel{}
)

type ctxKeyHomelabInspectLevel struct{}

func HomelabInspectLevelFromContext(ctx context.Context) HomelabInspectLevel {
	lvl, ok := ctx.Value(homelabInspectLevelKey).(HomelabInspectLevel)
	if !ok {
		return HomelabInspectLevelNone
	}
	return lvl
}

func WithHomelabInspectLevel(ctx context.Context, lvl HomelabInspectLevel) context.Context {
	return context.WithValue(ctx, homelabInspectLevelKey, lvl)
}
