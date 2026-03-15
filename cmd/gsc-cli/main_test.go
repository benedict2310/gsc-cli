package main

import (
	"context"
	"strings"
	"testing"
)

func TestRunAnalyticsQueryParsesFirstFlag(t *testing.T) {
	t.Parallel()

	err := runAnalyticsQuery(context.Background(), []string{
		"--site", "sc-domain:futurelab.studio",
		"--credentials-file", "/does/not/exist.json",
		"--start", "2026-03-01",
		"--end", "2026-03-14",
		"--dimensions", "query",
	})
	if err == nil {
		t.Fatal("runAnalyticsQuery() expected error")
	}
	if strings.Contains(err.Error(), "invalid start date") {
		t.Fatalf("runAnalyticsQuery() still skipped the first flag: %v", err)
	}
}
