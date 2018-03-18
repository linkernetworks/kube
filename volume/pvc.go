package volumes

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NewPVC returns the kubernetes persistent volume claim object
func NewPVC(pvc entity.PersistentVolumeClaimParameter) (*v1.PersistentVolumeClaim, error) {
	resources := make(v1.ResourceList)

	capacity, err := resource.ParseQuantity(pvc.Capacity)
	if err != nil {
		return nil, err
	}
	resources[v1.ResourceName("storage")] = capacity

	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pvc.Name,
			Labels: map[string]string{
				"kind": "workspace",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Limits:   resources,
				Requests: resources,
			},
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.PersistentVolumeAccessMode(pvc.AccessMode)},
			StorageClassName: &pvc.StorageClass,
		},
	}, nil
}

func CreatePVC(clientset *kubernetes.Clientset, pvc entity.PersistentVolumeClaimParameter, namespace string) error {
	kubePVC, err := NewPVC(pvc)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().PersistentVolumeClaims(namespace).Create(kubePVC)
	if err != nil {
		return err
	}

	return nil
}

func GetPVC(clientset *kubernetes.Clientset, name string, namespace string) (*v1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
}

func DeletePVC(clientset *kubernetes.Clientset, name string, namespace string) error {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}
