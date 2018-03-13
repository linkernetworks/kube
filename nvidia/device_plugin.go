package nvidia

import (
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const DevicePluginNvidiaGPU corev1.ResourceName = "nvidia.com/gpu"

func GetGPU(r *corev1.ResourceList) *resource.Quantity {
	if val, ok := (*r)[corev1.ResourceNvidiaGPU]; ok {
		log.Printf("default GPU resource: %s\n", corev1.ResourceNvidiaGPU)
		return &val
	}
	if val, ok := (*r)[DevicePluginNvidiaGPU]; ok {
		log.Printf("Nvidia device plugin resource: %s\n", DevicePluginNvidiaGPU)
		return &val
	}
	return &resource.Quantity{}
}
