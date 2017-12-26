package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/outcluster"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"fmt"
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

	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	newcf, err := outcluster.Connect(clientset, cf)
	assert.NoError(t, err)

	var nodeResults []*entity.Node
	ms := mongo.NewMongoService(newcf.Mongo.Url)
	assert.NotNil(t, ms)
	fmt.Println(cf.Mongo.Url)

	nts := New(clientset, ms)
	assert.NotNil(t, nts)
	signal := nts.Sync()

	// this context is for finding any data in node collection
	context := ms.NewContext()

	fmt.Println("checking result")
Watch:
	for {
		select {
		case <-signal:
			err = context.C(entity.NodeCollectionName).Find(nil).All(&nodeResults)
			assert.NoError(t, err)
			if len(nodeResults) != 0 {
				break Watch
			}
		}
	}

	fmt.Println("stopping")
	nts.Stop()
	assert.NotEqual(t, len(nodeResults), 0, "mongodb node collection is empty")
}
