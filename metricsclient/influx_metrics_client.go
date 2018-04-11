package metricsclient

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"

	client "github.com/influxdata/influxdb/client/v2"
)

// InfluxMetricsClient is a client which queries metrics of K8s Node/Pod/Container resource usage
// like CPU, memory, filesystem usage and network I/O from InfluxDB.
//
// When you use the batch methods (NodeCPUUsages, NodeMemUsages, PodCPUUsages,
// PodMemUsages, ContainerCPUUsages and ContainerMemUsages), remember the Heapster
// takes samples (采样) every 1min in current configuration, so for example if you
// call PodCPUUsages() and pass n=10 in the argument, you are asking the client
// 'I want the CPU usages of a Pod in the last ten minutes'. The results will
// contain 10 sequential records of CPU usage (if the InfluxDB has) in Millicores
// with timestamps in them. Results are orders by time, later ones in front.
type InfluxMetricsClient struct {
	influxc client.Client

	mtx sync.Mutex
	db  string
}

// NewForInfluxdb creates a InfluxMetricsClient
func NewForInfluxdb(c client.Client) *InfluxMetricsClient {
	return &InfluxMetricsClient{
		influxc: c,
		db:      "k8s", // by default
	}
}

// Use switches databases
func (c *InfluxMetricsClient) Use(db string) error {
	c.mtx.Lock()
	c.db = db
	c.mtx.Unlock()
	return nil
}

// QueryNamespaces searches the InfluxDB and returns the all Kubernetes namespaces of the cluster
func (c *InfluxMetricsClient) QueryNamespaces() ([]string, error) {
	sql := string(`SHOW TAG VALUES FROM "uptime" WITH KEY = "namespace_name"`)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	var namespaces []string
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				namespaces = append(namespaces, v[1].(string))
			}
		}
	}
	return namespaces, nil
}

// QueryNodes searches the InfluxDB and lists all nodes of the cluster
func (c *InfluxMetricsClient) QueryNodes() ([]string, error) {
	sql := string(`SHOW TAG VALUES FROM "uptime" WITH KEY = "nodename"`)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	var nodes []string
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				nodes = append(nodes, v[1].(string))
			}
		}
	}
	return nodes, nil
}

// QueryPods searches the InfluxDB and lists all Pods of a namespace
func (c *InfluxMetricsClient) QueryPods(namespace string) ([]string, error) {
	sql := fmt.Sprintf("SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"pod_name\" WHERE \"namespace_name\"='%s'", namespace)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	var pods []string
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				pods = append(pods, v[1].(string))
			}
		}
	}
	return pods, nil
}

// QueryPodCPUUsages searches the InfluxDB and returns a batch of last CPU usage records of a Pod
func (c *InfluxMetricsClient) QueryPodCPUUsages(namespace, pod string, limit int) ([]types.CPUUsage, error) {
	sql := fmt.Sprintf("SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = '%s' AND \"type\" = 'pod' AND \"pod_name\" = '%s' ORDER BY DESC LIMIT %d", namespace, pod, limit)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	return parseCPUUsages(results)
}

// QueryLastestPodCPU searches the InfluxDB and returns the last record of CPU usage (in Millicores) of a Pod
func (c *InfluxMetricsClient) QueryLastestPodCPU(namespace, pod string) (*time.Time, float64, error) {
	usages, err := c.QueryPodCPUUsages(namespace, pod, 1)
	if err != nil {
		return nil, 0.0, err
	}
	if len(usages) == 0 {
		return nil, 0.0, ErrNotFound
	}
	return &usages[0].Timestamp, usages[0].Usage, nil
}

// QueryPodMemUsages searches the InfluxDB and returns a batch of last memory usage records of a Pod
func (c *InfluxMetricsClient) QueryPodMemUsages(namespace, pod string, limit int) ([]types.MemUsage, error) {
	sql := fmt.Sprintf("SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = '%s' AND \"type\" = 'pod' AND \"pod_name\" = '%s' ORDER BY DESC LIMIT %d", namespace, pod, limit)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	return parseMemUsages(results)
}

// QueryLastestPodMem searches the InfluxDB and returns the last record of memory usage (in Bytes) of a Pod
func (c *InfluxMetricsClient) QueryLastestPodMem(namespace, pod string) (*time.Time, float64, error) {
	usages, err := c.QueryPodMemUsages(namespace, pod, 1)
	if err != nil {
		return nil, 0.0, err
	}
	if len(usages) == 0 {
		return nil, 0.0, ErrNotFound
	}
	return &usages[0].Timestamp, usages[0].Usage, nil
}

// QueryNodeCPUUsages searches the InfluxDB and returns a batch of last CPU usage records of a Node
func (c *InfluxMetricsClient) QueryNodeCPUUsages(node string, limit int) ([]types.CPUUsage, error) {
	sql := fmt.Sprintf("SELECT value FROM \"cpu/usage_rate\" WHERE \"type\" = 'node' AND \"nodename\" = '%s' ORDER BY DESC LIMIT %d", node, limit)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	return parseCPUUsages(results)
}

// QueryLastestNodeCPU searches the InfluxDB and returns the last record of CPU usage (in Millicores) of a Node
func (c *InfluxMetricsClient) QueryLastestNodeCPU(node string) (*time.Time, float64, error) {
	usages, err := c.QueryNodeCPUUsages(node, 1)
	if err != nil {
		return nil, 0.0, err
	}
	if len(usages) == 0 {
		return nil, 0.0, ErrNotFound
	}
	return &usages[0].Timestamp, usages[0].Usage, nil
}

// QueryNodeMemUsages searches the InfluxDB and returns a batch of last memory usage records of a Node
func (c *InfluxMetricsClient) QueryNodeMemUsages(node string, limit int) ([]types.MemUsage, error) {
	sql := fmt.Sprintf("SELECT value FROM \"memory/usage\" WHERE \"type\" = 'node' AND \"nodename\" = '%s' ORDER BY DESC LIMIT %d", node, limit)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	return parseMemUsages(results)
}

// QueryLastestNodeMem searches the InfluxDB and returns the last record of memory usage (in Bytes) of a Node
func (c *InfluxMetricsClient) QueryLastestNodeMem(node string) (*time.Time, float64, error) {
	usages, err := c.QueryNodeMemUsages(node, 1)
	if err != nil {
		return nil, 0.0, err
	}
	if len(usages) == 0 {
		return nil, 0.0, ErrNotFound
	}
	return &usages[0].Timestamp, usages[0].Usage, nil
}

// QueryContainers searches the InfluxDB and list all container names of a pod in specific namespace
func (c *InfluxMetricsClient) QueryContainers(namespace, pod string) ([]string, error) {
	sql := fmt.Sprintf("SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"container_name\" WHERE \"namespace_name\" = '%s' AND \"pod_name\" = '%s'", namespace, pod)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	var containers []string
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				val, ok := v[1].(string)
				if !ok {
					return nil, ErrTypeConvertion
				}
				containers = append(containers, val)
			}
		}
	}
	return containers, nil
}

// QueryContainerCPUUsages searches the InfluxDB and returns a batch of last CPU usage records of a Container
func (c *InfluxMetricsClient) QueryContainerCPUUsages(namespace, pod, container string, limit int) ([]types.CPUUsage, error) {
	sql := fmt.Sprintf("SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = '%s' AND \"pod_name\" = '%s' AND \"container_name\" = '%s' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT %d", namespace, pod, container, limit)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	var usages []types.CPUUsage
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				t, err := parseTime(v[0])
				if err != nil {
					return nil, err
				}
				val, err := parseValAsFloat(v[1])
				if err != nil {
					return nil, err
				}
				usages = append(usages, types.CPUUsage{Timestamp: t, Usage: val})
			}
		}
	}
	return usages, nil
}

// QueryLastestContainerCPU searches the InfluxDB and returns the last record of CPU usage (in Millicores) of a Container
func (c *InfluxMetricsClient) QueryLastestContainerCPU(namespace, pod, container string) (*time.Time, float64, error) {
	usages, err := c.QueryContainerCPUUsages(namespace, pod, container, 1)
	if err != nil {
		return nil, 0.0, err
	}
	if len(usages) == 0 {
		return nil, 0.0, ErrNotFound
	}
	return &usages[0].Timestamp, usages[0].Usage, nil
}

// QueryContainerMemUsages searches the InfluxDB and returns a batch of last memory usage records of a Container
func (c *InfluxMetricsClient) QueryContainerMemUsages(namespace, pod, container string, limit int) ([]types.MemUsage, error) {
	sql := fmt.Sprintf("SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = '%s' AND \"pod_name\" = '%s' AND \"container_name\" = '%s' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT %d", namespace, pod, container, limit)
	results, err := rawQuery(c.influxc, c.db, sql)
	if err != nil {
		return nil, err
	}
	var usages []types.MemUsage
	for _, r := range results {
		if r.Err != "" {
			return nil, errors.New(r.Err)
		}
		for _, s := range r.Series {
			for _, v := range s.Values {
				t, err := parseTime(v[0])
				if err != nil {
					return nil, err
				}
				val, err := parseValAsFloat(v[1])
				if err != nil {
					return nil, err
				}
				usages = append(usages, types.MemUsage{Timestamp: t, Usage: val})
			}
		}
	}
	return usages, nil
}

// QueryLastestContainerMem searches the InfluxDB and returns the last record of memory usage (in Bytes) of a Container
func (c *InfluxMetricsClient) QueryLastestContainerMem(namespace, pod, container string) (*time.Time, float64, error) {
	usages, err := c.QueryContainerMemUsages(namespace, pod, container, 1)
	if err != nil {
		return nil, 0.0, err
	}
	if len(usages) == 0 {
		return nil, 0.0, ErrNotFound
	}
	return &usages[0].Timestamp, usages[0].Usage, nil
}

// Close tears down the InfluxMetricsClient object
func (c *InfluxMetricsClient) Close() error {
	return c.influxc.Close()
}
