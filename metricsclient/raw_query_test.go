package metricsclient

import (
	"testing"

	"github.com/linkernetworks/kube/metricsclient/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRawQuery(t *testing.T) {

	ic := testutils.NewMockInfluxDBClient()
	assert.NotNil(t, ic)

	results, err := rawQuery(ic, "k8s", "SHOW STATS")

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}
