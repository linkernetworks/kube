package outcluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestConnectWith(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("TEST_K8S is not set.")
	}

	cf := config.Read("../../../config/testing.json")
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	DeleteNodePortServices(clientset)

	newcf, err := ConnectWith(clientset, cf)
	assert.NoError(t, err)

	m := mongo.NewMongoService(newcf.Mongo.Url)
	assert.NotNil(t, m)

	r := redis.NewService(newcf.Redis)
	assert.NotNil(t, r)
}

func TestAllocateNodePortServices(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("TEST_K8S is not set.")
	}

	cf := config.Read("../../../config/testing.json")
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	DeleteNodePortServices(clientset)

	err = AllocateNodePortServices(clientset, cf)
	assert.NoError(t, err)
}
