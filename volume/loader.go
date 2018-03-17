package volumes

import (
	"bitbucket.org/linkernetworks/aurora/src/types/container"

	v1 "k8s.io/api/core/v1"
)

// NewVolumes creates the kubernetes volume definition from the container
// volume def used by the pod spec.
func NewVolumes(defs []container.Volume) (volumes []v1.Volume) {
	for _, def := range defs {
		volumes = append(volumes, v1.Volume{
			Name: def.VolumeMount.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: def.ClaimName,
					ReadOnly:  false,
				},
			},
		})
	}

	return volumes
}

// NewVolumeMounts creates the kubernetes volume mount definition from the
// container volume def, it uses the defined volumes
func NewVolumeMounts(defs []container.Volume) (mounts []v1.VolumeMount) {
	for _, def := range defs {
		mounts = append(mounts, v1.VolumeMount{
			Name:      def.VolumeMount.Name,
			SubPath:   def.VolumeMount.SubPath,
			MountPath: def.VolumeMount.MountPath,
		})
	}

	return mounts
}
