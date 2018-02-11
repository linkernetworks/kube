package summary

import (
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/nvidia"
)

func QueryNodeGPUUsage(dt *deployment.KubeDeploymentTarget, name string) int64 {
	var totalReqGPU int64

	// FIXME should constraint to jobs not pods
	pods := dt.FetchActivePodsByNode(name)
	for _, p := range pods {
		for _, c := range p.Spec.Containers {
			totalReqGPU += nvidia.GetGPU(&c.Resources.Requests).Value()
		}
	}
	return totalReqGPU
}
