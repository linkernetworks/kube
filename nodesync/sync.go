package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"errors"
	"gopkg.in/mgo.v2/bson"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type NodeTrackService struct {
	clientset *kubernetes.Clientset
	mongo     *mongo.MongoService
	context   *mongo.Context
	stop      chan struct{}
}

func New(clientset *kubernetes.Clientset, mongo *mongo.MongoService) *NodeTrackService {
	stop := make(chan struct{})
	return &NodeTrackService{
		clientset,
		mongo,
		mongo.NewContext(),
		stop,
	}
}

func (nts *NodeTrackService) Sync() {
	_, controller := kubemon.WatchNodes(nts.clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			err := syncDatabase(nts.context, n, "UPDATE")
			if err != nil {
				logger.Error(err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			err := syncDatabase(nts.context, n, "DELETE")
			if err != nil {
				logger.Error(err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n := newObj.(*corev1.Node)
			err := syncDatabase(nts.context, n, "UPDATE")
			if err != nil {
				logger.Error(err)
			}
		},
	})
	go controller.Run(nts.stop)
}

func (nts *NodeTrackService) Wait() {
	<-nts.stop
}

func (nts *NodeTrackService) Stop() {
	defer nts.context.Close()
	var e struct{}
	nts.stop <- e
}

func syncDatabase(context *mongo.Context, no *corev1.Node, action string) error {
	node := &entity.Node{
		Name:              no.GetName(),
		ClusterName:       no.GetClusterName(),
		CreationTimestamp: no.GetCreationTimestamp().Time,
		Labels:            no.GetLabels(),
		Allocatable: entity.Allocatable{
			CPU:       no.Status.Allocatable.Cpu().MilliValue(),
			Memory:    no.Status.Allocatable.Memory().MilliValue(),
			POD:       no.Status.Allocatable.Pods().Value(),
			NvidiaGPU: no.Status.Allocatable.NvidiaGPU().MilliValue(),
		},
		Capacity: entity.Capacity{
			CPU:       no.Status.Capacity.Cpu().MilliValue(),
			Memory:    no.Status.Capacity.Memory().MilliValue(),
			POD:       no.Status.Capacity.Pods().Value(),
			NvidiaGPU: no.Status.Capacity.NvidiaGPU().MilliValue(),
		},
	}
	for _, addr := range no.Status.Addresses {
		switch addr.Type {
		case "InternalIP":
			node.InternalIP = addr.Address
		case "ExternalIP":
			node.ExternalIP = addr.Address
		case "Hostname":
			node.Hostname = addr.Address
		}
	}
	update := bson.M{"$set": node}
	q := bson.M{"name": node.Name}

	switch action {
	case "UPDATE":
		_, err := context.C(entity.NodeCollectionName).Upsert(q, update)
		return err
	case "DELETE":
		err := context.C(entity.NodeCollectionName).Remove(q)
		return err
	default:
		return errors.New("Unknown action")
	}
}
