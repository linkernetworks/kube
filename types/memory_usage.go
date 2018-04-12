package types

import "time"

// MemoryUsage is a simple wrapper for memory usage statistics from InfluxDB quering results
type MemoryUsage struct {
	Timestamp time.Time // time in RFC3339
	Usage     float64   // memory usage in Bytes
}
