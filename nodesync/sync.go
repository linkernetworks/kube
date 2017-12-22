package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"errors"
	"gopkg.in/mgo.v2/bson"
	core_v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func nodeSync(context *mongo.Context, n *core_v1.Node, action string) error {
	node := &entity.Node{
		// ID:                bson.NewObjectId(),
		Name:              n.GetName(),
		ClusterName:       n.GetClusterName(),
		CreationTimestamp: n.GetCreationTimestamp().Time,
		Labels:            n.GetLabels(),
		Allocatable: entity.Allocatable{
			CPU:       n.Status.Allocatable.Cpu().String(),
			Memory:    n.Status.Allocatable.Memory().String(),
			POD:       n.Status.Allocatable.Pods().String(),
			NvidiaGPU: n.Status.Allocatable.NvidiaGPU().String(),
		},
		Capacity: entity.Capacity{
			CPU:       n.Status.Capacity.Cpu().String(),
			Memory:    n.Status.Capacity.Memory().String(),
			POD:       n.Status.Capacity.Pods().String(),
			NvidiaGPU: n.Status.Capacity.NvidiaGPU().String(),
		},
	}
	for _, addr := range n.Status.Addresses {
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
	if action == "UPDATE" {
		_, err := context.C(entity.NodeCollectionName).Upsert(q, update)
		return err
	} else if action == "DELETE" {
		err := context.C(entity.NodeCollectionName).Remove(q)
		return err
	} else {
		return errors.New("Unknow action")
	}
	return nil

}

func track(clientset *kubernetes.Clientset, ms *mongo.MongoService) {
	context := ms.NewContext()
	defer context.Close()
	var e struct{}
	stop := make(chan struct{})
	_, controller := kubemon.WatchNodes(clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n := obj.(*core_v1.Node)
			logger.Debug("============= ADD =============")
			err := nodeSync(context, n, "UPDATE")
			if err != nil {
				logger.Error(err)
				stop <- e
			}
			logger.Debug("============= END ADD =============")
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*core_v1.Node)
			logger.Debug("============= DELETE =============")
			err := nodeSync(context, n, "DELETE")
			if err != nil {
				logger.Error(err)
				stop <- e
			}
			logger.Debug("============= END DELETE =============")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n := newObj.(*core_v1.Node)
			err := nodeSync(context, n, "UPDATE")
			if err != nil {
				logger.Error(err)
				stop <- e
			}
			logger.Debug("============= END UPDATE =============")
		},
	})
	go controller.Run(stop)
	<-stop
}
