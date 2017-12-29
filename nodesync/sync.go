package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"gopkg.in/mgo.v2/bson"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type NodeStats struct {
	Added   int
	Updated int
	Deleted int
}

type NodeSync struct {
	clientset *kubernetes.Clientset
	context   *mongo.Context
	stop      chan struct{}
	stats     NodeStats
}

func New(clientset *kubernetes.Clientset, m *mongo.MongoService) *NodeSync {
	stop := make(chan struct{})
	var stats NodeStats
	return &NodeSync{clientset, m.NewContext(), stop, stats}
}

type Signal chan bool

func (nts *NodeSync) Sync() Signal {
	signal := make(Signal, 1)

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				nodesName := nts.FetchNewNodesName()
				nts.Prune(nodesName)
			}
		}

	}()

	_, controller := kubemon.WatchNodes(nts.clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Added++

			node := CreateNodeEntity(n)
			logger.Info("Nodes state added")
			err := nts.UpsertNode(&node)
			if err != nil {
				logger.Error(err)
			}

			select {
			case signal <- true:
			}
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Updated++

			node := CreateNodeEntity(n)
			logger.Info("Nodes state deleted")
			err := nts.RemoveNodeByName(node.Name)
			if err != nil {
				logger.Error(err)
			}

			select {
			case signal <- true:
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n := newObj.(*corev1.Node)
			nts.stats.Deleted++

			node := CreateNodeEntity(n)
			logger.Info("Nodes state updated")
			err := nts.UpsertNode(&node)
			if err != nil {
				logger.Error(err)
			}

			select {
			case signal <- true:
			}
		},
	})
	go controller.Run(nts.stop)

	return signal
}

func (nts *NodeSync) Wait() {
	<-nts.stop
}

func (nts *NodeSync) Stop() {
	defer nts.context.Close()
	var e struct{}
	nts.stop <- e
}

func CreateNodeEntity(no *corev1.Node) entity.Node {
	node := entity.Node{
		Name:              no.GetName(),
		ClusterName:       no.GetClusterName(),
		CreationTimestamp: no.GetCreationTimestamp().Time,
		Labels:            createLabelSlice(no.GetLabels()),
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
	return node
}

func createLabelSlice(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		l := k + "=" + v
		s = append(s, l)
	}
	return s
}

func (nts *NodeSync) UpsertNode(node *entity.Node) error {
	update := bson.M{"$set": node}
	q := bson.M{"name": node.Name}
	_, err := nts.context.C(entity.NodeCollectionName).Upsert(q, update)
	return err
}

func (nts *NodeSync) RemoveNodeByName(name string) error {
	q := bson.M{"name": name}
	return nts.context.C(entity.NodeCollectionName).Remove(q)
}

func (nts *NodeSync) FetchNewNodesName() []string {
	var nodesName []string
	nodeList, _ := kubemon.GetNodes(nts.clientset)
	for _, no := range nodeList.Items {
		nodesName = append(nodesName, no.Name)
	}
	logger.Infof("Current Nodes: %v", nodesName)
	return nodesName
}

func (nts *NodeSync) Prune(newNodesName []string) error {
	results := []entity.Node{}
	err := nts.context.C(entity.NodeCollectionName).Find(nil).All(&results)
	if err != nil {
		return err
	}
	for _, r := range results {
		if !nodeInCluster(r.Name, newNodesName) {
			logger.Info("Pruning database...")
			err := nts.context.C(entity.NodeCollectionName).Remove(bson.M{"name": r.Name})
			if err != nil {
				logger.Error("Pruning database error")
				return err
			}
		}
	}
	return nil
}

func nodeInCluster(n string, list []string) bool {
	for _, l := range list {
		if l == n {
			return true
		}
	}
	return false
}
