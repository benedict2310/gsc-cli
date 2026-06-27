package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
	gscapi "google.golang.org/api/searchconsole/v1"
)

type Cache struct {
	path string
}

type Status struct {
	Path          string `json:"path"`
	AnalyticsRows int    `json:"analyticsRows"`
	PullLogs      int    `json:"pullLogs"`
	Inspections   int    `json:"inspections"`
}

type PullLog struct {
	Hash         string `json:"hash"`
	Site         string `json:"site"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	Dimensions   string `json:"dimensions,omitempty"`
	SearchType   string `json:"searchType,omitempty"`
	DataState    string `json:"dataState,omitempty"`
	RowCount     int    `json:"rowCount"`
	PagesFetched int    `json:"pagesFetched"`
	CreatedAt    string `json:"createdAt"`
}

type AnalyticsRecord struct {
	Site             string  `json:"site"`
	Date             string  `json:"date,omitempty"`
	Hour             string  `json:"hour,omitempty"`
	Page             string  `json:"page,omitempty"`
	Query            string  `json:"query,omitempty"`
	Country          string  `json:"country,omitempty"`
	Device           string  `json:"device,omitempty"`
	SearchAppearance string  `json:"searchAppearance,omitempty"`
	SearchType       string  `json:"searchType,omitempty"`
	DataState        string  `json:"dataState,omitempty"`
	Provisional      bool    `json:"provisional,omitempty"`
	Clicks           float64 `json:"clicks"`
	Impressions      float64 `json:"impressions"`
	CTR              float64 `json:"ctr"`
	Position         float64 `json:"position"`
	PulledAt         string  `json:"pulledAt"`
}

type InspectionRecord struct {
	gsc.InspectionResult
	Issues      []string `json:"issues,omitempty"`
	InspectedAt string   `json:"inspectedAt"`
}

type AnalyticsFilter struct {
	Site       string
	Date       string
	Hour       string
	Page       string
	Query      string
	SearchType string
	DataState  string
}

type fileData struct {
	Analytics   []AnalyticsRecord  `json:"analytics"`
	PullLogs    []PullLog          `json:"pullLogs"`
	Inspections []InspectionRecord `json:"inspections"`
}

func New(path string) *Cache { return &Cache{path: path} }

func (c *Cache) Path() string { return c.path }

func (c *Cache) Status() (Status, error) {
	data, err := c.load()
	if err != nil {
		return Status{}, err
	}
	return Status{Path: c.path, AnalyticsRows: len(data.Analytics), PullLogs: len(data.PullLogs), Inspections: len(data.Inspections)}, nil
}

func (c *Cache) SaveAnalytics(hash string, req *gscapi.SearchAnalyticsQueryRequest, result *gsc.AnalyticsResult, meta gsc.QueryAllMetadata, provisional bool) error {
	if req == nil || result == nil {
		return fmt.Errorf("save analytics: request and result are required")
	}
	data, err := c.load()
	if err != nil {
		return err
	}
	pulledAt := time.Now().UTC().Format(time.RFC3339)
	for _, row := range result.Rows {
		record := AnalyticsRecord{Site: result.Site, SearchType: req.SearchType, DataState: req.DataState, Provisional: provisional, Clicks: row.Clicks, Impressions: row.Impressions, CTR: row.CTR, Position: row.Position, PulledAt: pulledAt}
		applyKeys(&record, result.Dimensions, row.Keys)
		index := indexAnalytics(data.Analytics, record)
		if index >= 0 {
			data.Analytics[index] = record
		} else {
			data.Analytics = append(data.Analytics, record)
		}
	}
	data.PullLogs = append(data.PullLogs, PullLog{Hash: hash, Site: result.Site, StartDate: req.StartDate, EndDate: req.EndDate, Dimensions: strings.Join(result.Dimensions, ","), SearchType: req.SearchType, DataState: req.DataState, RowCount: meta.RowCount, PagesFetched: meta.PagesFetched, CreatedAt: pulledAt})
	return c.save(data)
}

func (c *Cache) FindAnalytics(filter AnalyticsFilter) ([]AnalyticsRecord, error) {
	data, err := c.load()
	if err != nil {
		return nil, err
	}
	var rows []AnalyticsRecord
	for _, row := range data.Analytics {
		if filter.Site != "" && row.Site != filter.Site {
			continue
		}
		if filter.Date != "" && row.Date != filter.Date {
			continue
		}
		if filter.Hour != "" && row.Hour != filter.Hour {
			continue
		}
		if filter.Page != "" && row.Page != filter.Page {
			continue
		}
		if filter.Query != "" && row.Query != filter.Query {
			continue
		}
		if filter.SearchType != "" && row.SearchType != filter.SearchType {
			continue
		}
		if filter.DataState != "" && row.DataState != filter.DataState {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Date != rows[j].Date {
			return rows[i].Date < rows[j].Date
		}
		if rows[i].Hour != rows[j].Hour {
			return rows[i].Hour < rows[j].Hour
		}
		if rows[i].Page != rows[j].Page {
			return rows[i].Page < rows[j].Page
		}
		return rows[i].Query < rows[j].Query
	})
	return rows, nil
}

func (c *Cache) SaveInspection(result gsc.InspectionResult, issues []string) error {
	data, err := c.load()
	if err != nil {
		return err
	}
	data.Inspections = append(data.Inspections, InspectionRecord{InspectionResult: result, Issues: append([]string(nil), issues...), InspectedAt: time.Now().UTC().Format(time.RFC3339)})
	return c.save(data)
}

func (c *Cache) load() (*fileData, error) {
	if strings.TrimSpace(c.path) == "" {
		return &fileData{}, nil
	}
	blob, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileData{}, nil
		}
		return nil, fmt.Errorf("read cache: %w", err)
	}
	if len(blob) == 0 {
		return &fileData{}, nil
	}
	var data fileData
	if err := json.Unmarshal(blob, &data); err != nil {
		return nil, fmt.Errorf("decode cache: %w", err)
	}
	return &data, nil
}

func (c *Cache) save(data *fileData) error {
	if strings.TrimSpace(c.path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	blob, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	return os.WriteFile(c.path, blob, 0o600)
}

func applyKeys(record *AnalyticsRecord, dimensions []string, keys []string) {
	for i, dimension := range dimensions {
		if i >= len(keys) {
			return
		}
		switch dimension {
		case "DATE":
			record.Date = keys[i]
		case "HOUR":
			record.Hour = keys[i]
		case "PAGE":
			record.Page = keys[i]
		case "QUERY":
			record.Query = keys[i]
		case "COUNTRY":
			record.Country = keys[i]
		case "DEVICE":
			record.Device = keys[i]
		case "SEARCH_APPEARANCE":
			record.SearchAppearance = keys[i]
		}
	}
}

func indexAnalytics(rows []AnalyticsRecord, target AnalyticsRecord) int {
	for i, row := range rows {
		if row.Site == target.Site && row.Date == target.Date && row.Hour == target.Hour && row.Page == target.Page && row.Query == target.Query && row.Country == target.Country && row.Device == target.Device && row.SearchAppearance == target.SearchAppearance && row.SearchType == target.SearchType {
			return i
		}
	}
	return -1
}
