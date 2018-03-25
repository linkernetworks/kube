package volumechecker

import (
	"errors"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PrefixPodName = "fs-check-"

var ErrMountUnAvailable = errors.New("Volume Unavailable")

func NewVolumeCheckPod(id string) v1.Pod {
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
				Command:         []string{"sleep", "100"},
			},
			},
		},
	}
}

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
