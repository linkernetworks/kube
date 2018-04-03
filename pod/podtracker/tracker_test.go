package podtracker

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func TestTrackUntilCompletion(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "testing",
			GenerateName: "test-completion-pod-",
			// Name:      "test-completion-pod",
			Labels: map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "i-will-sleep",
					Image:   "alpine:3.7",
					Command: []string{"/bin/sh", "-c", "sleep 1"},
				},
			},
			RestartPolicy: "Never",
		},
	}

	cf := config.MustRead("../../../../config/testing.json")
	k := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := k.NewClientset()
	assert.NoError(t, err)

	tracker := PodStatusTracker{Clientset: clientset}

	created, err := clientset.Core().Pods("testing").Create(&pod)
	assert.NoError(t, err)
	t.Logf("pod created: pod=%s", created.Name)
	defer clientset.Core().Pods("testing").Delete(created.Name, nil)

	in, _ := tracker.TrackUntilCompletion("testing", fields.ParseSelectorOrDie("metadata.name="+created.Name))
	for message := range in {
		t.Logf("%+v", message)
	}
}
