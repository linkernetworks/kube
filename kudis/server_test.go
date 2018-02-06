package kudis

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	redis "bitbucket.org/linkernetworks/aurora/src/service/redis"

	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	config config.Config
}

func (suite *ServerTestSuite) TestCleanUp() {
	cf := suite.config
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := New(rds, dts)
	suite.NotNil(server)

	dt, err := server.GetDeploymentTarget("default")
	suite.NoError(err)

	var subscription Subscription = NewPodLogsSubscription(rds, "default", dt, "mongo-0", "mongo-sidecar", 10)
	suite.NotNil(subscription)

	success, reason, err := server.Subscribe(subscription)
	suite.NoError(err)
	suite.True(success)
	suite.T().Logf("reason: %s", reason)

	for i := 0; i < 4; i++ {
		var err = server.CleanUp()
		suite.NoError(err)
	}
}

func (suite *ServerTestSuite) TestSubscribe() {
	cf := suite.config
	rds := redis.New(cf.Redis)
	dts := deployment.LoadDeploymentTargets(cf.JobController.DeploymentTargets, rds)

	server := New(rds, dts)
	suite.NotNil(server)

	dt, err := server.GetDeploymentTarget("default")
	suite.NoError(err)

	var subscription Subscription = NewPodLogsSubscription(rds, "default", dt, "mongo-0", "mongo", 10)
	suite.NotNil(subscription)

	success, reason, err := server.Subscribe(subscription)
	suite.NoError(err)
	suite.True(success)
	suite.T().Logf("reason: %s", reason)

	err = subscription.Stop()
	suite.NoError(err)
}

func TestServerTestSuite(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	suite.Run(t, &ServerTestSuite{config: cf})
}
