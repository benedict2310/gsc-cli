package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	cfg "github.com/bbleimschein/gsc-cli/internal/config"
	gsc "github.com/bbleimschein/gsc-cli/internal/searchconsole"
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
	Date    string `json:"date,omitempty"`
}

type doctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type doctorReport struct {
	Version string        `json:"version"`
	Site    string        `json:"site,omitempty"`
	Checks  []doctorCheck `json:"checks"`
}

func renderVersion(mode string, info versionInfo) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, info)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "VERSION\t%s\n", info.Version)
	fmt.Fprintf(tw, "COMMIT\t%s\n", info.Commit)
	fmt.Fprintf(tw, "DATE\t%s\n", info.Date)
	return tw.Flush()
}

func renderDoctor(mode string, report doctorReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CHECK\tSTATUS\tDETAIL")
	for _, check := range report.Checks {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", check.Name, check.Status, check.Detail)
	}
	return tw.Flush()
}

func writeJSON(w *os.File, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func statusForMatch(found bool) string {
	if found {
		return "PASS"
	}
	return "FAIL"
}

func hasSite(sites []gsc.Site, target string) bool {
	for _, site := range sites {
		if site.SiteURL == target {
			return true
		}
	}
	return false
}
