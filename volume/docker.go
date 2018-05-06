package volume

import (
	v1 "k8s.io/api/core/v1"
)

var DockerVolume = v1.Volume{
	Name: "docker-sock",
	VolumeSource: v1.VolumeSource{
		HostPath: &v1.HostPathVolumeSource{
			Path: "/var/run/docker.sock",
		},
	},
}

var DockerVolumeMount = v1.VolumeMount{
	Name:      "docker-sock",
	MountPath: "/var/run/docker.sock",
}
