package main

import (
	"strings"
	"testing"
)

func TestSiteAccessDetailExplainsMissingAccess(t *testing.T) {
	if got := siteAccessDetail("sc-domain:example.com", "bot@project.iam.gserviceaccount.com", true); got != "sc-domain:example.com" {
		t.Fatalf("found detail = %q", got)
	}
	got := siteAccessDetail("sc-domain:example.com", "bot@project.iam.gserviceaccount.com", false)
	for _, want := range []string{"sc-domain:example.com", "bot@project.iam.gserviceaccount.com", "Users and permissions"} {
		if !strings.Contains(got, want) {
			t.Fatalf("detail %q does not contain %q", got, want)
		}
	}
}
