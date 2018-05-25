package jobtracker

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/linkernetworks/config"
	"bitbucket.org/linkernetworks/aurora/src/jobcontroller/types"
	"github.com/linkernetworks/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"k8s.io/apimachinery/pkg/api/resource"

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
							Image:   "alpine:3.7",
							Command: []string{"/bin/sh", "-c", fmt.Sprintf("sleep %d", seconds)},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{"cpu": resource.MustParse("20m")},
								Requests: v1.ResourceList{
									"cpu": resource.MustParse("10m"),
								},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

}

func TestJobTracker(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}
	//Create a K8S Job
	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)
	assert.NotNil(t, clientset)

	id := bson.NewObjectId().Hex()
	kubejob := CreateKubernetesSleepJob(id, 2)

	resp, err := clientset.BatchV1().Jobs(kubeNamespce).Create(kubejob)
	//Use this defer to make sure the job is been deleted after the test
	defer clientset.BatchV1().Jobs("default").Delete(kubejob.Name, &metav1.DeleteOptions{})
	assert.NotNil(t, resp)
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
	assert.NoError(t, err)

	//Delete K8S Job
	err = clientset.BatchV1().Jobs("default").Delete(kubejob.Name, &metav1.DeleteOptions{})
	assert.NoError(t, err)
	cv.Wait()
	m.Unlock()
}
