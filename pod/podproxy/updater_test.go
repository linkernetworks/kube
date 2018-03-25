package podproxy

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"

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

	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)

	updater := ProxyAddressUpdater{
		Clientset: clientset,
		Namespace: "default",
		PortName:  "mongo",
	}
	_ = updater
}
