package volume

import (
	"bitbucket.org/linkernetworks/aurora/src/types/container"

	v1 "k8s.io/api/core/v1"
)

func NewVolume(def *container.Volume) v1.Volume {
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

// NewVolumes creates the kubernetes volume definition from the container
// volume def used by the pod spec.
func NewVolumes(defs []container.Volume) (volumes []v1.Volume) {
	for _, def := range defs {
		volumes = append(volumes, NewVolume(&def))
	}

	return volumes
}

// NewVolumeMounts creates the kubernetes volume mount definition from the
// container volume def, it uses the defined volumes
func NewVolumeMount(def *container.Volume) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      def.VolumeMount.Name,
		SubPath:   def.VolumeMount.SubPath,
		MountPath: def.VolumeMount.MountPath,
	}
}

// NewVolumeMounts creates the kubernetes volume mount definition from the
// container volume def, it uses the defined volumes
func NewVolumeMounts(defs []container.Volume) (mounts []v1.VolumeMount) {
	for _, def := range defs {
		mounts = append(mounts, NewVolumeMount(&def))
	}

	return mounts
}
