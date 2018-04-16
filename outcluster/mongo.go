package outcluster

import (
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"bitbucket.org/linkernetworks/aurora/src/logger"
)

func AllocateMongoExternalService(clientset *kubernetes.Clientset, name string) error {
	logger.Infof("Labeling podindex on mongo-0 pod for %s", name)
	if _, err := clientset.Core().Pods("default").Patch("mongo-0", types.JSONPatchType, []byte(`[ { "op": "add", "path": "/metadata/labels/podindex", "value": "0" } ]`)); err != nil {
		return err
	}

	logger.Infof("Checking %s service...", name)
	_, err := clientset.Core().Services("default").Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.Infof("Creating service: %s", name)
		s := NewMongoExternalService(name)
		_, err = clientset.Core().Services("default").Create(s)
		return err
	}
	return err
}
