package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tuxdude/zzzlogi"

	"github.com/clarketm/json"
)

const (
	logLevelEnvVar   = "HOMELAB_LOG_LEVEL"
	logLevelEnvDebug = "debug"
	logLevelEnvTrace = "trace"
)

var (
	loggerKey          = ctxKeyLogger{}
	dockerAPIClientKey = ctxKeyDockerAPIClient{}
	hostInfoKey        = ctxKeyHostInfo{}
)

type stringSet map[string]struct{}

type ctxKeyLogger struct{}
type ctxKeyDockerAPIClient struct{}
type ctxKeyHostInfo struct{}

func log(ctx context.Context) zzzlogi.Logger {
	logger, ok := ctx.Value(loggerKey).(zzzlogi.Logger)
	if !ok {
		panic("Unable to retriever logger from context")
	}
	return logger
}

func dockerAPIClientFromContext(ctx context.Context) (dockerAPIClient, bool) {
	client, ok := ctx.Value(dockerAPIClientKey).(dockerAPIClient)
	return client, ok
}

func hostInfoFromContext(ctx context.Context) (*hostInfo, bool) {
	client, ok := ctx.Value(hostInfoKey).(*hostInfo)
	return client, ok
}

func withLogger(ctx context.Context, logger zzzlogi.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// This is used purely by tests.
// nolint:unused
func withDockerAPIClient(ctx context.Context, client dockerAPIClient) context.Context {
	return context.WithValue(ctx, dockerAPIClientKey, client)
}

// This is used purely by tests.
// nolint:unused
func withHostInfo(ctx context.Context, host *hostInfo) context.Context {
	return context.WithValue(ctx, hostInfoKey, host)
}

// Returns the JSON formatted string representation of the specified object.
func prettyPrintJSON(x interface{}) string {
	p, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(p)
}

func logToErrorAndReturn(ctx context.Context, format string, args ...interface{}) error {
	log(ctx).Errorf(format, args...)
	log(ctx).ErrorEmpty()
	return fmt.Errorf(format, args...)
}

// TODO: Remove this after this function is used.
// nolint (unused)
func logToWarnAndReturn(ctx context.Context, format string, args ...interface{}) error {
	log(ctx).Warnf(format, args...)
	log(ctx).WarnEmpty()
	return fmt.Errorf(format, args...)
}

func isLogLevelDebug() bool {
	return isEnvValue(logLevelEnvVar, logLevelEnvDebug)
}

func isLogLevelTrace() bool {
	return isEnvValue(logLevelEnvVar, logLevelEnvTrace)
}

func isEnvValue(envVar string, envValue string) bool {
	val, isVarSet := os.LookupEnv(envVar)
	return isVarSet && val == envValue
}

func newBool(b bool) *bool {
	return &b
}
