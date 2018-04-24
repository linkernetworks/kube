package outcluster

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	k8ssvc "bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
)

type RewriteSettings struct {
	RewriteJobServer  bool
	RewriteJobUpdater bool
	RewriteKudis      bool
}

func MustParseLocalRewrite(str string) (settings RewriteSettings) {
	settings.RewriteJobServer = true
	settings.RewriteJobUpdater = true
	settings.RewriteKudis = true
	for _, key := range strings.Split(str, ",") {
		switch key {
		case "jobserver":
			settings.RewriteJobServer = false
		case "jobupdater":
			settings.RewriteJobUpdater = false
		case "kudis":
			settings.RewriteKudis = false
		case "":
		default:
			panic(fmt.Errorf("key %s is not supported.", key))
		}
	}
	return settings
}

func DeleteNodePortServices(clientset *kubernetes.Clientset) error {
	logger.Info("Deleting mongo-external service")
	clientset.Core().Services("default").Delete("mongo-external", nil)

	logger.Info("Deleting redis-external service")
	clientset.Core().Services("default").Delete("redis-external", nil)

	logger.Info("Deleting redis-external service")
	clientset.Core().Services("default").Delete("influxdb-external", nil)

	logger.Info("Deleting jobserver service")
	clientset.Core().Services("default").Delete("jobserver-external", nil)
	return nil
}

func AllocateNodePortServices(clientset *kubernetes.Clientset, cf config.Config) error {
	if err := AllocateMongoExternalService(clientset, "mongo-external"); err != nil {
		return err
	}
	if err := AllocateRedisExternalService(clientset, "redis-external"); err != nil {
		return err
	}
	if err := AllocateInfluxdbExternalService(clientset, "influxdb-external"); err != nil {
		return err
	}
	if err := AllocateJobServerExternalService(clientset, "jobserver-external"); err != nil {
		return err
	}
	if err := AllocateKudisExternalService(clientset, "kudis-external"); err != nil {
		return err
	}
	return nil
}

// ConnectWith creates the external services and rewrite the config
func ConnectWith(clientset *kubernetes.Clientset, cf config.Config, settings RewriteSettings) (config.Config, error) {
	var dst = cf

	if err := AllocateNodePortServices(clientset, cf); err != nil {
		return dst, err
	}

	return Rewrite(clientset, dst, settings)
}

func ConnectAndRewrite(cf config.Config, settings RewriteSettings) (config.Config, error) {
	if cf.Kubernetes == nil {
		return cf, fmt.Errorf("kubernetes config is not defined, can't convert config to load kubernetes service")
	}

	svc := k8ssvc.NewFromConfig(cf.Kubernetes)
	clientset, err := svc.NewClientset()
	if err != nil {
		return cf, err
	}
	return ConnectWith(clientset, cf, settings)
}

func Rewrite(clientset *kubernetes.Clientset, cf config.Config, settings RewriteSettings) (config.Config, error) {
	var dst = cf
	var err error

	/*
		if cf.Kubernetes.OutCluster.AddressType != "" {
			node, address, err := DiscoverVisibleNodeByAddressType(clientset, cf.Kubernetes.OutCluster.AddressType)
		}
	*/

	node, address, err := DiscoverVisibleNode(clientset)
	if err != nil {
		return dst, err
	}

	if node == nil {
		return dst, fmt.Errorf("node not found")
	}

	logger.Infof("Found node address: %v", address)

	mongo, err := clientset.Core().Services("default").Get("mongo-external", metav1.GetOptions{})
	if err != nil {
		return dst, err
	}
	for _, port := range mongo.Spec.Ports {
		dst.Mongo.Url = fmt.Sprintf("mongodb://%s:%d/aurora", address, port.NodePort)
		logger.Infof("Rewrited mongodb url to %s", dst.Mongo.Url)
		break
	}

	redis, err := clientset.Core().Services("default").Get("redis-external", metav1.GetOptions{})
	if err != nil {
		return dst, err
	}
	for _, port := range redis.Spec.Ports {
		dst.Redis.Host = address
		dst.Redis.Port = port.NodePort
		logger.Infof("Rewrited redis address to %s", dst.Redis.Addr())
		break
	}

	influxdb, err := clientset.Core().Services("default").Get("influxdb-external", metav1.GetOptions{})
	if err != nil {
		return dst, err
	}
	for _, port := range influxdb.Spec.Ports {
		dst.Influxdb.Url = "http://" + address + ":" + strconv.Itoa(int(port.NodePort))
		logger.Infof("Rewrited influxdb address to %s", dst.Influxdb.Url)
		break
	}

	if settings.RewriteJobServer {
		jobserver, err := clientset.Core().Services("default").Get("jobserver-external", metav1.GetOptions{})
		if err != nil {
			return dst, err
		}
		for _, port := range jobserver.Spec.Ports {
			dst.JobServer.Host = address
			dst.JobServer.Port = port.NodePort
			logger.Infof("Rewrited jobserver address to %s", dst.JobServer.Addr())
			break
		}
	}

	if settings.RewriteKudis {
		svc, err := clientset.Core().Services("default").Get("kudis-external", metav1.GetOptions{})
		if err != nil {
			return dst, err
		}
		for _, port := range svc.Spec.Ports {
			dst.Kudis.Host = address
			dst.Kudis.Port = port.NodePort
			logger.Infof("Rewrited kudis address to %s", dst.Kudis.Addr())
			break
		}
	}

	return dst, nil
}

func NewInfluxdbExternalService(name string) *v1.Service {
	return NewNodePortService(name, NodePortServiceParams{
		Labels: map[string]string{"environment": "testing"},
		Selector: map[string]string{
			"service": "influxdb",
		},
		PortName:   "influxdb",
		TargetPort: 8086,
		NodePort:   32086,
	})
}

func NewMongoExternalService(name string) *v1.Service {
	return NewNodePortService(name, NodePortServiceParams{
		Labels: map[string]string{"environment": "testing"},
		Selector: map[string]string{
			"service":  "mongo",
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
