package searchconsole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gscapi "google.golang.org/api/searchconsole/v1"
)

func TestQueryAnalyticsShapesRequest(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
		gotBody   gscapi.SearchAnalyticsQueryRequest
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responseAggregationType":"byPage","rows":[{"keys":["q","/"],"clicks":2,"impressions":10,"ctr":0.2,"position":3.5}]}`))
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{
		HTTPClient: server.Client(),
		Endpoint:   server.URL + "/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := &gscapi.SearchAnalyticsQueryRequest{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-03",
		Dimensions: []string{"QUERY", "PAGE"},
		SearchType: "WEB",
		RowLimit:   100,
	}
	result, err := client.QueryAnalytics(context.Background(), "sc-domain:futurelab.studio", req)
	if err != nil {
		t.Fatalf("QueryAnalytics() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s, want POST", gotMethod)
	}
	if !strings.Contains(gotPath, "/sites/sc-domain:futurelab.studio/searchAnalytics/query") {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody.StartDate != "2026-03-01" || gotBody.RowLimit != 100 {
		t.Fatalf("request body = %#v", gotBody)
	}
	if len(result.Rows) != 1 || result.Rows[0].Clicks != 2 {
		t.Fatalf("result = %#v", result)
	}
}

func TestInspectURLShapesRequest(t *testing.T) {
	t.Parallel()

	var gotBody gscapi.InspectUrlIndexRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/urlInspection/index:inspect") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"inspectionResult":{"inspectionResultLink":"https://search.google.com/test","indexStatusResult":{"verdict":"PASS","coverageState":"Submitted and indexed","robotsTxtState":"ALLOWED","pageFetchState":"SUCCESSFUL"}}}`))
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{
		HTTPClient: server.Client(),
		Endpoint:   server.URL + "/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := client.InspectURL(context.Background(), &gscapi.InspectUrlIndexRequest{
		SiteUrl:       "sc-domain:futurelab.studio",
		InspectionUrl: "https://futurelab.studio/",
		LanguageCode:  "en-US",
	})
	if err != nil {
		t.Fatalf("InspectURL() error = %v", err)
	}

	if gotBody.SiteUrl != "sc-domain:futurelab.studio" || gotBody.InspectionUrl != "https://futurelab.studio/" {
		t.Fatalf("request body = %#v", gotBody)
	}
	if result.Verdict != "PASS" || result.PageFetchState != "SUCCESSFUL" {
		t.Fatalf("result = %#v", result)
	}
}

func TestListSitesSanitizesAPIErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"code":403,"message":"forbidden\nraw detail"}}`, http.StatusForbidden)
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{
		HTTPClient: server.Client(),
		Endpoint:   server.URL + "/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.ListSites(context.Background())
	if err == nil {
		t.Fatal("ListSites() expected error")
	}
	if strings.Contains(err.Error(), "\n") {
		t.Fatalf("error contains newline: %q", err)
	}
}
