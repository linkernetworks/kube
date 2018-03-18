package outcluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
)

func TestDiscoverVisibleNodes(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip()
		return
	}

	cf := config.MustRead("../../../config/testing.json")
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	node, addr, err := DiscoverVisibleNode(clientset)
	assert.NoError(t, err)

	assert.True(t, addr != "")
	assert.NotNil(t, node)
}
