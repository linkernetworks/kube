package testutils

import (
	client "github.com/influxdata/influxdb/client/v2"
)

// DataRecord is the structure of an element in testdata JSON
type DataRecord struct {
	Cmd   string
	Query client.Query
	Resp  client.Response
}
