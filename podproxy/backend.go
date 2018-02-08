package podproxy

import (
	"fmt"

	"bitbucket.org/linkernetworks/aurora/src/entity"

	v1 "k8s.io/api/core/v1"
)

// SelectPodContainerPort selects the container port from the given port by the port name
// This method is called by NewProxyBackendFromPodStatus
func SelectPodContainerPort(pod *v1.Pod, portname string) (containerPort int32, found bool) {
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

// NewProxyBackendFromPodStatus creates the proxy backend struct from the pod object.
func NewProxyBackendFromPodStatus(pod *v1.Pod, portname string) (*entity.ProxyBackend, error) {
	port, ok := SelectPodContainerPort(pod, portname)
	if !ok {
		return nil, fmt.Errorf("portname %s not found", portname)
	}
	return &entity.ProxyBackend{
		IP:        pod.Status.PodIP,
		Port:      int(port),
		Connected: pod.Status.PodIP != "",
	}, nil
}
