package podproxy

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"github.com/stretchr/testify/assert"
)

const (
	testingConfigPath = "../../../../config/testing.json"
)

func TestUpdater(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)

	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	mongoService := mongo.New(cf.Mongo.Url)
	redisService := redis.New(cf.Redis)

	clientset, err := kubernetesService.CreateClientset()
	assert.NoError(t, err)

	updater := DocumentProxyInfoUpdater{
		Clientset:      clientset,
		Namespace:      "default",
		Redis:          redisService,
		Mongo:          mongoService,
		CollectionName: "testobjs",
		PortName:       "mongo",
	}
	_ = updater
}
