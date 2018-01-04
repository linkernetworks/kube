package nodesync

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"log"
)

const DevicePluginNvidiaGPU corev1.ResourceName = "nvidia.com/gpu"

func GetNvidiaGPU(r *corev1.ResourceList) *resource.Quantity {
	if val, ok := (*r)[corev1.ResourceNvidiaGPU]; ok {
		log.Printf("Use default resource name: %s\n", corev1.ResourceNvidiaGPU)
		return &val
	}
	if val, ok := (*r)[DevicePluginNvidiaGPU]; ok {
		log.Println("Use Nvidia device plugin resource name: %s\n", DevicePluginNvidiaGPU)
		return &val
	}
	return &resource.Quantity{}
}
