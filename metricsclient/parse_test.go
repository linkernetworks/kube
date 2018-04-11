package metricsclient

import (
	"encoding/json"
	"testing"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseCPUUsages(t *testing.T) {
	rawText := []byte(`
	[
	  {
	    "Series": [
	      {
	        "name": "cpu/usage_rate",
	        "columns": [
	          "time",
	          "value"
	        ],
	        "values": [
	          [
	            "2018-04-10T05:00:00Z",
	            0
	          ],
	          [
	            "2018-04-10T04:59:00Z",
	            1.23
	          ],
	          [
	            "2018-04-10T04:58:00Z",
	            472
	          ]
	        ]
	      }
	    ],
	    "Messages": null
	  }
	]
	`)
	var results []client.Result
	err := json.Unmarshal(rawText, &results)

	assert.NoError(t, err)
	assert.Len(t, results[0].Series[0].Values, 3)
}

func TestParseMemoryUsages(t *testing.T) {
	rawText := []byte(`
	[
	  {
	    "Series": [
	      {
	        "name": "memory/usage",
	        "columns": [
	          "time",
	          "value"
	        ],
	        "values": [
	          [
	            "2018-04-10T05:19:00Z",
	            2955399168
	          ],
	          [
	            "2018-04-10T05:18:00Z",
	            2950483968
	          ],
	          [
	            "2018-04-10T05:17:00Z",
	            2941882368
	          ]
	        ]
	      }
	    ],
	    "Messages": null
	  }
	]
	`)
	var results []client.Result
	err := json.Unmarshal(rawText, &results)
	assert.NoError(t, err)

	usages, err := parseMemoryUsages(results)

	assert.NoError(t, err)
	assert.Len(t, usages, 3)
	assert.NotZero(t, usages[0].Usage)
}

func TestParseValAsFloat(t *testing.T) {
	v := float64(1.23)

	val, err := parseValAsFloat(v)

	assert.NoError(t, err)
	assert.Equal(t, 1.23, val)
}

// 2: json.Number
func TestParseValAsFloat2(t *testing.T) {
	v := json.Number("0")

	val, err := parseValAsFloat(v)

	assert.NoError(t, err)
	assert.Equal(t, 0.0, val)
}

// 3: json.Number
func TestParseValAsFloat3(t *testing.T) {
	v := json.Number("1.23")

	val, err := parseValAsFloat(v)

	assert.NoError(t, err)
	assert.Equal(t, 1.23, val)
}

// 3: json.Number
func TestParseValAsFloat4(t *testing.T) {
	v := json.Number("123")

	val, err := parseValAsFloat(v)

	assert.NoError(t, err)
	assert.Equal(t, 123.0, val)
}
