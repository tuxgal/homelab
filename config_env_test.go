package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/tuxdude/zzzlog"
)

var configEnvTests = []struct {
	name    string
	ctxInfo *testContextInfo
	test    func(*testing.T, context.Context, string)
}{
	{
		name:    "Config Env - New",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			wantEnvMap := envMap{
				"$$HOST_IP$$":                  "10.76.77.78",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
			}

			env := newConfigEnv(ctx)
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env map", wantEnvMap, env.env) {
				return
			}
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, env.envKeyOrder) {
				return
			}
		},
	},
	{
		name:    "Config Env - Override - No overlap",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			override := envMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
				"BAZ1": "baz1",
			}
			overrideOrder := []string{
				"FOO1",
				"BAR1",
				"BAZ1",
			}
			wantEnvMap := envMap{
				"$$HOST_IP$$":                  "10.76.77.78",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost",
				"$$FOO1$$":                     "foo1",
				"$$BAR1$$":                     "bar1",
				"$$BAZ1$$":                     "baz1",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
				"$$FOO1$$",
				"$$BAR1$$",
				"$$BAZ1$$",
			}

			env := newConfigEnv(ctx)
			got := env.override(ctx, override, overrideOrder)
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name:    "Config Env - Override - With overlap",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			override := envMap{
				"FOO1":                     "foo1",
				"HOST_IP":                  "10.76.77.178",
				"BAR1":                     "bar1",
				"HUMAN_FRIENDLY_HOST_NAME": "FakeHost2",
			}
			overrideOrder := []string{
				"FOO1",
				"HOST_IP",
				"BAR1",
				"HUMAN_FRIENDLY_HOST_NAME",
			}
			wantEnvMap := envMap{
				"$$HOST_IP$$":                  "10.76.77.178",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost2",
				"$$FOO1$$":                     "foo1",
				"$$BAR1$$":                     "bar1",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
				"$$FOO1$$",
				"$$BAR1$$",
			}

			env := newConfigEnv(ctx)
			got := env.override(ctx, override, overrideOrder)
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - No Match",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "host_ip=$HOST_IP$$;host_name=$$HOST_NAME$;friendly=HUMAN_FRIENDLY_HOST_NAME"
			want := "host_ip=$HOST_IP$$;host_name=$$HOST_NAME$;friendly=HUMAN_FRIENDLY_HOST_NAME"

			env := newConfigEnv(ctx)
			got := env.apply(input)
			if !testCmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - Single Match",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "host_ip=$HOST_IP$$;host_name=$$HOST_NAME$$;friendly=HUMAN_FRIENDLY_HOST_NAME"
			want := "host_ip=$HOST_IP$$;host_name=fakehost;friendly=HUMAN_FRIENDLY_HOST_NAME"

			env := newConfigEnv(ctx)
			got := env.apply(input)
			if !testCmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - Multi Match",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "host_ip=$$HOST_IP$$;host_name=$$HOST_NAME$$;friendly=$$HUMAN_FRIENDLY_HOST_NAME$$"
			want := "host_ip=10.76.77.78;host_name=fakehost;friendly=FakeHost"

			env := newConfigEnv(ctx)
			got := env.apply(input)
			if !testCmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - Recursive Match",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "foo=$$FOO$$"
			want := "foo=$$BAR$$"

			env := newConfigEnv(ctx).override(
				ctx,
				envMap{
					"FOO": "$$BAR$$",
					"BAR": "$$BAZ$$",
					"BAZ": "baz1",
				},
				[]string{
					"FOO",
					"BAR",
					"BAZ",
				},
			)
			got := env.apply(input)
			if !testCmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - New - Missing Host Info",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, _ context.Context, tc string) {
			want := `Unable to find host info in context`
			logger := newCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := withLogger(context.Background(), logger)

			defer testExpectPanic(t, "newConfigEnv", tc, want)
			_ = newConfigEnv(ctx)
		},
	},
	{
		name:    "Config Env - Override - Unequal Lengths Between Override Map And Order",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			want := `Override map \(len:2\) and order slice \(len:1\) are of unequal lengths`
			override := envMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
			}
			overrideOrder := []string{
				"FOO1",
			}
			logger := newCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx = withLogger(ctx, logger)
			env := newConfigEnv(ctx)

			defer testExpectPanic(t, "configEnv.override()", tc, want)
			_ = env.override(ctx, override, overrideOrder)
		},
	},
	{
		name:    "Config Env - Override - Invalid Override Map Key In Order",
		ctxInfo: &testContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			want := `Expected key BAZ1 not found in override map input`
			override := envMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
			}
			overrideOrder := []string{
				"FOO1",
				"BAZ1",
			}
			logger := newCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx = withLogger(ctx, logger)
			env := newConfigEnv(ctx)

			defer testExpectPanic(t, "configEnv.override()", tc, want)
			_ = env.override(ctx, override, overrideOrder)
		},
	},
}

func TestConfigEnv(t *testing.T) {
	for _, tc := range configEnvTests {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, newTestContext(tc.ctxInfo), tc.name)
		})
	}
}