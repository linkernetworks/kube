package testutils

import (
	"testing"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/stretchr/testify/assert"
)

func TestImpletation(t *testing.T) {
	var _ client.Client = NewMockInfluxDBClient()
}

func TestNewMockInfluxDBClient(t *testing.T) {
	c := NewMockInfluxDBClient()

	assert.NotNil(t, c)
}

func TestQuery(t *testing.T) {
	c := NewMockInfluxDBClient()

	q := client.Query{
		Command:  "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"namespace_name\"",
		Database: "k8s",
	}

	resp, err := c.Query(q)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
