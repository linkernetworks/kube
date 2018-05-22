package outcluster

import (
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/linkernetworks/logger"
)

func AllocateJobServerExternalService(clientset *kubernetes.Clientset, name string) error {
	logger.Infof("Checking %s service...", name)
	_, err := clientset.Core().Services("default").Get(name, metav1.GetOptions{})

	logger.Infof("Creating service: %s", name)
	if errors.IsNotFound(err) {
		s := NewJobServerExternalService(name)
		_, err = clientset.Core().Services("default").Create(s)
		return err
	}
	return err
}

func NewJobServerExternalService(name string) *v1.Service {
	return NewNodePortService(name, NodePortServiceParams{
		Labels: map[string]string{"app": "jobserver"},
		Selector: map[string]string{
			"app": "jobserver",
		},
		PortName:   "jobserver",
		TargetPort: 50051,
		NodePort:   30051,
	})
}
