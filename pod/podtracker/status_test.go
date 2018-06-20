package podtracker

import (
	"os"
	"testing"

	"github.com/linkernetworks/config"
	"github.com/linkernetworks/service/kubernetes"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func TestTrackUntilCompletion(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip()
		return
	}

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "testing",
			GenerateName: "test-completion-pod-",
			Labels:       map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "test",
					Image:   "alpine:3.7",
					Command: []string{"/bin/sh", "-c", "echo hello"},
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

	in := tracker.TrackUntilCompletion("testing", fields.ParseSelectorOrDie("metadata.name="+created.Name))
	for message := range in {
		t.Logf("message: %+v", message)
		assert.NotEmpty(t, message.Phase)
		assert.Nil(t, message.Error)
	}
	tracker.Stop()
}
