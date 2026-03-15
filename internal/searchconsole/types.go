package searchconsole

type Site struct {
	SiteURL         string `json:"siteUrl"`
	PermissionLevel string `json:"permissionLevel"`
}

type AnalyticsResult struct {
	Site                    string         `json:"site"`
	StartDate               string         `json:"startDate"`
	EndDate                 string         `json:"endDate"`
	Dimensions              []string       `json:"dimensions"`
	ResponseAggregationType string         `json:"responseAggregationType,omitempty"`
	Rows                    []AnalyticsRow `json:"rows"`
}

type AnalyticsRow struct {
	Keys        []string `json:"keys"`
	Clicks      float64  `json:"clicks"`
	Impressions float64  `json:"impressions"`
	CTR         float64  `json:"ctr"`
	Position    float64  `json:"position"`
}

type Sitemap struct {
	Path           string `json:"path"`
	IsPending      bool   `json:"isPending"`
	LastSubmitted  string `json:"lastSubmitted,omitempty"`
	LastDownloaded string `json:"lastDownloaded,omitempty"`
	Warnings       int64  `json:"warnings"`
	Errors         int64  `json:"errors"`
}

type InspectionResult struct {
	Site                 string   `json:"site"`
	InspectionURL        string   `json:"inspectionUrl"`
	Verdict              string   `json:"verdict,omitempty"`
	CoverageState        string   `json:"coverageState,omitempty"`
	IndexingState        string   `json:"indexingState,omitempty"`
	PageFetchState       string   `json:"pageFetchState,omitempty"`
	RobotsTxtState       string   `json:"robotsTxtState,omitempty"`
	CrawledAs            string   `json:"crawledAs,omitempty"`
	LastCrawlTime        string   `json:"lastCrawlTime,omitempty"`
	GoogleCanonical      string   `json:"googleCanonical,omitempty"`
	UserCanonical        string   `json:"userCanonical,omitempty"`
	ReferringURLs        []string `json:"referringUrls,omitempty"`
	Sitemaps             []string `json:"sitemaps,omitempty"`
	InspectionResultLink string   `json:"inspectionResultLink,omitempty"`
}

type OverviewReport struct {
	Site       string            `json:"site"`
	StartDate  string            `json:"startDate"`
	EndDate    string            `json:"endDate"`
	InspectURL string            `json:"inspectUrl"`
	TopPages   *AnalyticsResult  `json:"topPages"`
	TopQueries *AnalyticsResult  `json:"topQueries"`
	Sitemaps   []Sitemap         `json:"sitemaps"`
	Inspection *InspectionResult `json:"inspection"`
}

type CompareSummary struct {
	CurrentStart        string  `json:"currentStart"`
	CurrentEnd          string  `json:"currentEnd"`
	PreviousStart       string  `json:"previousStart"`
	PreviousEnd         string  `json:"previousEnd"`
	CurrentRows         int     `json:"currentRows"`
	PreviousRows        int     `json:"previousRows"`
	CurrentImpressions  float64 `json:"currentImpressions"`
	PreviousImpressions float64 `json:"previousImpressions"`
	CurrentClicks       float64 `json:"currentClicks"`
	PreviousClicks      float64 `json:"previousClicks"`
	ImpressionDelta     float64 `json:"impressionDelta"`
	ClickDelta          float64 `json:"clickDelta"`
}

type CompareReport struct {
	Site      string           `json:"site"`
	Dimension string           `json:"dimension"`
	Summary   CompareSummary   `json:"summary"`
	Current   *AnalyticsResult `json:"current"`
	Previous  *AnalyticsResult `json:"previous"`
}
