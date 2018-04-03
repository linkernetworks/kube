package jobtracker

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"

	"github.com/stretchr/testify/assert"

	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func TestTrackUntilCompletion(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip()
		return
	}

	var backoffLimit int32 = 0
	job := batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "testing",
			GenerateName: "test-completion-job-",
			Labels:       map[string]string{},
		},
		Spec: batch.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: v1.PodTemplateSpec{

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
			},
		},
	}

	cf := config.MustRead("../../../../config/testing.json")
	k := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := k.NewClientset()
	assert.NoError(t, err)

	tracker := JobStatusTracker{Clientset: clientset}

	created, err := clientset.BatchV1().Jobs("testing").Create(&job)
	assert.NoError(t, err)
	t.Logf("job created: job=%s", created.Name)
	defer clientset.BatchV1().Jobs("testing").Delete(created.Name, nil)

	in := tracker.TrackUntilCompletion("testing", fields.ParseSelectorOrDie("metadata.name="+created.Name))
	for message := range in {
		t.Logf("message: %+v", message)
	}
	// the watcher listen to the stop channel, we need to close the channel to stop the
	// jitter wait loop
	tracker.Stop()
}
