package nodesync

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const ResourceNvidiaGPU corev1.ResourceName = "nvidia.com/gpu"

func GetNvidiaGPU(r *corev1.ResourceList) *resource.Quantity {
	if val, ok := (*r)[ResourceNvidiaGPU]; ok {
		return &val
	}
	return &resource.Quantity{}
}
