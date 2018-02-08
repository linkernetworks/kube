package podproxy

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"

	v1 "k8s.io/api/core/v1"
)

type ProxyInfoProvider interface {
	Host() string
	Port() string
	BaseURL() string
}

func NewPodInfo(pod *v1.Pod) *entity.PodInfo {
	return &entity.PodInfo{
		Phase:     pod.Status.Phase,
		Message:   pod.Status.Message,
		Reason:    pod.Status.Reason,
		StartTime: pod.Status.StartTime,
	}
}
