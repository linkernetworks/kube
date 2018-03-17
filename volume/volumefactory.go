package volumes

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VolumeFactory struct {
}

func (n VolumeFactory) NewPV(name string, capacity string, accessMode string, provider entity.VolumeProvider) (*v1.PersistentVolume, error) {
	reList := make(v1.ResourceList)

	storage, err := resource.ParseQuantity(capacity)
	if err != nil {
		return nil, err
	}
	reList[v1.ResourceName("storage")] = storage

	return &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity:               reList,
			PersistentVolumeSource: provider.PersistentVolumeSource(),

			AccessModes: []v1.PersistentVolumeAccessMode{v1.PersistentVolumeAccessMode(accessMode)},
		},
	}, nil
}
