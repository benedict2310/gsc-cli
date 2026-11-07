package main

import "testing"

func TestDaysBetween(t *testing.T) {
	if got := daysBetween("2026-10-04", "2026-11-01"); got != 28 {
		t.Fatalf("daysBetween() = %d, want 28", got)
	}
}
