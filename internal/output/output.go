package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	cfg "github.com/bbleimschein/gsc-cli/internal/config"
	gsc "github.com/bbleimschein/gsc-cli/internal/searchconsole"
)

func RenderSites(w io.Writer, mode string, sites []gsc.Site) error {
	if mode == cfg.OutputJSON {
		return writeJSON(w, sites)
	}

	tw := newWriter(w)
	fmt.Fprintln(tw, "SITE_URL\tPERMISSION_LEVEL")
	for _, site := range sites {
		fmt.Fprintf(tw, "%s\t%s\n", site.SiteURL, site.PermissionLevel)
	}
	return tw.Flush()
}

func RenderAnalytics(w io.Writer, mode string, result *gsc.AnalyticsResult) error {
	if mode == cfg.OutputJSON {
		return writeJSON(w, result)
	}

	tw := newWriter(w)
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

func RenderSitemaps(w io.Writer, mode string, sitemaps []gsc.Sitemap) error {
	if mode == cfg.OutputJSON {
		return writeJSON(w, sitemaps)
	}

	tw := newWriter(w)
	fmt.Fprintln(tw, "PATH\tPENDING\tLAST_SUBMITTED\tLAST_DOWNLOADED\tWARNINGS\tERRORS")
	for _, sitemap := range sitemaps {
		fmt.Fprintf(
			tw,
			"%s\t%t\t%s\t%s\t%d\t%d\n",
			sitemap.Path,
			sitemap.IsPending,
			sitemap.LastSubmitted,
			sitemap.LastDownloaded,
			sitemap.Warnings,
			sitemap.Errors,
		)
	}
	return tw.Flush()
}

func RenderInspection(w io.Writer, mode string, result *gsc.InspectionResult) error {
	if mode == cfg.OutputJSON {
		return writeJSON(w, result)
	}

	tw := newWriter(w)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	rows := [][2]string{
		{"SITE", result.Site},
		{"INSPECTION_URL", result.InspectionURL},
		{"VERDICT", result.Verdict},
		{"COVERAGE_STATE", result.CoverageState},
		{"INDEXING_STATE", result.IndexingState},
		{"PAGE_FETCH_STATE", result.PageFetchState},
		{"ROBOTS_TXT_STATE", result.RobotsTxtState},
		{"CRAWLED_AS", result.CrawledAs},
		{"LAST_CRAWL_TIME", result.LastCrawlTime},
		{"GOOGLE_CANONICAL", result.GoogleCanonical},
		{"USER_CANONICAL", result.UserCanonical},
		{"REFERRING_URLS", strings.Join(result.ReferringURLs, ", ")},
		{"SITEMAPS", strings.Join(result.Sitemaps, ", ")},
		{"INSPECTION_RESULT_LINK", result.InspectionResultLink},
	}
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
	}
	return tw.Flush()
}

func RenderOverview(w io.Writer, mode string, report *gsc.OverviewReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(w, report)
	}

	tw := newWriter(w)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "SITE\t%s\n", report.Site)
	fmt.Fprintf(tw, "START_DATE\t%s\n", report.StartDate)
	fmt.Fprintf(tw, "END_DATE\t%s\n", report.EndDate)
	fmt.Fprintf(tw, "INSPECT_URL\t%s\n", report.InspectURL)
	_ = tw.Flush()

	fmt.Fprintln(w)
	fmt.Fprintln(w, "TOP_PAGES")
	if err := RenderAnalytics(w, cfg.OutputTable, report.TopPages); err != nil {
		return err
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "TOP_QUERIES")
	if err := RenderAnalytics(w, cfg.OutputTable, report.TopQueries); err != nil {
		return err
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "SITEMAPS")
	if err := RenderSitemaps(w, cfg.OutputTable, report.Sitemaps); err != nil {
		return err
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "INSPECTION")
	return RenderInspection(w, cfg.OutputTable, report.Inspection)
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func newWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
