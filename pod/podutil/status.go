package podutil

import (
	v1 "k8s.io/api/core/v1"
)

// Find the container statuses that are in "Waiting"
func FindWaitingContainerStatuses(pod *v1.Pod) (cslist []v1.ContainerStatus) {
	for _, cs := range pod.Status.ContainerStatuses {
		// Skip the container status that is not waiting
		if cs.State.Waiting == nil {
			continue
		}
		cslist = append(cslist, cs)
	}
	return cslist
}
