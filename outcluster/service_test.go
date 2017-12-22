package outcluster

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
)

func TestAllocateNodePortServices(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.Skip()
		return
	}

	cf := config.Read("../../../config/testing.json")
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	DeleteNodePortServices(clientset)

	err = AllocateNodePortServices(clientset, cf)
	assert.NoError(t, err)
}
