package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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

func renderCompare(mode string, report gsc.CompareReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "SITE\t%s\n", report.Site)
	fmt.Fprintf(tw, "DIMENSION\t%s\n", report.Dimension)
	fmt.Fprintf(tw, "CURRENT_RANGE\t%s to %s\n", report.Summary.CurrentStart, report.Summary.CurrentEnd)
	fmt.Fprintf(tw, "PREVIOUS_RANGE\t%s to %s\n", report.Summary.PreviousStart, report.Summary.PreviousEnd)
	fmt.Fprintf(tw, "CURRENT_ROWS\t%d\n", report.Summary.CurrentRows)
	fmt.Fprintf(tw, "PREVIOUS_ROWS\t%d\n", report.Summary.PreviousRows)
	fmt.Fprintf(tw, "CURRENT_IMPRESSIONS\t%.0f\n", report.Summary.CurrentImpressions)
	fmt.Fprintf(tw, "PREVIOUS_IMPRESSIONS\t%.0f\n", report.Summary.PreviousImpressions)
	fmt.Fprintf(tw, "IMPRESSION_DELTA\t%.0f\n", report.Summary.ImpressionDelta)
	fmt.Fprintf(tw, "CURRENT_CLICKS\t%.0f\n", report.Summary.CurrentClicks)
	fmt.Fprintf(tw, "PREVIOUS_CLICKS\t%.0f\n", report.Summary.PreviousClicks)
	fmt.Fprintf(tw, "CLICK_DELTA\t%.0f\n", report.Summary.ClickDelta)
	_ = tw.Flush()

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "CURRENT")
	if err := renderAnalyticsTable(report.Current); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "PREVIOUS")
	return renderAnalyticsTable(report.Previous)
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

func sumAnalyticsRows(rows []gsc.AnalyticsRow) (float64, float64) {
	var impressions float64
	var clicks float64
	for _, row := range rows {
		impressions += row.Impressions
		clicks += row.Clicks
	}
	return impressions, clicks
}

func renderAnalyticsTable(result *gsc.AnalyticsResult) error {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	headers := append(append([]string{}, result.Dimensions...), "CLICKS", "IMPRESSIONS", "CTR", "POSITION")
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range result.Rows {
		values := append([]string{}, row.Keys...)
		values = append(values,
			formatFloat(row.Clicks),
			formatFloat(row.Impressions),
			formatFloat(row.CTR),
			formatFloat(row.Position),
		)
		fmt.Fprintln(tw, strings.Join(values, "\t"))
	}
	return tw.Flush()
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
