package types

import (
	"time"
)

// CPUUsage is a simple wrapper for CPU usage statistics from InfluxDB quering results
type CPUUsage struct {
	Timestamp time.Time // time in RFC3339
	Usage     float64   // CPU usage in Millicores
}
