package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type MetricType int

const (
	MetricTypeUnknown MetricType = 0
	MetricTypeCounter MetricType = 1
	MetricTypeGauge   MetricType = 2
)

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

type MetricsDashboard struct {
	id          string
	displayName string
	columns     uint32
	charts      []MetricsChart
}

func (d *MetricsDashboard) UnmarshalJSON(data []byte) error {
	fields := map[string]interface{}{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var ok bool

	d.displayName, ok = fields["displayName"].(string)
	if !ok {
		return errors.New("displayName is not a string type")
	}

	columnsString, ok := fields["columns"].(string)
	if !ok {
		return errors.New("columns is not a string type")
	}

	columns, err := strconv.ParseUint(columnsString, 10, 32)
	if err != nil {
		return err
	}
	d.columns = uint32(columns)

	chartsArray, ok := fields["charts"].([]interface{})
	if !ok {
		return errors.New("charts is not an array type")
	}

	for i := 0; i < len(chartsArray); i++ {
		chartMap, ok := chartsArray[i].(map[string]interface{})
		if !ok {
			return errors.New("chart is not an object type")
		}

		chartJSON, err := json.Marshal(chartMap)
		if err != nil {
			return err
		}

		var chart MetricsChart
		if err := json.Unmarshal(chartJSON, &chart); err != nil {
			return err
		}
		d.charts = append(d.charts, chart)
	}

	d.id, ok = fields["id"].(string)
	if !ok {
		return errors.New("id is not a string type")
	}

	return nil
}

func (d MetricsDashboard) GenerateDashboardJSON(etag string) (string, error) {
	charts := []interface{}{}

	for _, chart := range d.charts {
		metrics := []interface{}{}

		for _, metric := range chart.metrics {
			var aggregatorType string

			if metric.metricType == MetricTypeCounter {
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
						"filter": fmt.Sprintf("metric.type=\"custom.googleapis.com/%s\" resource.type=\"gce_instance\"", metric.id),
					},
				},
			}

			metrics = append(metrics, m)
		}

		c := map[string]interface{}{
			"title": chart.title,
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
		"displayName": d.displayName,
		"gridLayout": map[string]interface{}{
			"columns": fmt.Sprintf("%d", d.columns),
			"widgets": charts,
		},
		"name": "projects/network-next-v3-stackdriver-ws/dashboards/" + d.id,
	}

	if len(etag) > 0 {
		fields["etag"] = etag
	}

	result, err := json.Marshal(&fields)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type MetricsChart struct {
	title   string
	metrics []Metric
}

func (c *MetricsChart) UnmarshalJSON(data []byte) error {
	fields := map[string]interface{}{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var ok bool

	c.title, ok = fields["title"].(string)
	if !ok {
		return errors.New("title is not a string type")
	}

	metricsArray, ok := fields["metrics"].([]interface{})
	if !ok {
		return errors.New("metrics is not an array type")
	}

	for i := 0; i < len(metricsArray); i++ {
		metricMap, ok := metricsArray[i].(map[string]interface{})
		if !ok {
			return errors.New("metric is not an object type")
		}

		metricJSON, err := json.Marshal(metricMap)
		if err != nil {
			return err
		}

		var metric Metric
		if err := json.Unmarshal(metricJSON, &metric); err != nil {
			return err
		}
		c.metrics = append(c.metrics, metric)
	}

	return nil
}

type Metric struct {
	metricType MetricType
	id         string
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
	m.metricType = metricType

	m.id, ok = fields["id"].(string)
	if !ok {
		return errors.New("id is not a string type")
	}

	return nil
}

func getMetricsDashboards() error {
	return nil
}

func setMetricsDashboards(dashboards []MetricsDashboard) error {
	for _, dashboard := range dashboards {
		success, output := bashQuiet(fmt.Sprintf("gcloud monitoring dashboards list --project=network-next-v3-stackdriver-ws --format=json --filter=\"name:%s\"", dashboard.id))
		if !success {
			return fmt.Errorf("could not create dashboard: %s", output)
		}

		if output == "Listed 0 items.\n" {
			// Create
			dashboardJSON, err := dashboard.GenerateDashboardJSON("")
			if err != nil {
				return err
			}

			success, output := bashQuiet(fmt.Sprintf("gcloud monitoring dashboards create --project=network-next-v3-stackdriver-ws --config='''%s'''", dashboardJSON))
			if !success {
				return fmt.Errorf("could not create dashboard: %s", output)
			}
		} else {
			// Update - need to use the same etag
			var responseDashboard []interface{}
			err := json.Unmarshal([]byte(output), &responseDashboard)
			if err != nil {
				return err
			}

			dashboardJSON, err := dashboard.GenerateDashboardJSON(responseDashboard[0].(map[string]interface{})["etag"].(string))
			if err != nil {
				return err
			}

			success, output := bashQuiet(fmt.Sprintf("gcloud monitoring dashboards update %s --project=network-next-v3-stackdriver-ws --config='''%s'''", dashboard.id, dashboardJSON))
			if !success {
				return fmt.Errorf("could not update dashboard %s: %s", dashboard.id, output)
			}
		}
	}

	return nil
}
