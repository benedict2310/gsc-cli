package main

import (
	"encoding/json"
	"strings"
	"testing"

	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

func TestIndexabilityRowKeepsCoverageAndLastCrawl(t *testing.T) {
	row := newIndexabilityRow("https://example.com/", gsc.InspectionResult{
		Verdict:       "PASS",
		CoverageState: "Submitted and indexed",
		LastCrawlTime: "2026-11-09T01:03:26Z",
	}, nil)

	raw, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"coverageState":"Submitted and indexed"`, `"lastCrawlTime":"2026-11-09T01:03:26Z"`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("row JSON = %s, want %s", raw, want)
		}
	}
}
