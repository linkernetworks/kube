package testutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

var (
	// ErrNotFound is returned when testdata do not contain the query situation
	ErrNotFound = errors.New("not found")
)

// MockInfluxDBClient emulates a InfluxDB and returns test data
type MockInfluxDBClient struct {
}

// NewMockInfluxDBClient creates a MockInfluxDBClient
func NewMockInfluxDBClient() *MockInfluxDBClient {
	return &MockInfluxDBClient{}
}

func (c MockInfluxDBClient) Close() error {
	return nil
}

func (c MockInfluxDBClient) Ping(d time.Duration) (time.Duration, string, error) {
	return time.Duration(0), "", nil
}

// Query loads testdata and find the match result
func (c MockInfluxDBClient) Query(q client.Query) (*client.Response, error) {
	b := bytes.NewReader(testdata)

	var records []DataRecord
	d := json.NewDecoder(b)
	if err := d.Decode(&records); err != nil {
		return nil, err
	}

	for _, r := range records {
		if reflect.DeepEqual(r.Query, q) {
			return &r.Resp, nil
		}
	}

	return nil, nil
}

func (c MockInfluxDBClient) Write(points client.BatchPoints) error {
	return nil
}
