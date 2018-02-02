package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/nvidia"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"gopkg.in/mgo.v2/bson"
	corev1 "k8s.io/api/core/v1"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

type NodeStats struct {
	Added   int
	Updated int
	Deleted int
}

type NodeCh chan *entity.Node
type Signal chan bool

type NodeSync struct {
	clientset *kubernetes.Clientset
	context   *mongo.Session
	dt        *deployment.KubeDeploymentTarget
	updateC   NodeCh
	deleteC   NodeCh
	stop      chan struct{}
	signal    Signal
	stats     NodeStats
	t         int
}

func New(clientset *kubernetes.Clientset, m *mongo.Service, dt *deployment.KubeDeploymentTarget) *NodeSync {
	var stats NodeStats

	t := ostrconv.Atoi(os.Getenv("NODE_RESOURCE_PERIODIC"))
	if len(t) == 0 {
		t = 3
	}

	return &NodeSync{
		clientset: clientset,
		context:   m.NewSession(),
		dt:        dt,
		updateC:   make(NodeCh, 10),
		deleteC:   make(NodeCh, 10),
		stop:      make(chan struct{}),
		signal:    make(Signal, 1),
		stats:     stats,
		t:         t,
	}
}

func (nts *NodeSync) Sync() Signal {
	// cleanup old node record
	logger.Info("cleaning old record")
	nts.Prune()

	logger.Info("start watching node change events...")
	// keep watch node change events
	_, controller := kubemon.WatchNodes(nts.clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ne := &entity.Node{}
			n := obj.(*corev1.Node)
			nts.stats.Added++

			ne.LoadMeta(n)
			ne.LoadSystemInfo(n.Status.NodeInfo)
			ne.LoadAllocatableResource(n.Status.Allocatable)
			ne.LoadCapacityResource(n.Status.Capacity)

			logger.Infof("[Event] node %s added", ne.Name)

			nts.updateC <- ne
			select {
			case nts.signal <- true:
			}
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Deleted++

			logger.Infof("[Event] node %s deleted", n.Name)

			nts.deleteC <- &entity.Node{Name: n.Name}
			select {
			case nts.signal <- true:
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ne := &entity.Node{}
			n := newObj.(*corev1.Node)
			nts.stats.Updated++

			ne.LoadMeta(n)
			ne.LoadSystemInfo(n.Status.NodeInfo)
			ne.LoadAllocatableResource(n.Status.Allocatable)
			ne.LoadCapacityResource(n.Status.Capacity)

			logger.Infof("[Event] node %s updated", ne.Name)

			nts.updateC <- ne
			select {
			case nts.signal <- true:
			}
		},
	})
	go controller.Run(nts.stop)
	go nts.NodeRoutine()
	go nts.NodeEventHandler()
	return nts.signal
}

func (nts *NodeSync) Wait() {
	<-nts.stop
}

func (nts *NodeSync) Stop() {
	var e struct{}
	nts.stop <- e
}

func (nts *NodeSync) UpsertNode(node *entity.Node) error {
	update := bson.M{"$set": node}
	q := bson.M{"name": node.Name}
	// debug use showing log messages
	logger.Infof("%#v", node)
	_, err := nts.context.C(entity.NodeCollectionName).Upsert(q, update)
	return err
}

func (nts *NodeSync) RemoveNodeByName(name string) error {
	q := bson.M{"name": name}
	return nts.context.C(entity.NodeCollectionName).Remove(q)
}

func (nts *NodeSync) Prune() error {
	var newNodeNames []string
	var err error
	nodesList, err := nts.dt.GetNodes()
	if err != nil {
		return err
	}
	for _, n := range nodesList {
		newNodeNames = append(newNodeNames, n.Name)
	}
	nodes := []entity.Node{}
	err = nts.context.C(entity.NodeCollectionName).Find(nil).Select(bson.M{"name": 1}).All(&nodes)
	if err != nil {
		return err
	}
	// check mongodb record if node doesn't exist in current cluster than remove the record
	for _, n := range nodes {
		if !nodeInCluster(n.Name, newNodeNames) {
			logger.Info("Pruning node %s", n.Name)
			nts.deleteC <- &entity.Node{Name: n.Name}
		}
	}
	return nil
}

func (nts *NodeSync) NodeRoutine() {
	logger.Infof("Node routine periodic time is set to %d minutes", nts.t)
	ticker := time.NewTicker(time.Duration(nts.t) * time.Minute)
	for {
		select {
		case <-ticker.C:
			// clean routine
			nts.Prune()
			// update routine
			nodes, err := nts.dt.GetNodesResource()
			logger.Infof("fetching all nodes. found %d nodes", len(nodes))
			if err != nil {
				logger.Errorf("Get nodes resources fail: %v", err)

			}
			for _, n := range nodes {
				nts.updateC <- n
			}

		}
	}

}

func (nts *NodeSync) NodeEventHandler() {
	for {
		select {
		case ne := <-nts.updateC:
			logger.Info("Receive a update event")
			// fetch all active pods on a node
			pods := nts.dt.FetchActivePodsByNode(ne.Name)

			ne.UpdatePodsLimitResource(pods)
			ne.UpdatePodsRequestResource(pods)

			err := nts.UpsertNode(ne)
			if err != nil {
				logger.Error("Upsert node error:", err)
			}
		case ne := <-nts.deleteC:
			logger.Info("Receive a delete event")
			err := nts.RemoveNodeByName(ne.Name)
			if err != nil {
				logger.Error("Remove node error:", err)
			}

		}
	}
}

func nodeInCluster(n string, list []string) bool {
	for _, l := range list {
		if l == n {
			return true
		}
	}
	return false
}
