package outcluster

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/logger"
)

func DeleteNodePortServices(clientset *kubernetes.Clientset) error {
	logger.Info("Deleting mongo-external service")
	clientset.Core().Services("default").Delete("mongo-external", nil)

	logger.Info("Deleting redis-external service")
	clientset.Core().Services("default").Delete("redis-external", nil)
	return nil
}

func AllocateNodePortServices(clientset *kubernetes.Clientset, cf config.Config) error {
	if err := AllocateMongoExternalService(clientset, "mongo-external"); err != nil {
		return err
	}
	if err := AllocateRedisExternalService(clientset, "redis-external"); err != nil {
		return err
	}
	return nil
}

func AllocateRedisExternalService(clientset *kubernetes.Clientset, name string) error {
	logger.Infof("Checking %s service...", name)
	s, err := clientset.Core().Services("default").Get(name, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	logger.Infof("Creating service: %s", name)
	s = NewRedisExternalService(name)
	_, err = clientset.Core().Services("default").Create(s)
	return err
}

func AllocateMongoExternalService(clientset *kubernetes.Clientset, name string) error {
	logger.Infof("Labeling podindex on mongo-0 pod for %s", name)
	_, err := clientset.Core().Pods("default").Patch("mongo-0", types.JSONPatchType, []byte(`[ { "op": "add", "path": "/metadata/labels/podindex", "value": "0" } ]`))
	if err != nil {
		return err
	}

	logger.Infof("Checking %s service...", name)
	s, err := clientset.Core().Services("default").Get(name, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	logger.Infof("Creating service: %s", name)
	s = NewMongoExternalService(name)
	_, err = clientset.Core().Services("default").Create(s)
	return err
}

func Connect(clientset *kubernetes.Clientset, cf config.Config) (config.Config, error) {
	var dst = cf
	var err error

	node, address, err := DiscoverVisibleNode(clientset)
	if err != nil {
		return dst, err
	}
	_ = node

	if err := AllocateNodePortServices(clientset, cf); err != nil {
		return dst, err
	}

	mongo, err := clientset.Core().Services("default").Get("mongo-external", metav1.GetOptions{})
	if err != nil {
		return dst, err
	}
	for _, port := range mongo.Spec.Ports {
		if port.Name == "mongo" {
			dst.Mongo.Url = fmt.Sprintf("mongodb://%s:%d/aurora", address, port.NodePort)
			logger.Infof("Rewrited mongodb url to %s", dst.Mongo.Url)
		}
	}

	redis, err := clientset.Core().Services("default").Get("redis-external", metav1.GetOptions{})
	if err != nil {
		return dst, err
	}
	for _, port := range redis.Spec.Ports {
		if port.Name == "redis" {
			dst.Redis.Host = address
			dst.Redis.Port = port.NodePort
			logger.Infof("Rewrited redis address to %s", dst.Redis.Addr())
		}
	}

	return dst, nil
}

func NewRedisExternalService(name string) *v1.Service {
	return NewNodePortService(name, NodePortServiceParams{
		Labels: map[string]string{"environment": "testing"},
		Selector: map[string]string{
			// TODO: use role for consistency
			"name": "redis",
		},
		PortName:   "redis",
		TargetPort: 6379,
		NodePort:   32199,
	})

}

func NewMongoExternalService(name string) *v1.Service {
	return NewNodePortService(name, NodePortServiceParams{
		Labels: map[string]string{"environment": "testing"},
		Selector: map[string]string{
			"role":     "mongo",
			"podindex": "0",
		},
		PortName:   "mongo",
		TargetPort: 27017,
		NodePort:   31717,
	})
}

type NodePortServiceParams struct {
	PortName string

	Labels   map[string]string
	Selector map[string]string

	TargetPort int32
	NodePort   int32
}

func NewNodePortService(name string, params NodePortServiceParams) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: params.Labels,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: params.Selector,
			Ports: []v1.ServicePort{
				{
					Name:       params.PortName,
					Port:       params.TargetPort,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: params.TargetPort},
					NodePort:   params.NodePort,
				},
			},
		},
	}
}
