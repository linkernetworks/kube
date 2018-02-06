package kudis

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	redis "bitbucket.org/linkernetworks/aurora/src/service/redis"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := New(rds, dts)
	assert.NotNil(t, server)

	dt, err := server.GetDeploymentTarget("default")
	assert.NoError(t, err)

	var subscription Subscription = NewPodLogsSubscription(rds, "default", dt, "mongo-0", "mongo", 10)
	assert.NotNil(t, subscription)

	success, reason, err := server.Subscribe(subscription)
	assert.NoError(t, err)
	assert.True(t, success)
	t.Logf("reason: %s", reason)

	err = subscription.Stop()
	assert.NoError(t, err)
}
