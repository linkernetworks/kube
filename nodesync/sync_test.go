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

	cf := config.Read(testingConfigPath)

	ksvc := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	ms := mongo.NewMongoService(cf.Mongo.Url)
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

	nts.Stop()
	updated := nts.stats.Added + nts.stats.Deleted + nts.stats.Updated
	assert.NotEqual(t, updated, 0, "should be added, deleted or updated events")
}
