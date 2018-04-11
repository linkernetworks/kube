package types

import "time"

// MemUsage is a simple wrapper for memory usage statistics from InfluxDB quering results
type MemUsage struct {
	Timestamp time.Time // time in RFC3339
	Usage     float64   // memory usage in Bytes
}
