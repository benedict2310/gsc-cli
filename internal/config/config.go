package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	gscapi "google.golang.org/api/searchconsole/v1"
)

const (
	OutputJSON  = "json"
	OutputTable = "table"

	envCredentialsFile = "GSC_CLI_CREDENTIALS_FILE"
	envSiteURL         = "GSC_CLI_SITE_URL"
)

var (
	allowedOutputs = map[string]struct{}{
		OutputJSON:  {},
		OutputTable: {},
	}
	allowedDimensions = map[string]struct{}{
		"DATE":              {},
		"QUERY":             {},
		"PAGE":              {},
		"COUNTRY":           {},
		"DEVICE":            {},
		"SEARCH_APPEARANCE": {},
		"HOUR":              {},
	}
	allowedSearchTypes = map[string]struct{}{
		"WEB":         {},
		"IMAGE":       {},
		"VIDEO":       {},
		"NEWS":        {},
		"DISCOVER":    {},
		"GOOGLE_NEWS": {},
	}
	allowedDataStates = map[string]struct{}{
		"FINAL":      {},
		"ALL":        {},
		"HOURLY_ALL": {},
	}
)

type Common struct {
	CredentialsFile string
	Site            string
	Output          string
}

type AnalyticsInput struct {
	StartDate  string
	EndDate    string
	Dimensions string
	SearchType string
	DataState  string
	Limit      int64
	StartRow   int64
}

type InspectInput struct {
	Site         string
	URL          string
	LanguageCode string
}

type DateRange struct {
	Start string
	End   string
}

var allowedDayWindows = map[int]int{
	7:  7,
	28: 28,
	90: 90,
}

func (c *Common) Resolve(requireSite bool) error {
	c.CredentialsFile = firstNonEmpty(c.CredentialsFile, os.Getenv(envCredentialsFile))
	c.Site = firstNonEmpty(c.Site, os.Getenv(envSiteURL))
	c.Output = strings.ToLower(firstNonEmpty(c.Output, OutputJSON))

	if err := validateCredentialsFile(c.CredentialsFile); err != nil {
		return err
	}
	if requireSite && strings.TrimSpace(c.Site) == "" {
		return errors.New("site is required; pass --site or set GSC_CLI_SITE_URL")
	}
	if _, ok := allowedOutputs[c.Output]; !ok {
		return fmt.Errorf("unsupported output mode %q; supported values are json and table", c.Output)
	}
	return nil
}

func BuildAnalyticsRequest(input AnalyticsInput) (*gscapi.SearchAnalyticsQueryRequest, error) {
	startDate := strings.TrimSpace(input.StartDate)
	endDate := strings.TrimSpace(input.EndDate)
	if err := validateDate(startDate); err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	if err := validateDate(endDate); err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}
	if startDate > endDate {
		return nil, errors.New("start date must be less than or equal to end date")
	}
	if input.Limit < 1 || input.Limit > 25000 {
		return nil, errors.New("limit must be between 1 and 25000")
	}
	if input.StartRow < 0 {
		return nil, errors.New("start-row must be greater than or equal to 0")
	}

	dimensions, err := parseDimensions(input.Dimensions)
	if err != nil {
		return nil, err
	}
	searchType, err := normalizeSearchType(input.SearchType)
	if err != nil {
		return nil, err
	}
	dataState, err := normalizeDataState(input.DataState)
	if err != nil {
		return nil, err
	}
	if contains(dimensions, "HOUR") && len(dimensions) > 1 {
		return nil, errors.New(`dimension "HOUR" cannot be combined with other dimensions`)
	}
	if contains(dimensions, "HOUR") && dataState != "HOURLY_ALL" {
		return nil, errors.New(`dimension "HOUR" requires data-state "hourly_all"`)
	}

	return &gscapi.SearchAnalyticsQueryRequest{
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: dimensions,
		SearchType: searchType,
		DataState:  dataState,
		RowLimit:   input.Limit,
		StartRow:   input.StartRow,
	}, nil
}

func BuildInspectRequest(input InspectInput) (*gscapi.InspectUrlIndexRequest, error) {
	site := strings.TrimSpace(input.Site)
	targetURL := strings.TrimSpace(input.URL)
	languageCode := firstNonEmpty(strings.TrimSpace(input.LanguageCode), "en-US")

	if site == "" {
		return nil, errors.New("site is required")
	}
	if targetURL == "" {
		return nil, errors.New("url is required")
	}
	if !looksLikeURL(targetURL) {
		return nil, errors.New("url must start with http:// or https://")
	}
	if languageCode == "" {
		return nil, errors.New("language is required")
	}

	return &gscapi.InspectUrlIndexRequest{
		SiteUrl:       site,
		InspectionUrl: targetURL,
		LanguageCode:  languageCode,
	}, nil
}

func ResolveDateRange(startDate, endDate string) (DateRange, error) {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		end := time.Now().AddDate(0, 0, -1)
		start := end.AddDate(0, 0, -27)
		return DateRange{
			Start: start.Format("2006-01-02"),
			End:   end.Format("2006-01-02"),
		}, nil
	}
	if startDate == "" || endDate == "" {
		return DateRange{}, errors.New("start and end dates must be provided together")
	}
	if err := validateDate(startDate); err != nil {
		return DateRange{}, fmt.Errorf("invalid start date: %w", err)
	}
	if err := validateDate(endDate); err != nil {
		return DateRange{}, fmt.Errorf("invalid end date: %w", err)
	}
	if startDate > endDate {
		return DateRange{}, errors.New("start date must be less than or equal to end date")
	}
	return DateRange{Start: startDate, End: endDate}, nil
}

func ResolveDateRangeWithDays(startDate, endDate string, days int) (DateRange, error) {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)

	if days != 0 && (startDate != "" || endDate != "") {
		return DateRange{}, errors.New("days cannot be combined with start or end dates")
	}
	if days == 0 {
		return ResolveDateRange(startDate, endDate)
	}
	if _, ok := allowedDayWindows[days]; !ok {
		return DateRange{}, errors.New("days must be one of 7, 28, or 90")
	}

	end := time.Now().AddDate(0, 0, -1)
	start := end.AddDate(0, 0, -(days - 1))
	return DateRange{
		Start: start.Format("2006-01-02"),
		End:   end.Format("2006-01-02"),
	}, nil
}

func ParseDayWindow(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	switch value {
	case "7":
		return 7, nil
	case "28":
		return 28, nil
	case "90":
		return 90, nil
	default:
		return 0, errors.New("days must be one of 7, 28, or 90")
	}
}

func ParsePeriodRange(value string) (DateRange, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return DateRange{}, errors.New("period must be in YYYY-MM-DD:YYYY-MM-DD format")
	}
	return ResolveDateRange(parts[0], parts[1])
}

func ResolveBaseline(current DateRange, baseline string) (DateRange, error) {
	switch strings.ToLower(strings.TrimSpace(baseline)) {
	case "", "previous-period":
		return PreviousDateRange(current)
	case "same-period-last-year":
		start, err := time.Parse("2006-01-02", current.Start)
		if err != nil {
			return DateRange{}, fmt.Errorf("parse current start date: %w", err)
		}
		end, err := time.Parse("2006-01-02", current.End)
		if err != nil {
			return DateRange{}, fmt.Errorf("parse current end date: %w", err)
		}
		return DateRange{Start: start.AddDate(-1, 0, 0).Format("2006-01-02"), End: end.AddDate(-1, 0, 0).Format("2006-01-02")}, nil
	default:
		return ParsePeriodRange(baseline)
	}
}

var namedPeriodPattern = regexp.MustCompile(`^last-(\d+)([dm])$`)

func ResolveNamedPeriod(value string, now time.Time) (DateRange, error) {
	matches := namedPeriodPattern.FindStringSubmatch(strings.ToLower(strings.TrimSpace(value)))
	if len(matches) != 3 {
		return DateRange{}, errors.New("period must look like last-7d or last-16m")
	}
	count, err := strconv.Atoi(matches[1])
	if err != nil || count < 1 {
		return DateRange{}, errors.New("period count must be positive")
	}
	end := now.AddDate(0, 0, -1)
	switch matches[2] {
	case "d":
		start := end.AddDate(0, 0, -(count - 1))
		return DateRange{Start: start.Format("2006-01-02"), End: end.Format("2006-01-02")}, nil
	case "m":
		start := subtractMonthsClamped(end, count).AddDate(0, 0, 1)
		return DateRange{Start: start.Format("2006-01-02"), End: end.Format("2006-01-02")}, nil
	default:
		return DateRange{}, errors.New("unsupported period unit")
	}
}

// subtractMonthsClamped moves back whole calendar months, clamping the day to the target
// month's length so 31 March minus one month is 28 February, not 3 March.
func subtractMonthsClamped(t time.Time, months int) time.Time {
	firstOfTarget := time.Date(t.Year(), t.Month()-time.Month(months), 1, 0, 0, 0, 0, t.Location())
	lastDay := firstOfTarget.AddDate(0, 1, -1).Day()
	day := t.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(firstOfTarget.Year(), firstOfTarget.Month(), day, 0, 0, 0, 0, t.Location())
}

func PreviousDateRange(current DateRange) (DateRange, error) {
	start, err := time.Parse("2006-01-02", current.Start)
	if err != nil {
		return DateRange{}, fmt.Errorf("parse current start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", current.End)
	if err != nil {
		return DateRange{}, fmt.Errorf("parse current end date: %w", err)
	}
	days := int(end.Sub(start).Hours()/24) + 1
	prevEnd := start.AddDate(0, 0, -1)
	prevStart := prevEnd.AddDate(0, 0, -(days - 1))
	return DateRange{
		Start: prevStart.Format("2006-01-02"),
		End:   prevEnd.Format("2006-01-02"),
	}, nil
}

func EndRangeAt(r DateRange, end string) (DateRange, error) {
	end = strings.TrimSpace(end)
	if end == "" || end >= r.End {
		return r, nil
	}
	latest, err := time.Parse("2006-01-02", end)
	if err != nil {
		return DateRange{}, fmt.Errorf("parse end date: %w", err)
	}
	currentEnd, err := time.Parse("2006-01-02", r.End)
	if err != nil {
		return DateRange{}, fmt.Errorf("parse current end date: %w", err)
	}
	start, err := time.Parse("2006-01-02", r.Start)
	if err != nil {
		return DateRange{}, fmt.Errorf("parse current start date: %w", err)
	}
	shift := currentEnd.Sub(latest)
	return DateRange{
		Start: start.Add(-shift).Format("2006-01-02"),
		End:   end,
	}, nil
}

func InferInspectionURL(site string) (string, error) {
	site = strings.TrimSpace(site)
	if site == "" {
		return "", errors.New("site is required")
	}
	if strings.HasPrefix(site, "sc-domain:") {
		domain := strings.TrimPrefix(site, "sc-domain:")
		if domain == "" {
			return "", errors.New("invalid domain property")
		}
		return "https://" + domain + "/", nil
	}
	if looksLikeURL(site) {
		if strings.HasSuffix(site, "/") {
			return site, nil
		}
		return site + "/", nil
	}
	return "", errors.New("unable to infer inspection URL from site")
}

func parseDimensions(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("at least one dimension is required")
	}

	parts := strings.Split(value, ",")
	normalized := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		dimension := normalizeDimension(part)
		if dimension == "" {
			continue
		}
		if _, ok := allowedDimensions[dimension]; !ok {
			return nil, fmt.Errorf("unsupported dimension %q", part)
		}
		if _, ok := seen[dimension]; ok {
			continue
		}
		seen[dimension] = struct{}{}
		normalized = append(normalized, dimension)
	}
	if len(normalized) == 0 {
		return nil, errors.New("at least one dimension is required")
	}
	return normalized, nil
}

func normalizeDimension(value string) string {
	normalized := strings.TrimSpace(value)
	switch strings.ToLower(normalized) {
	case "searchappearance", "search_appearance", "search-appearance":
		return "SEARCH_APPEARANCE"
	default:
		return strings.ToUpper(normalized)
	}
}

func normalizeSearchType(value string) (string, error) {
	searchType := strings.ToUpper(strings.TrimSpace(firstNonEmpty(value, "WEB")))
	if _, ok := allowedSearchTypes[searchType]; !ok {
		return "", fmt.Errorf("unsupported search type %q", value)
	}
	return searchType, nil
}

func normalizeDataState(value string) (string, error) {
	dataState := strings.ToUpper(strings.TrimSpace(firstNonEmpty(value, "FINAL")))
	if _, ok := allowedDataStates[dataState]; !ok {
		return "", fmt.Errorf("unsupported data state %q", value)
	}
	return dataState, nil
}

func validateDate(value string) error {
	if value == "" {
		return errors.New("value is required")
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return errors.New("must be in YYYY-MM-DD format")
	}
	return nil
}

func validateCredentialsFile(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("credentials file is required; pass --credentials-file or set GSC_CLI_CREDENTIALS_FILE")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("credentials file %q does not exist", filepath.Clean(path))
		}
		return fmt.Errorf("stat credentials file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("credentials file %q is a directory", filepath.Clean(path))
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func looksLikeURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
