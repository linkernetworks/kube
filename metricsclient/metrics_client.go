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
	// QueryNodeMemUsages returns lastest n memory usages of a Node
	QueryNodeMemUsages(node string, n int) ([]types.MemUsage, error)
	// QueryLastestNodeCPU returns the most recent CPU usage of a Node
	QueryLastestNodeCPU(node string) (*time.Time, float64, error)
	// QueryLastestNodeMem returns the most recent memory usage of a Node
	QueryLastestNodeMem(node string) (*time.Time, float64, error)

	// QueryPods returns names of all Pods in the namespace
	QueryPods(namespace string) ([]string, error)
	// QueryPodCPUUsages returns lastest n CPU usages of a Pod
	QueryPodCPUUsages(namespace, pod string, n int) ([]types.CPUUsage, error)
	// QueryPodMemUsages returns lastest n memory usages of a Pod
	QueryPodMemUsages(namespace, pod string, n int) ([]types.MemUsage, error)
	// QueryLastestPodCPU returns the most recent CPU usage of a Pod
	QueryLastestPodCPU(namespace, pod string) (*time.Time, float64, error)
	// QueryLastestPodMem returns the most recent memory usage of a Pod
	QueryLastestPodMem(namespace, pod string) (*time.Time, float64, error)

	// QueryContainers returns names of all containers in a Pod
	QueryContainers(namespace, pod string) ([]string, error)
	// QueryContainerCPUUsages returns lastest n CPU usages of a Container
	QueryContainerCPUUsages(namespace, pod, container string, n int) ([]types.CPUUsage, error)
	// QueryContainerMemUsages returns lastest n memroy usages of a Container
	QueryContainerMemUsages(namespace, pod, container string, n int) ([]types.MemUsage, error)
	// QueryLastestContainerCPU returns the most recent CPU usage of a Container
	QueryLastestContainerCPU(namespace, pod, container string) (*time.Time, float64, error)
	// QueryLastestContainerMem returns the most recent memroy usage of a Container
	QueryLastestContainerMem(namespace, pod, container string) (*time.Time, float64, error)
}
