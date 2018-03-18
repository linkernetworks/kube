package volume

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	k8ssvc "bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	testingConfigPath = "../../../config/testing.json"
	pvcName           = "testing"
	namespace         = "default"
)

func TestCreatePersistentVolumeClaim(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}
	cf := config.MustRead(testingConfigPath)
	ksvc := k8ssvc.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	pvc := entity.PersistentVolumeClaimParameter{
		Name:         pvcName,
		StorageClass: "slow-one",
		Capacity:     "1Gi",
		AccessMode:   "ReadWriteOnce",
	}

	err = CreatePVC(clientset, pvc, namespace)
	assert.NoError(t, err)
}

func TestGetPersistentVolumeClaim(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}
	cf := config.MustRead(testingConfigPath)
	ksvc := k8ssvc.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	pvc, err := GetPVC(clientset, pvcName, namespace)
	assert.NoError(t, err)
	assert.NotNil(t, pvc)
	assert.Equal(t, pvcName, pvc.Name)
}

func TestDeletePersistentVolumeClaim(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}
	cf := config.MustRead(testingConfigPath)
	ksvc := k8ssvc.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.CreateClientset()
	assert.NoError(t, err)

	err = DeletePVC(clientset, pvcName, namespace)
	assert.NoError(t, err)
}
