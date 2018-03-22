package kudis

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	redis "bitbucket.org/linkernetworks/aurora/src/service/redis"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/stretchr/testify/assert"
)

func TestCleanUp(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("require kubernetes")
		return
	}
	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := NewServer(rds, dts)
	assert.NotNil(t, server)

	dt, err := server.GetDeploymentTarget("default")
	assert.NoError(t, err)

	var subscription Subscription = NewPodLogSubscription(rds, "default", dt, "mongo-0", "mongo-sidecar", 10)
	assert.NotNil(t, subscription)

	success, reason, err := server.Subscribe(subscription)
	assert.NoError(t, err)
	assert.True(t, success)
	t.Logf("reason: %s", reason)

	for i := 0; i < 4; i++ {
		var err = server.CleanUp()
		assert.NoError(t, err)
	}
}

func TestSubscribePodLogs(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("require kubernetes")
		return
	}

	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := NewServer(rds, dts)
	assert.NotNil(t, server)

	dt, err := server.GetDeploymentTarget("default")
	assert.NoError(t, err)

	var subscription Subscription = NewPodLogSubscription(rds, "default", dt, "mongo-0", "mongo", 10)
	assert.NotNil(t, subscription)

	success, reason, err := server.Subscribe(subscription)
	assert.NoError(t, err)
	assert.True(t, success)
	t.Logf("reason: %s", reason)

	err = subscription.Stop()
	assert.NoError(t, err)
}

func TestSubscribeJobLogs(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("require kubernetes")
		return
	}

	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := NewServer(rds, dts)
	assert.NotNil(t, server)

	dt, err := server.GetDeploymentTarget("default")
	assert.NoError(t, err)

	kdt := dt.(*deployment.KubeDeploymentTarget)
	clientset := kdt.GetClientset()

	// testing purpose for creating a dummy job
	job := createKubernetesDummyJob("hello")
	_, err = deployKubenetesJob(clientset, "default", job)
	assert.NoError(t, err)

	// should waiting the job state to succeed
	err = waitUntilContainerSucceed(clientset, "default", job)
	assert.NoError(t, err)

	var subscription Subscription = NewJobLogSubscription(rds, "default", dt, "hello", "hello", 10)
	assert.NotNil(t, subscription)

	success, reason, err := server.Subscribe(subscription)
	assert.NoError(t, err)
	assert.True(t, success)
	t.Logf("reason: %s", reason)

	err = subscription.Stop()
	assert.NoError(t, err)

	// cleanup
	err = deleteKubenetesJob(clientset, "default", job)
	assert.NoError(t, err)
}

func TestSubscribePodEvent(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("require kubernetes")
		return
	}

	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := NewServer(rds, dts)
	assert.NotNil(t, server)

	dt, err := server.GetDeploymentTarget("default")
	assert.NoError(t, err)

	var subscription Subscription = NewPodEventSubscription(rds, "default", dt, "mongo-0")
	assert.NotNil(t, subscription)

	success, reason, err := server.Subscribe(subscription)
	assert.NoError(t, err)
	assert.True(t, success)
	t.Logf("reason: %s", reason)

	err = subscription.Stop()
	assert.NoError(t, err)
}

func createKubernetesDummyJob(name string) *batchV1.Job {
	return &batchV1.Job{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: batchV1.JobSpec{
			Template: coreV1.PodTemplateSpec{
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Name:  name,
							Image: "hello-world:latest",
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
}

func deployKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) (*batchV1.Job, error) {
	return clientset.BatchV1().Jobs(namespace).Create(job)
}

func deleteKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) error {
	opts := metaV1.NewDeleteOptions(0)
	return clientset.BatchV1().Jobs(namespace).Delete(job.GetName(), opts)
}

func waitUntilContainerSucceed(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) error {
	for {
		j, err := clientset.BatchV1().Jobs(namespace).Get(job.GetName(), metaV1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Succeeded == 1 || j.Status.Failed == 1 {
			return nil
		}
	}
}
