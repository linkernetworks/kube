package types

import (
	v1 "k8s.io/api/core/v1"
)

// Object as Pod
type PodFactory interface {
	NewPod(podName string, labels map[string]string) v1.Pod
}

type DeploymentIDProvider interface {
	DeploymentID() string
}

type PodDeployment interface {
	DeploymentIDProvider
	PodFactory
}

type PodLabelProvider interface {
	PodLabels() map[string]string
}
