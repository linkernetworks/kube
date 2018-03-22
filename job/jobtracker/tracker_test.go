package jobtracker

import (
	"fmt"
	"sync"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/jobcontroller/types"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testingConfigPath = "../../../../config/testing.json"
	kubeNamespce      = "default"
)

func CreateKubernetesSleepJob(name string, seconds int) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "i-will-sleep",
							Image:   "alpine:latest",
							Command: []string{"/bin/sh", "-c", fmt.Sprintf("sleep %d", seconds)},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

}

func TestJobTracker(t *testing.T) {
	//Create a K8S Job
	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)
	assert.NotNil(t, clientset)

	id := bson.NewObjectId().Hex()
	kubejob := CreateKubernetesSleepJob(id, 2)

	resp, err := clientset.BatchV1().Jobs(kubeNamespce).Create(kubejob)

	//Wait Phase
	tracker := New(clientset, kubeNamespce, kubejob.Name)
	tracker.WaitForPhase(types.PhaseSucceeded)

	//Wait Delete Event
	var m sync.Mutex
	var cv = sync.NewCond(&m)
	m.Lock()

	var handler = func(job *batchv1.Job) (stop bool) {
		logger.Infof("Waiting for job delete Event")
		m.Lock()
		cv.Broadcast()
		m.Unlock()
		return stop
	}
	tracker.TrackDelete(handler)
	assert.NotNil(t, resp)
	assert.NoError(t, err)

	//Delete K8S Job
	err = clientset.BatchV1().Jobs("default").Delete(kubejob.Name, &metav1.DeleteOptions{})
	assert.NoError(t, err)
	cv.Wait()
	m.Unlock()
}
