package nodesync

import (
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

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type NodeStats struct {
	Added   int
	Updated int
	Deleted int
}

type NodeCh chan entity.Node
type Signal chan bool

type NodeSync struct {
	clientset *kubernetes.Clientset
	context   *mongo.Session
	updateC   NodeCh
	deleteC   NodeCh
	stop      chan struct{}
	signal    Signal
	stats     NodeStats
	t         int
}

func New(clientset *kubernetes.Clientset, m *mongo.Service) *NodeSync {
	var stats NodeStats
	t, _ := strconv.Atoi(os.Getenv("NODE_RESOURCE_PERIODIC"))
	return &NodeSync{
		clientset: clientset,
		context:   m.NewSession(),
		updateC:   make(NodeCh, 1),
		deleteC:   make(NodeCh, 1),
		stop:      make(chan struct{}),
		signal:    make(Signal, 1),
		stats:     stats,
		t:         t,
	}
}

func (nts *NodeSync) Sync() Signal {
	ne := entity.Node{}
	// cleanup old node record
	logger.Info("cleaning old record")
	nts.Prune()

	logger.Info("start watching node change events...")
	// keep watch node change events
	_, controller := kubemon.WatchNodes(nts.clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Added++

			ne.LoadMeta(n)
			ne.LoadSystemInfo(n.Status.NodeInfo)
			ne.LoadAllocatableResource(n.Status.Allocatable)
			ne.LoadCapacityResource(n.Status.Capacity)

			logger.Info("[Event] nodes state added")

			nts.updateC <- ne
			select {
			case nts.signal <- true:
			}
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Deleted++

			ne.LoadMeta(n)
			logger.Info("[Event] nodes state deleted")

			nts.deleteC <- ne
			select {
			case nts.signal <- true:
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n := newObj.(*corev1.Node)
			nts.stats.Updated++

			ne.LoadMeta(n)
			ne.LoadSystemInfo(n.Status.NodeInfo)
			ne.LoadAllocatableResource(n.Status.Allocatable)
			ne.LoadCapacityResource(n.Status.Capacity)

			logger.Info("[Event] nodes state updated")

			nts.updateC <- ne
			select {
			case nts.signal <- true:
			}
		},
	})
	go controller.Run(nts.stop)
	go nts.StartPrune()
	go nts.ResourceUpdater()
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

func (nts *NodeSync) FetchNodes() []corev1.Node {
	var nodes []corev1.Node
	nodeList, _ := kubemon.GetNodes(nts.clientset)
	for _, no := range nodeList.Items {
		nodes = append(nodes, no)
	}
	return nodes
}

func (nts *NodeSync) FetchPodsByNode(name string) []corev1.Pod {
	var pods []corev1.Pod
	podList, _ := kubemon.GetPods(nts.clientset, corev1.NamespaceAll)
	for _, po := range podList.Items {
		if (po.Status.Phase == "Pending" || po.Status.Phase == "Running") && po.Spec.NodeName == name {
			pods = append(pods, po)
		}
	}
	return pods
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
	logger.Infof("Node information prune periodic time is set to %d minutes", nts.t)
	ticker := time.NewTicker(time.Duration(nts.t) * time.Minute)
	for {
		select {
		case <-ticker.C:
			nts.Prune()
		case ne := <-nts.deleteC:
			logger.Info("Receive a delete event")
			err := nts.RemoveNodeByName(ne.Name)
			if err != nil {
				logger.Error("Remove node error:", err)
			}

		}
	}
}

func (nts *NodeSync) ResourceUpdater() {
	logger.Infof("Pod resource update periodic time is set to %d minutes", nts.t)
	ticker := time.NewTicker(time.Duration(nts.t) * time.Minute)
	for {
		select {
		case <-ticker.C:
			ne := entity.Node{}
			logger.Info("fetching all nodes")
			nodes := nts.FetchNodes()
			logger.Infof("found %d nodes", len(nodes))
			for _, n := range nodes {
				pods := nts.FetchPodsByNode(n.Name)

				ne.LoadMeta(&n)
				ne.LoadSystemInfo(n.Status.NodeInfo)
				ne.LoadAllocatableResource(n.Status.Allocatable)
				ne.LoadCapacityResource(n.Status.Capacity)
				ne.UpdatePodsLimitResource(pods)
				ne.UpdatePodsRequestResource(pods)

				err := nts.UpsertNode(&ne)
				if err != nil {
					logger.Error("Upsert node error:", err)
				}
			}
		case ne := <-nts.updateC:
			logger.Info("Receive a node add/update event")
			pods := nts.FetchPodsByNode(ne.Name)

			ne.UpdatePodsLimitResource(pods)
			ne.UpdatePodsRequestResource(pods)

			err := nts.UpsertNode(&ne)
			if err != nil {
				logger.Error("Upsert node error:", err)
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
