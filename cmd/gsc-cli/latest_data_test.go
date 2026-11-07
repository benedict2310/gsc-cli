package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cfg "github.com/benedict2310/gsc-cli/internal/config"
	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

func TestAlignToLatestDataEndsWindowAtLastDateWithRows(t *testing.T) {
	var gotBody struct {
		StartDate  string   `json:"startDate"`
		EndDate    string   `json:"endDate"`
		Dimensions []string `json:"dimensions"`
		DataState  string   `json:"dataState"`
	}
	rows := `{"rows":[{"keys":["2026-11-03"],"impressions":5},{"keys":["2026-11-04"],"impressions":7}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rows))
	}))
	defer server.Close()
	client, err := gsc.New(context.Background(), gsc.Options{HTTPClient: server.Client(), Endpoint: server.URL + "/"})
	if err != nil {
		t.Fatalf("gsc.New() error = %v", err)
	}

	window := cfg.DateRange{Start: "2026-10-10", End: "2026-11-06"}
	got, err := alignToLatestData(context.Background(), client, "sc-domain:example.com", window, "web", "final")
	if err != nil {
		t.Fatalf("alignToLatestData() error = %v", err)
	}
	if got != (cfg.DateRange{Start: "2026-10-08", End: "2026-11-04"}) {
		t.Fatalf("alignToLatestData() = %#v", got)
	}
	if gotBody.EndDate != "2026-11-06" || gotBody.StartDate != "2026-10-24" || len(gotBody.Dimensions) != 1 || gotBody.Dimensions[0] != "DATE" || gotBody.DataState != "FINAL" {
		t.Fatalf("probe request = %#v", gotBody)
	}

	rows = `{}`
	got, err = alignToLatestData(context.Background(), client, "sc-domain:example.com", window, "web", "final")
	if err != nil || got != window {
		t.Fatalf("alignToLatestData() without rows = %#v, %v; want unchanged", got, err)
	}

	got, err = alignToLatestData(context.Background(), client, "sc-domain:example.com", window, "web", "hourly_all")
	if err != nil || got != window {
		t.Fatalf("alignToLatestData() for hourly_all = %#v, %v; want unchanged", got, err)
	}
}
