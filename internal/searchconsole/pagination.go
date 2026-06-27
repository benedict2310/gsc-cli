package searchconsole

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	gscapi "google.golang.org/api/searchconsole/v1"
)

const queryAllPageSize int64 = 25000

type QueryAllOptions struct {
	MaxRows int
}

type QueryAllMetadata struct {
	PagesFetched int  `json:"pagesFetched"`
	RowCount     int  `json:"rowCount"`
	Truncated    bool `json:"truncated"`
}

func (c *Client) QueryAnalyticsAll(ctx context.Context, site string, req *gscapi.SearchAnalyticsQueryRequest, opts QueryAllOptions) (*AnalyticsResult, QueryAllMetadata, error) {
	if req == nil {
		return nil, QueryAllMetadata{}, fmt.Errorf("query analytics all: request is required")
	}

	base := *req
	base.Dimensions = append([]string(nil), req.Dimensions...)
	base.DimensionFilterGroups = cloneFilterGroups(req.DimensionFilterGroups)
	base.RowLimit = queryAllPageSize
	base.StartRow = 0

	var (
		allRows      []AnalyticsRow
		pagesFetched int
		meta         QueryAllMetadata
	)
	for {
		current := base
		current.StartRow = int64(len(allRows))
		result, err := c.QueryAnalytics(ctx, site, &current)
		if err != nil {
			return nil, QueryAllMetadata{}, err
		}
		pagesFetched++
		allRows = append(allRows, result.Rows...)
		if opts.MaxRows > 0 && len(allRows) >= opts.MaxRows {
			allRows = append([]AnalyticsRow(nil), allRows[:opts.MaxRows]...)
			meta.Truncated = true
			meta.PagesFetched = pagesFetched
			meta.RowCount = len(allRows)
			result.Rows = allRows
			return result, meta, nil
		}
		if int64(len(result.Rows)) < queryAllPageSize {
			meta.PagesFetched = pagesFetched
			meta.RowCount = len(allRows)
			result.Rows = allRows
			return result, meta, nil
		}
	}
}

func AnalyticsRequestHash(site string, req *gscapi.SearchAnalyticsQueryRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("analytics request hash: request is required")
	}
	type filter struct {
		Dimension  string `json:"dimension,omitempty"`
		Expression string `json:"expression,omitempty"`
		Operator   string `json:"operator,omitempty"`
	}
	type filterGroup struct {
		GroupType string   `json:"groupType,omitempty"`
		Filters   []filter `json:"filters,omitempty"`
	}
	payload := struct {
		Site                  string        `json:"site"`
		StartDate             string        `json:"startDate,omitempty"`
		EndDate               string        `json:"endDate,omitempty"`
		Dimensions            []string      `json:"dimensions,omitempty"`
		Type                  string        `json:"type,omitempty"`
		DataState             string        `json:"dataState,omitempty"`
		AggregationType       string        `json:"aggregationType,omitempty"`
		DimensionFilterGroups []filterGroup `json:"dimensionFilterGroups,omitempty"`
	}{
		Site:            site,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Dimensions:      append([]string(nil), req.Dimensions...),
		Type:            req.SearchType,
		DataState:       req.DataState,
		AggregationType: req.AggregationType,
	}

	for _, group := range req.DimensionFilterGroups {
		if group == nil {
			continue
		}
		entry := filterGroup{GroupType: group.GroupType}
		for _, item := range group.Filters {
			if item == nil {
				continue
			}
			entry.Filters = append(entry.Filters, filter{Dimension: item.Dimension, Expression: item.Expression, Operator: item.Operator})
		}
		sort.Slice(entry.Filters, func(i, j int) bool {
			left := entry.Filters[i]
			right := entry.Filters[j]
			if left.Dimension != right.Dimension {
				return left.Dimension < right.Dimension
			}
			if left.Operator != right.Operator {
				return left.Operator < right.Operator
			}
			return left.Expression < right.Expression
		})
		payload.DimensionFilterGroups = append(payload.DimensionFilterGroups, entry)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("analytics request hash: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func cloneFilterGroups(groups []*gscapi.ApiDimensionFilterGroup) []*gscapi.ApiDimensionFilterGroup {
	cloned := make([]*gscapi.ApiDimensionFilterGroup, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		groupClone := &gscapi.ApiDimensionFilterGroup{GroupType: group.GroupType}
		for _, item := range group.Filters {
			if item == nil {
				continue
			}
			groupClone.Filters = append(groupClone.Filters, &gscapi.ApiDimensionFilter{Dimension: item.Dimension, Expression: item.Expression, Operator: item.Operator})
		}
		cloned = append(cloned, groupClone)
	}
	return cloned
}
