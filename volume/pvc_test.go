package volume

import (
	"github.com/linkernetworks/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	k8ssvc "github.com/linkernetworks/service/kubernetes"
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
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	params := entity.PersistentVolumeClaimParameter{
		Name:         pvcName,
		StorageClass: "slow-one",
		Capacity:     "1Gi",
		AccessMode:   "ReadWriteOnce",
	}

	pvc, err := NewPVC(params)
	assert.NoError(t, err)

	created, err := CreatePVC(clientset, namespace, pvc)
	assert.NoError(t, err)
	assert.NotNil(t, created)
}

func TestGetPersistentVolumeClaim(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}
	cf := config.MustRead(testingConfigPath)
	ksvc := k8ssvc.NewFromConfig(cf.Kubernetes)
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	pvc, err := GetPVC(clientset, namespace, pvcName)
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
	clientset, err := ksvc.NewClientset()
	assert.NoError(t, err)

	err = DeletePVC(clientset, namespace, pvcName)
	assert.NoError(t, err)
}
