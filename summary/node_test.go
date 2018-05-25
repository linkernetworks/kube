package summary

import (
	"os"
	"testing"

	"github.com/linkernetworks/config"
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"github.com/stretchr/testify/assert"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

func TestQueryNodeGPUUsage(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	// Get mongo service
	cf := config.MustRead(testingConfigPath)

	// kubernetes service (config)
	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)

	// kubernetes clientset (get from kubernetes svc)
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	// create deployment target (pass kubernetes clientset)
	dt := deployment.KubeDeploymentTarget{
		Clientset: clientset,
		Namespace: "testing",
		Redis:     nil,
	}

	nodes, err := dt.GetNodes()
	assert.NoError(t, err)
	assert.True(t, len(nodes) > 0)

	usage := QueryNodeGPUUsage(&dt, nodes[0].Name)
	assert.NotNil(t, usage)
}
