package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersionInfoFallsBackToBuildInfo(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "5bac3f5abc"},
			{Key: "vcs.time", Value: "2026-06-29T08:02:00Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	got := resolveVersionInfo("dev", "", "", info)
	if got != (versionInfo{Version: "dev", Commit: "5bac3f5abc-dirty", Date: "2026-06-29T08:02:00Z"}) {
		t.Fatalf("resolveVersionInfo() = %#v", got)
	}

	info.Main.Version = "v0.3.0"
	if got := resolveVersionInfo("dev", "", "", info); got.Version != "v0.3.0" {
		t.Fatalf("module version not used: %#v", got)
	}

	stamped := resolveVersionInfo("v1.0.0", "abc", "2026-01-01", info)
	if stamped != (versionInfo{Version: "v1.0.0", Commit: "abc", Date: "2026-01-01"}) {
		t.Fatalf("ldflags values overridden: %#v", stamped)
	}

	if got := resolveVersionInfo("dev", "", "", nil); got != (versionInfo{Version: "dev"}) {
		t.Fatalf("nil build info = %#v", got)
	}
}
