package volume

import (
	"bitbucket.org/linkernetworks/aurora/src/types/container"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

// NewClaimBased creates the kubernetes volume object from the container volume
// definition.
func NewClaimBased(def *container.Volume) v1.Volume {
	return v1.Volume{
		Name: def.VolumeMount.Name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: def.ClaimName,
				ReadOnly:  false,
			},
		},
	}
}

// NewVolumesFromDef creates the kubernetes volume definition from the container
// volume def used by the pod spec.
func NewVolumesFromDef(defs []container.Volume) (volumes []v1.Volume) {
	for _, def := range defs {
		volumes = append(volumes, NewClaimBased(&def))
	}

	return volumes
}

// NewVolumeMounts creates the kubernetes volume mount definition from the
// container volume def, it uses the defined volumes
func NewVolumeMountFromDef(def *container.Volume) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      def.VolumeMount.Name,
		SubPath:   def.VolumeMount.SubPath,
		MountPath: def.VolumeMount.MountPath,
	}
}

// NewVolumeMountsFromDef creates the kubernetes volume mount definition from the
// container volume def, it uses the defined volumes
func NewVolumeMountsFromDef(defs []container.Volume) (mounts []v1.VolumeMount) {
	for _, def := range defs {
		mounts = append(mounts, NewVolumeMountFromDef(&def))
	}

	return mounts
}

func AttachVolumesToPod(defs []container.Volume, pod *v1.Pod) {
	pod.Spec.Volumes = append(pod.Spec.Volumes, NewVolumesFromDef(defs)...)
	for idx, container := range pod.Spec.Containers {
		pod.Spec.Containers[idx].VolumeMounts = append(container.VolumeMounts, NewVolumeMountsFromDef(defs)...)
	}
}

func AttachVolumeToPod(def *container.Volume, pod *v1.Pod) {
	pod.Spec.Volumes = append(pod.Spec.Volumes, NewClaimBased(def))
	for idx, container := range pod.Spec.Containers {
		pod.Spec.Containers[idx].VolumeMounts = append(container.VolumeMounts, NewVolumeMountFromDef(def))
	}
}

func AttachVolumesToJob(defs []container.Volume, job *batchv1.Job) {
	podSpec := &job.Spec.Template.Spec
	podSpec.Volumes = append(podSpec.Volumes, NewVolumesFromDef(defs)...)
	for idx, container := range podSpec.Containers {
		podSpec.Containers[idx].VolumeMounts = append(container.VolumeMounts, NewVolumeMountsFromDef(defs)...)
	}
}

func AttachVolumeToJob(def *container.Volume, job *batchv1.Job) {
	podSpec := &job.Spec.Template.Spec
	podSpec.Volumes = append(podSpec.Volumes, NewClaimBased(def))
	for idx, container := range podSpec.Containers {
		podSpec.Containers[idx].VolumeMounts = append(container.VolumeMounts, NewVolumeMountFromDef(def))
	}
}
