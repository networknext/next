package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// MetricType represents the different types of metrics that the backend uses
type MetricType int

const (
	MetricTypeUnknown MetricType = 0
	MetricTypeCounter MetricType = 1
	MetricTypeGauge   MetricType = 2
)

// MetricTypeText converts the metric type to text
func MetricTypeText(metricType MetricType) string {
	switch metricType {
	case MetricTypeCounter:
		return "counter"
	case MetricTypeGauge:
		return "gauge"
	default:
		return "unknown"
	}
}

// ParseMetricType parses a string representation to its corresponding metric type
func ParseMetricType(metricType string) MetricType {
	switch metricType {
	case "counter":
		return MetricTypeCounter
	case "gauge":
		return MetricTypeGauge
	default:
		return MetricTypeUnknown
	}
}

// MetricsDashboard contains all metric data for a single dashboard in StackDriver
type MetricsDashboard struct {
	ID          string         `json:"id"`          // A unique ID for the dashboard that is found in the URL
	Etag        string         `json:"etag"`        // A unique tag that changes each time the dashboard is updated
	DisplayName string         `json:"displayName"` // The display name for the dashboard
	Columns     string         `json:"columns"`     // The number of charts the dashboard shows per column
	Charts      []MetricsChart `json:"charts"`      // The list of charts
}

// MarshalCompleteDashboardJSON generates the full JSON representation of the metrics dashboard
// that the gcloud CLI expects to receive for creating and updating dashboards.
// Note that this is NOT the same "compact" JSON schema that the next tool uses.
// The "compact" JSON schema is handled by the default implementation of json.Marshal()
func (d MetricsDashboard) MarshalCompleteDashboardJSON() (string, error) {
	charts := []interface{}{}

	for _, chart := range d.Charts {
		metrics := []interface{}{}

		for _, metric := range chart.Metrics {
			var aggregatorType string

			if metric.MetricType == MetricTypeCounter {
				aggregatorType = "REDUCE_SUM"
			} else {
				aggregatorType = "REDUCE_MAX"
			}

			m := map[string]interface{}{
				"minAlignmentPeriod": "60s",
				"plotType":           "LINE",
				"timeSeriesQuery": map[string]interface{}{
					"timeSeriesFilter": map[string]interface{}{
						"aggregation": map[string]interface{}{
							"crossSeriesReducer": aggregatorType,
							"perSeriesAligner":   "ALIGN_MEAN",
						},
						"filter": fmt.Sprintf("metric.type=\"custom.googleapis.com/%s\" resource.type=\"gce_instance\"", metric.ID),
					},
				},
			}

			metrics = append(metrics, m)
		}

		c := map[string]interface{}{
			"title": chart.Title,
			"xyChart": map[string]interface{}{
				"chartOptions": map[string]interface{}{
					"mode": "COLOR",
				},
				"dataSets":          metrics,
				"timeshiftDuration": "0s",
				"yAxis": map[string]interface{}{
					"label": "y1Axis",
					"scale": "LINEAR",
				},
			},
		}

		charts = append(charts, c)
	}

	fields := map[string]interface{}{
		"displayName": d.DisplayName,
		"gridLayout": map[string]interface{}{
			"columns": d.Columns,
			"widgets": charts,
		},
		"name": "projects/network-next-v3-stackdriver-ws/dashboards/" + d.ID,
	}

	if len(d.Etag) > 0 {
		fields["etag"] = d.Etag
	}

	result, err := json.Marshal(&fields)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// UnmarshalCompleteDashboardJSON unmarshals the JSON from the gcloud CLI to a metrics dashboard type.
// Note that this unmarshals the complete JSON schma, not the "compact" schema the next tool uses.
// The "compact" JSON schema is handled by the default implementation of json.Unmarshal()
func (d *MetricsDashboard) UnmarshalCompleteDashboardJSON(fields map[string]interface{}) error {
	idString := fields["name"].(string)
	slashIndex := strings.LastIndex(idString, "/")
	if slashIndex == -1 {
		return errors.New("dashboard ID has bad form")
	}

	d.ID = idString[slashIndex+1:]
	d.Etag = fields["etag"].(string)
	d.DisplayName = fields["displayName"].(string)
	d.Columns = fields["gridLayout"].(map[string]interface{})["columns"].(string)

	var charts []MetricsChart
	chartsInterfaceArray := fields["gridLayout"].(map[string]interface{})["widgets"].([]interface{})
	for i := 0; i < len(chartsInterfaceArray); i++ {
		chartMap := chartsInterfaceArray[i].(map[string]interface{})

		chart := MetricsChart{
			Title: chartMap["title"].(string),
		}

		var metrics []Metric
		metricsInterfaceArray := chartMap["xyChart"].(map[string]interface{})["dataSets"].([]interface{})
		for j := 0; j < len(metricsInterfaceArray); j++ {
			metricMap := metricsInterfaceArray[j].(map[string]interface{})["timeSeriesQuery"].(map[string]interface{})["timeSeriesFilter"].(map[string]interface{})

			var metricType MetricType
			switch metricMap["aggregation"].(map[string]interface{})["crossSeriesReducer"].(string) {
			case "REDUCE_SUM":
				metricType = MetricTypeCounter

			case "REDUCE_MAX":
				metricType = MetricTypeGauge
			}

			filter := metricMap["filter"].(string)
			filter = strings.TrimPrefix(filter, "metric.type=\"custom.googleapis.com/")
			filter = strings.TrimSuffix(filter, "\" resource.type=\"gce_instance\"")

			metric := Metric{
				MetricType: metricType,
				ID:         filter,
			}

			metrics = append(metrics, metric)
		}

		chart.Metrics = metrics

		charts = append(charts, chart)
	}

	d.Charts = charts
	return nil
}

// UnmarshalCompleteDashboardArrayJSON unmarshals the JSON from the gcloud CLI to an array of metrics dashboards.
// Note that this unmarshals the complete JSON schma, not the "compact" schema the next tool uses.
func UnmarshalCompleteDashboardArrayJSON(data []byte) ([]MetricsDashboard, error) {
	dashboardInterfaceArray := []interface{}{}
	if err := json.Unmarshal(data, &dashboardInterfaceArray); err != nil {
		return nil, err
	}

	var dashboards []MetricsDashboard
	for i := 0; i < len(dashboardInterfaceArray); i++ {
		var dashboard MetricsDashboard
		if err := dashboard.UnmarshalCompleteDashboardJSON(dashboardInterfaceArray[i].(map[string]interface{})); err != nil {
			return nil, err
		}

		dashboards = append(dashboards, dashboard)
	}

	return dashboards, nil
}

// MetricsChart represents a single chart on a metrics dashboard
type MetricsChart struct {
	Title   string   `json:"title"`   // The name of the chart
	Metrics []Metric `json:"metrics"` // The metrics within the chart
}

// Metric represents a single metric within a chart in a dashboard
type Metric struct {
	MetricType MetricType `json:"type"` // The type of metric, this will impact how the data is aggregated
	ID         string     `json:"id"`   // The metric's unique ID
}

func (m Metric) MarshalJSON() ([]byte, error) {
	fields := map[string]interface{}{
		"type": MetricTypeText(m.MetricType),
		"id":   m.ID,
	}

	return json.Marshal(fields)
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	fields := map[string]interface{}{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	metricTypeString, ok := fields["type"].(string)
	if !ok {
		return errors.New("type is not a string type")
	}

	metricType := ParseMetricType(metricTypeString)
	if metricType == MetricTypeUnknown {
		return fmt.Errorf("unknown metric type %q", metricTypeString)
	}
	m.MetricType = metricType

	m.ID, ok = fields["id"].(string)
	if !ok {
		return errors.New("id is not a string type")
	}

	return nil
}

func getMetricsDashboards(dashboardFilter string) ([]MetricsDashboard, error) {
	var success bool
	var output string

	if len(dashboardFilter) > 0 {
		success, output = bashQuiet(fmt.Sprintf("gcloud monitoring dashboards list --project=network-next-v3-stackdriver-ws --format=json --filter=\"name:%s\"", dashboardFilter))
		if !success {
			return nil, fmt.Errorf("could not list dashboards: %s", output)
		}
	} else {
		success, output = bashQuiet("gcloud monitoring dashboards list --project=network-next-v3-stackdriver-ws --format=json")
		if !success {
			return nil, fmt.Errorf("could not list dashboards: %s", output)
		}
	}

	return UnmarshalCompleteDashboardArrayJSON([]byte(output))
}

func setMetricsDashboards(dashboards []MetricsDashboard) error {
	for _, dashboard := range dashboards {
		success, output := bashQuiet(fmt.Sprintf("gcloud monitoring dashboards list --project=network-next-v3-stackdriver-ws --format=json --filter=\"name:%s\"", dashboard.ID))
		if !success {
			return fmt.Errorf("could not create dashboard: %s", output)
		}

		if output == "Listed 0 items.\n" {
			// Create
			dashboardJSON, err := dashboard.MarshalCompleteDashboardJSON()
			if err != nil {
				return err
			}

			success, output := bashQuiet(fmt.Sprintf("gcloud monitoring dashboards create --project=network-next-v3-stackdriver-ws --config='''%s'''", dashboardJSON))
			if !success {
				return fmt.Errorf("could not create dashboard: %s", output)
			}

			fmt.Printf("Created dashboard %s\n", dashboard.ID)
		} else {
			// Update - need to use the same etag
			var responseDashboard []interface{}
			err := json.Unmarshal([]byte(output), &responseDashboard)
			if err != nil {
				return err
			}

			dashboard.Etag = responseDashboard[0].(map[string]interface{})["etag"].(string)

			dashboardJSON, err := dashboard.MarshalCompleteDashboardJSON()
			if err != nil {
				return err
			}

			success, output := bashQuiet(fmt.Sprintf("gcloud monitoring dashboards update %s --project=network-next-v3-stackdriver-ws --config='''%s'''", dashboard.ID, dashboardJSON))
			if !success {
				return fmt.Errorf("could not update dashboard %s: %s", dashboard.ID, output)
			}

			fmt.Printf("Updated dashboard %s\n", dashboard.ID)
		}
	}

	return nil
}
