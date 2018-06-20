package metricsclient

import (
	"testing"

	"github.com/linkernetworks/kube/metricsclient/testutils"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/stretchr/testify/assert"
)

func newTestMetricsClient(t *testing.T) *InfluxMetricsClient {
	t.Helper()

	ic := testutils.NewMockInfluxDBClient()
	assert.NotNil(t, ic)

	return NewForInfluxdb(ic)
}

// InfluxMetricsClient must implement MetricsClient
func TestImplementation(t *testing.T) {
	c := newTestMetricsClient(t)

	// if it's not implemented, build error
	var _ MetricsClient = c
}

func TestNewForInfluxdb(t *testing.T) {
	// This is test URL and the constructor won't dial it, no need to be reachable
	var influxURL = "http://monitoring-influxdb.kube-system:8086"
	ic, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: influxURL,
	})
	assert.NoError(t, err)

	c := NewForInfluxdb(ic)
	assert.NotNil(t, c)
}

func TestUse(t *testing.T) {
	c := newTestMetricsClient(t)

	var db = "test"
	err := c.Use(db)

	assert.NoError(t, err)
	assert.Equal(t, db, c.db)
}

func TestQueryNamespaces(t *testing.T) {
	c := newTestMetricsClient(t)

	ns, err := c.QueryNamespaces()

	assert.NoError(t, err)
	assert.NotEmpty(t, ns)
	assert.Contains(t, ns, "default")
	assert.Contains(t, ns, "kube-system")
}

func TestQueryNodes(t *testing.T) {
	c := newTestMetricsClient(t)

	nodes, err := c.QueryNodes()

	assert.NoError(t, err)
	assert.NotEmpty(t, nodes)
	assert.Contains(t, nodes, "docker-for-desktop")
}

func TestQueryNodeCPUUsages(t *testing.T) {
	c := newTestMetricsClient(t)

	usages, err := c.QueryNodeCPUUsages("docker-for-desktop", 10)

	assert.NoError(t, err)
	assert.NotEmpty(t, usages)
	assert.Len(t, usages, 10)
}

func TestQueryLastestNodeCPU(t *testing.T) {
	c := newTestMetricsClient(t)

	tim, usage, err := c.QueryLastestNodeCPU("docker-for-desktop")

	assert.NoError(t, err)
	assert.NotNil(t, tim)
	assert.True(t, usage > 0)
}

func TestQueryNodeMemoryUsages(t *testing.T) {
	c := newTestMetricsClient(t)

	usages, err := c.QueryNodeMemoryUsages("docker-for-desktop", 10)

	assert.NoError(t, err)
	assert.NotEmpty(t, usages)
	assert.Len(t, usages, 10)
}

func TestQueryLastestNodeMemory(t *testing.T) {
	c := newTestMetricsClient(t)

	tim, usage, err := c.QueryLastestNodeMemory("docker-for-desktop")

	assert.NoError(t, err)
	assert.NotNil(t, tim)
	assert.True(t, usage > 0)
}

func TestQueryPods(t *testing.T) {
	c := newTestMetricsClient(t)

	pods, err := c.QueryPods("default")

	assert.NoError(t, err)
	assert.NotEmpty(t, pods)
	assert.Contains(t, pods, "mongo-0")
}

func TestQueryPodCPUUsages(t *testing.T) {
	c := newTestMetricsClient(t)

	usages, err := c.QueryPodCPUUsages("default", "mongo-0", 10)

	assert.NoError(t, err)
	assert.NotEmpty(t, usages)
	assert.Len(t, usages, 10)
}

func TestQueryLastestPodCPU(t *testing.T) {
	c := newTestMetricsClient(t)

	tim, usage, err := c.QueryLastestPodCPU("default", "mongo-0")

	assert.NoError(t, err)
	assert.NotNil(t, tim)
	assert.True(t, usage > 0)
}

func TestQueryPodMemoryUsages(t *testing.T) {
	c := newTestMetricsClient(t)

	usages, err := c.QueryPodMemoryUsages("default", "mongo-0", 10)

	assert.NoError(t, err)
	assert.Len(t, usages, 10)
}

func TestQueryLastestPodMemory(t *testing.T) {
	c := newTestMetricsClient(t)

	tim, usage, err := c.QueryLastestPodMemory("default", "mongo-0")

	assert.NoError(t, err)
	assert.NotNil(t, tim)
	assert.True(t, usage > 0)
}

func TestQueryContainers(t *testing.T) {
	c := newTestMetricsClient(t)

	containers, err := c.QueryContainers("default", "mongo-0")

	assert.NoError(t, err)
	assert.NotEmpty(t, containers)
	assert.Contains(t, containers, "mongo")
	assert.Contains(t, containers, "mongo-sidecar")
}

func TestQueryContainerCPUUsages(t *testing.T) {
	c := newTestMetricsClient(t)

	usages, err := c.QueryContainerCPUUsages("default", "mongo-0", "mongo-sidecar", 10)

	assert.NoError(t, err)
	assert.NotEmpty(t, usages)
	assert.Len(t, usages, 10)
}

func TestQueryLastestContainerCPU(t *testing.T) {
	c := newTestMetricsClient(t)

	tim, usage, err := c.QueryLastestContainerCPU("default", "mongo-0", "mongo-sidecar")

	assert.NoError(t, err)
	assert.NotNil(t, tim)
	assert.True(t, usage > 0)
}

func TestQueryContainerMemoryUsages(t *testing.T) {
	c := newTestMetricsClient(t)

	usages, err := c.QueryContainerMemoryUsages("default", "mongo-0", "mongo-sidecar", 10)

	assert.NoError(t, err)
	assert.NotEmpty(t, usages)
	assert.Len(t, usages, 10)
}

func TestQueryLastestContainerMemory(t *testing.T) {
	c := newTestMetricsClient(t)

	tim, usage, err := c.QueryLastestContainerMemory("default", "mongo-0", "mongo-sidecar")

	assert.NoError(t, err)
	assert.NotNil(t, tim)
	assert.True(t, usage > 0)
}

func TestClose(t *testing.T) {
	c := newTestMetricsClient(t)

	err := c.Close()
	assert.NoError(t, err)
}
