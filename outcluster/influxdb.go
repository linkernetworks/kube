package outcluster

import (
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/linkernetworks/logger"
)

func AllocateInfluxdbExternalService(clientset *kubernetes.Clientset, name string) error {
	logger.Infof("Checking %s service...", name)
	_, err := clientset.Core().Services("default").Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.Infof("Creating service: %s", name)
		s := NewInfluxdbExternalService(name)
		_, err = clientset.Core().Services("default").Create(s)
		return err
	}
	return err
}
