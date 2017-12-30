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
	// fetch all nodes from cluster and save to mongodb
	logger.Info("fetching all nodes")
	nodes := nts.FetchNodes()
	for _, n := range nodes {
		nodeEntity := LoadNodeEntity(n)
		err := nts.UpsertNode(&nodeEntity)
		if err != nil {
			logger.Error("Upsert node error:", err)
		}
	}

	// cleanup old node record
	logger.Info("cleaning old record")
	nts.Prune()

	logger.Info("start watching node change events...")
	// keep watch node change events
	_, controller := kubemon.WatchNodes(nts.clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Added++

			node := LoadNodeEntity(n)
			logger.Info("[Event] nodes state added")
			err := nts.UpsertNode(&node)
			if err != nil {
				logger.Error("Upsert node error:", err)
			}

			select {
			case signal <- true:
			}
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Updated++

			node := LoadNodeEntity(n)
			logger.Info("[Event] nodes state deleted")
			err := nts.RemoveNodeByName(node.Name)
			if err != nil {
				logger.Error("Remove node error:", err)
			}

			select {
			case signal <- true:
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n := newObj.(*corev1.Node)
			nts.stats.Deleted++

			node := LoadNodeEntity(n)
			logger.Info("[Event] nodes state updated")
			err := nts.UpsertNode(&node)
			if err != nil {
				logger.Error("Upsert node error:", err)
			}

			select {
			case signal <- true:
			}
		},
	})
	go controller.Run(nts.stop)
	go nts.StartPrune()
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

func LoadNodeEntity(no *corev1.Node) entity.Node {
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

func (nts *NodeSync) FetchNodes() []*corev1.Node {
	var nodes []*corev1.Node
	nodeList, _ := kubemon.GetNodes(nts.clientset)
	for _, no := range nodeList.Items {
		nodes = append(nodes, &no)
	}
	return nodes
}

func (nts *NodeSync) Prune() error {
	var newNodeNames []string
	nodesList := nts.FetchNodes()
	for _, n := range nodesList {
		newNodeNames = append(newNodeNames, n.Name)
	}
	nodes := []entity.Node{}
	err := nts.context.C(entity.NodeCollectionName).Find(nil).Select(bson.M{"name": 1}).All(&nodes)
	if err != nil {
		return err
	}
	// check mongodb record if node doesn't exist in current cluster than remove the record
	for _, n := range nodes {
		if !nodeInCluster(n.Name, newNodeNames) {
			logger.Info("Pruning nodes...")
			err := nts.context.C(entity.NodeCollectionName).Remove(bson.M{"name": n.Name})
			if err != nil {
				logger.Error("Pruning nodes error")
				return err
			}
		}
	}
	return nil
}

func (nts *NodeSync) StartPrune() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			logger.Info("[Polling] pruning nodes...")
			nts.Prune()
		}
	}
}

func createLabelSlice(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		l := k + "=" + v
		s = append(s, l)
	}
	return s
}

func nodeInCluster(n string, list []string) bool {
	for _, l := range list {
		if l == n {
			return true
		}
	}
	return false
}
