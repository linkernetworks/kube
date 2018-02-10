package podutil

import (
	v1 "k8s.io/api/core/v1"
)

// SelectPodContainerPort selects the container port from the given port by the port name
// This method is called by NewProxyBackendFromPodStatus
func FindContainerPort(pod *v1.Pod, portname string) (containerPort int32, found bool) {
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == portname {
				containerPort = port.ContainerPort
				found = true
				return
			}
		}
	}
	return containerPort, found
}
