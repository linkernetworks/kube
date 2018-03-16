package volumechecker

import (
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"errors"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PrefixPodName = "fs-check-"

var ErrMountUnAvailable = errors.New("Volume Unavailable")

func NewVolume(containerVolumes []container.Volume) []v1.Volume {
	volumes := []v1.Volume{}
	for _, v := range containerVolumes {
		volumes = append(volumes, v1.Volume{
			Name: v.VolumeMount.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: v.ClaimName,
				},
			},
		})
	}
	return volumes
}

func NewVolumeMount(containerVolumes []container.Volume) []v1.VolumeMount {
	mounts := []v1.VolumeMount{}
	for _, v := range containerVolumes {
		mounts = append(mounts, v1.VolumeMount{
			Name:      v.VolumeMount.Name,
			MountPath: v.VolumeMount.MountPath,
			SubPath:   v.VolumeMount.SubPath,
		})
	}
	return mounts
}

func NewVolumeCheckPod(id string, volume []container.Volume) v1.Pod {
	volumes := NewVolume(volume)
	volumeMounts := NewVolumeMount(volume)
	name := PrefixPodName + id
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			Containers: []v1.Container{{
				Image:           "busybox:latest",
				Name:            name,
				ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
				VolumeMounts:    volumeMounts,
				Command:         []string{"sleep", "100"},
			},
			},
			Volumes: volumes,
		},
	}
}

/*
	for select {
		o <--.

	}

*/
func Check(ch chan *v1.Pod, podName string, timeout int) error {
	//We return nil iff the POD's status is running within timeout seconds.
	find := false
	ticker := time.NewTicker(time.Duration(timeout) * time.Second)
Watch:
	for {
		select {
		case pod := <-ch:
			if pod.ObjectMeta.Name != podName {
				continue
			}

			if v1.PodRunning == pod.Status.Phase {
				find = true
				break Watch
			}
		case <-ticker.C:
			break Watch
		}
	}

	if !find {
		return ErrMountUnAvailable
	}
	return nil
}
