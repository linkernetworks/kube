package outcluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"github.com/linkernetworks/mongo"
	"github.com/linkernetworks/redis"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestConnectWith(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("TEST_K8S is not set.")
	}

	cf := config.MustRead("../../../config/testing.json")
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	// DeleteNodePortServices(clientset)

	newcf, err := ConnectWith(clientset, cf, RewriteSettings{})
	assert.NoError(t, err)

	m := mongo.New(newcf.Mongo.Url)
	assert.NotNil(t, m)

	r := redis.New(newcf.Redis)
	assert.NotNil(t, r)
}

func TestAllocateNodePortServices(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip("TEST_K8S is not set.")
	}

	cf := config.MustRead("../../../config/testing.json")
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	// DeleteNodePortServices(clientset)

	err = AllocateNodePortServices(clientset, cf)
	assert.NoError(t, err)
}
