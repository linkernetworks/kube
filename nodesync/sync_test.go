package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubeconfig"
	// "bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"os"
	"testing"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

func TestNodeSync(t *testing.T) {
	if _, ok := os.LookupEnv("TEST_K8S"); !ok {
		t.Skip("Skip kubernetes related tests")
		return
	}

	cf := config.Read(testingConfigPath)

	config, err := kubeconfig.Load("", "")
	assert.NoError(t, err)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	assert.NoError(t, err)

	var nodeResults []*entity.Node
	ms := mongo.NewMongoService(cf.Mongo.Url)
	nts := New(clientset, ms)
	nts.Sync()

	// this context is for finding any data in node collection
	context := ms.NewContext()
Watch:
	for {
		err = context.C(entity.NodeCollectionName).Find(nil).All(&nodeResults)
		assert.NoError(t, err)
		if len(nodeResults) != 0 {
			break Watch
		}
	}
	nts.Stop()
	assert.NotEqual(t, len(nodeResults), 0, "mongodb node collection is empty")
}
