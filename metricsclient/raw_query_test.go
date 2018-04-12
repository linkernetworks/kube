package metricsclient

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/kubernetes/metricsclient/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRawQuery(t *testing.T) {

	ic := testutils.NewMockInfluxDBClient()
	assert.NotNil(t, ic)

	results, err := rawQuery(ic, "k8s", "SHOW STATS")

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}
