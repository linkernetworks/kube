package volume

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ParseQuantity(quantity string) (resource.Quantity, error) {
	return resource.ParseQuantity(quantity)
}

// NewPVC returns the kubernetes persistent volume claim object
func NewPVC(params entity.PersistentVolumeClaimParameter) (*v1.PersistentVolumeClaim, error) {
	resources := make(v1.ResourceList)

	capacity, err := ParseQuantity(params.Capacity)
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
			Name: params.Name,
			Labels: map[string]string{
				"kind": "workspace",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Limits:   resources,
				Requests: resources,
			},
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.PersistentVolumeAccessMode(params.AccessMode)},
			StorageClassName: &params.StorageClass,
		},
	}, nil
}

func CreatePVC(clientset *kubernetes.Clientset, namespace string, pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
}

func GetPVC(clientset *kubernetes.Clientset, namespace string, name string) (*v1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
}

func DeletePVC(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}
