package podproxy

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
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

	redisService := redis.New(cf.Redis)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)

	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)

	updater := ProxyAddressUpdater{
		Clientset: clientset,
		Namespace: "default",
		PortName:  "mongo",
		Redis:     redisService,
	}
	_ = updater
}
