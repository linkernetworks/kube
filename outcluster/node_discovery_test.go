package outcluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/linkernetworks/config"
	"github.com/linkernetworks/service/kubernetes"
)

func TestDiscoverVisibleNodes(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip()
		return
	}

	cf := config.MustRead("../../../config/testing.json")

	t.Logf("config: %+v", cf.Kubernetes)

	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)

	restConfig, err := ksvc.LoadConfig()
	assert.NoError(t, err)
	t.Logf("rest config: %+v", restConfig)

	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	// assert.Equal(t, "ExternalIP", cf.Kubernetes.OutCluster.AddressType)

	node, addr, err := DiscoverVisibleNode(clientset)
	assert.NoError(t, err)
	assert.NotNil(t, node)

	assert.True(t, addr != "")
	assert.NotNil(t, node)
}
