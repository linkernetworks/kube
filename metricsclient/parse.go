package metricsclient

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/linkernetworks/kube/types"

	client "github.com/influxdata/influxdb/client/v2"
)

// parseMemoryUsages parses timestamps and CPU usages from raw SQL results
func parseMemoryUsages(results []client.Result) ([]types.MemoryUsage, error) {
	var memUsages []types.MemoryUsage
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				t, err := parseTime(v[0])
				if err != nil {
					return nil, err
				}
				val, err := parseValAsFloat(v[1])
				if err != nil {
					return nil, err
				}
				memUsages = append(memUsages, types.MemoryUsage{Timestamp: t, Usage: val})
			}
		}
	}
	return memUsages, nil
}

// parseCPUUsages parses timestamps and CPU usages from raw SQL results
func parseCPUUsages(results []client.Result) ([]types.CPUUsage, error) {
	var cpuUsages []types.CPUUsage
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				t, err := parseTime(v[0])
				if err != nil {
					return nil, err
				}
				val, err := parseValAsFloat(v[1])
				if err != nil {
					return nil, err
				}
				cpuUsages = append(cpuUsages, types.CPUUsage{Timestamp: t, Usage: val})
			}
		}
	}
	return cpuUsages, nil
}

func parseValAsFloat(v interface{}) (float64, error) {
	switch v.(type) {
	case float64:
		val, ok := v.(float64)
		if !ok {
			return 0.0, ErrTypeConvertion
		}
		return val, nil
	case json.Number:
		jn, ok := v.(json.Number)
		if !ok {
			return 0.0, ErrTypeConvertion
		}
		val, err := jn.Float64()
		if err != nil {
			return 0.0, err
		}
		return val, nil
	}
	return 0.0, ErrTypeConvertion
}

func parseTime(v interface{}) (time.Time, error) {
	val, ok := v.(string)
	if !ok {
		return time.Time{}, ErrTypeConvertion
	}
	return time.Parse(time.RFC3339, val)
}
