package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/benedict2310/gsc-cli/internal/commandspec"
	cfg "github.com/benedict2310/gsc-cli/internal/config"
	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
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

func renderCommandManifest(mode string, manifest commandspec.Manifest) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, manifest)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tSUMMARY\tREQUIRES_CREDENTIALS\tREQUIRES_SITE\tAPIS_USED\tOUTPUT_MODES\tSAFETY\tWRITE_SCOPE\tFLAGS")
	for _, command := range manifest.Commands {
		fmt.Fprintf(tw, "%s\t%s\t%t\t%t\t%s\t%s\t%s\t%t\t%s\n", command.Path, command.Summary, command.RequiresCredentials, command.RequiresSite, strings.Join(command.APIsUsed, ","), strings.Join(command.OutputModes, ","), command.Safety, command.WriteScopeRequired, flagNames(command.Flags))
	}
	return tw.Flush()
}

func renderCommandSpec(mode string, command commandspec.Command) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, command)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "PATH\t%s\n", command.Path)
	fmt.Fprintf(tw, "SUMMARY\t%s\n", command.Summary)
	fmt.Fprintf(tw, "DESCRIPTION\t%s\n", command.Description)
	fmt.Fprintf(tw, "REQUIRES_CREDENTIALS\t%t\n", command.RequiresCredentials)
	fmt.Fprintf(tw, "REQUIRES_SITE\t%t\n", command.RequiresSite)
	fmt.Fprintf(tw, "APIS_USED\t%s\n", strings.Join(command.APIsUsed, ","))
	fmt.Fprintf(tw, "SAFETY\t%s\n", command.Safety)
	fmt.Fprintf(tw, "WRITE_SCOPE_REQUIRED\t%t\n", command.WriteScopeRequired)
	fmt.Fprintf(tw, "OUTPUT_MODES\t%s\n", strings.Join(command.OutputModes, ","))
	fmt.Fprintf(tw, "FLAGS\t%s\n", flagSummaries(command.Flags))
	fmt.Fprintf(tw, "EXAMPLES\t%s\n", strings.Join(command.Examples, " | "))
	fmt.Fprintf(tw, "CAVEATS\t%s\n", strings.Join(command.Caveats, " | "))
	fmt.Fprintf(tw, "RELATED_COMMANDS\t%s\n", strings.Join(command.RelatedCommands, ","))
	return tw.Flush()
}

func flagNames(flags []commandspec.Flag) string {
	names := make([]string, 0, len(flags))
	for _, flag := range flags {
		names = append(names, flag.Name)
	}
	return strings.Join(names, ",")
}

func flagSummaries(flags []commandspec.Flag) string {
	parts := make([]string, 0, len(flags))
	for _, flag := range flags {
		part := flag.Name + ":" + flag.Type
		if flag.Required {
			part += ":required"
		}
		if flag.Default != "" {
			part += ":default=" + flag.Default
		}
		if flag.EnvFallback != "" {
			part += ":env=" + flag.EnvFallback
		}
		if len(flag.AllowedValues) > 0 {
			part += ":allowed=" + strings.Join(flag.AllowedValues, "/")
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
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

func siteAccessDetail(site, account string, found bool) string {
	if found {
		return site
	}
	if account == "" {
		account = "the service account"
	}
	return fmt.Sprintf("%s is not visible to %s; add that account as a user on the property in Search Console under Settings > Users and permissions", site, account)
}
