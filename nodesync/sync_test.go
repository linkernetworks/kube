package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"github.com/stretchr/testify/assert"
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
	os.Setenv("NODE_RESOURCE_PERIODIC", "3")

	cf := config.MustRead(testingConfigPath)

	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	ms := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, ms)

	nts := New(clientset, ms)
	assert.NotNil(t, nts)
	signal := nts.Sync()

Watch:
	for {
		select {
		case <-signal:
			if (nts.stats.Added) != 0 {
				break Watch
			}
		}
	}

	nodes := nts.FetchNodes()
	if assert.NotEmpty(t, nodes) {
		assert.NotNil(t, nodes[0])
	}

	nts.Stop()
	updated := nts.stats.Added + nts.stats.Deleted + nts.stats.Updated
	assert.NotEqual(t, updated, 0, "should be added, deleted or updated events")
}

func TestNodeInCluster(t *testing.T) {
	if _, ok := os.LookupEnv("TEST_K8S"); !ok {
		t.Skip("Skip kubernetes related tests")
		return
	}

	nodes := []string{"gke-aurora-dev-default-pool-c6346db6-7x9b", "gke-aurora-dev-notebook-pool-a7f99c3f-mww8"}

	a := nodeInCluster("gke-aurora-dev-notebook-pool-a7f99c3f-mww8", nodes)
	assert.True(t, a)

	b := nodeInCluster("gke-aurora-dev-notebook-pool-a7f99c3f-999", nodes)
	assert.False(t, b)

}
