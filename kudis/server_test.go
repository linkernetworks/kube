package kudis

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/testutils"
	redis "bitbucket.org/linkernetworks/aurora/src/service/redis"

	"github.com/stretchr/testify/assert"
)

func TestCleanUp(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("require kubernetes")
		return
	}
	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobServer.DeploymentTargets, rds)

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
	dts := deployment.LoadDeploymentTargets(cf.JobServer.DeploymentTargets, rds)

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
	dts := deployment.LoadDeploymentTargets(cf.JobServer.DeploymentTargets, rds)

	server := NewServer(rds, dts)
	assert.NotNil(t, server)

	dt, err := server.GetDeploymentTarget("default")
	assert.NoError(t, err)

	kdt := dt.(*deployment.KubeDeploymentTarget)
	clientset := kdt.GetClientset()

	// testing purpose for creating a dummy job
	job := testutils.CreateKubernetesDummyJob("hello")
	_, err = testutils.DeployKubenetesJob(clientset, "default", job)
	assert.NoError(t, err)

	// should waiting the job state to succeed
	err = testutils.WaitUntilJobComplete(clientset, "default", job)
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
	defer testutils.DeleteKubenetesJob(clientset, "default", job)
}

func TestSubscribePodEvent(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("require kubernetes")
		return
	}

	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobServer.DeploymentTargets, rds)

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
