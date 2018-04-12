package metricsclient

import (
	"time"

	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"
)

// MetricsClient is an abstract of client which can retrieve Node/Pod/Container metrics
type MetricsClient interface {
	// QueryNamespaces returns all namespaces of the cluster
	QueryNamespaces() ([]string, error)

	// QueryNodes returns names of all Nodes in the cluster
	QueryNodes() ([]string, error)
	// QueryNodeCPUUsages returns lastest n CPU usages of a Node
	QueryNodeCPUUsages(node string, n int) ([]types.CPUUsage, error)
	// QueryNodeMemoryUsages returns lastest n memory usages of a Node
	QueryNodeMemoryUsages(node string, n int) ([]types.MemoryUsage, error)
	// QueryLastestNodeCPU returns the most recent CPU usage of a Node
	QueryLastestNodeCPU(node string) (*time.Time, float64, error)
	// QueryLastestNodeMemory returns the most recent memory usage of a Node
	QueryLastestNodeMemory(node string) (*time.Time, float64, error)

	// QueryPods returns names of all Pods in the namespace
	QueryPods(namespace string) ([]string, error)
	// QueryPodCPUUsages returns lastest n CPU usages of a Pod
	QueryPodCPUUsages(namespace, pod string, n int) ([]types.CPUUsage, error)
	// QueryPodMemoryUsages returns lastest n memory usages of a Pod
	QueryPodMemoryUsages(namespace, pod string, n int) ([]types.MemoryUsage, error)
	// QueryLastestPodCPU returns the most recent CPU usage of a Pod
	QueryLastestPodCPU(namespace, pod string) (*time.Time, float64, error)
	// QueryLastestPodMemory returns the most recent memory usage of a Pod
	QueryLastestPodMemory(namespace, pod string) (*time.Time, float64, error)

	// QueryContainers returns names of all containers in a Pod
	QueryContainers(namespace, pod string) ([]string, error)
	// QueryContainerCPUUsages returns lastest n CPU usages of a Container
	QueryContainerCPUUsages(namespace, pod, container string, n int) ([]types.CPUUsage, error)
	// QueryContainerMemoryUsages returns lastest n memroy usages of a Container
	QueryContainerMemoryUsages(namespace, pod, container string, n int) ([]types.MemoryUsage, error)
	// QueryLastestContainerCPU returns the most recent CPU usage of a Container
	QueryLastestContainerCPU(namespace, pod, container string) (*time.Time, float64, error)
	// QueryLastestContainerMemory returns the most recent memroy usage of a Container
	QueryLastestContainerMemory(namespace, pod, container string) (*time.Time, float64, error)
}
