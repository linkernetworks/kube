package nodesync

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubemon"
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

type NodeSync struct {
	clientset *kubernetes.Clientset
	context   *mongo.Session
	stop      chan struct{}
	stats     NodeStats
	t         int
}

func New(clientset *kubernetes.Clientset, m *mongo.Service) *NodeSync {
	stop := make(chan struct{})
	var stats NodeStats
	t, _ := strconv.Atoi(os.Getenv("NODE_RESOURCE_PERIODIC"))
	return &NodeSync{clientset, m.NewSession(), stop, stats, t}
}

type Signal chan bool

func (nts *NodeSync) Sync() Signal {
	signal := make(Signal, 1)
	nodeEvent := make(chan entity.Node)

	// cleanup old node record
	logger.Info("cleaning old record")
	nts.Prune()

	logger.Info("start watching node change events...")
	// keep watch node change events
	_, controller := kubemon.WatchNodes(nts.clientset, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Added++

			nodeEntity := LoadNodeEntity(n)
			logger.Info("[Event] nodes state added")

			nodeEvent <- nodeEntity
			select {
			case signal <- true:
			}
		},
		DeleteFunc: func(obj interface{}) {
			n := obj.(*corev1.Node)
			nts.stats.Deleted++

			nodeEntity := LoadNodeEntity(n)
			logger.Info("[Event] nodes state deleted")

			nodeEvent <- nodeEntity
			select {
			case signal <- true:
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n := newObj.(*corev1.Node)
			nts.stats.Updated++

			nodeEntity := LoadNodeEntity(n)
			logger.Info("[Event] nodes state updated")

			nodeEvent <- nodeEntity
			select {
			case signal <- true:
			}
		},
	})
	go controller.Run(nts.stop)
	go nts.StartPrune(nodeEvent)
	go nts.ResourceUpdater(nodeEvent)
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
			Memory:    no.Status.Allocatable.Memory().Value(),
			POD:       no.Status.Allocatable.Pods().Value(),
			NvidiaGPU: GetNvidiaGPU(&no.Status.Allocatable).Value(),
		},
		Capacity: entity.Capacity{
			CPU:       no.Status.Capacity.Cpu().MilliValue(),
			Memory:    no.Status.Capacity.Memory().Value(),
			POD:       no.Status.Capacity.Pods().Value(),
			NvidiaGPU: GetNvidiaGPU(&no.Status.Capacity).Value(),
		},
		NodeInfo: entity.NodeSystemInfo{
			MachineID:               no.Status.NodeInfo.MachineID,
			KernelVersion:           no.Status.NodeInfo.KernelVersion,
			OSImage:                 no.Status.NodeInfo.OSImage,
			ContainerRuntimeVersion: no.Status.NodeInfo.ContainerRuntimeVersion,
			KubeletVersion:          no.Status.NodeInfo.KubeletVersion,
			OperatingSystem:         no.Status.NodeInfo.OperatingSystem,
			Architecture:            no.Status.NodeInfo.Architecture,
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

func UpdateResourceInfo(node *entity.Node, pods []corev1.Pod) {
	var totalReqCPU, totalReqMem, totalReqGPU, totalReqPod int64
	var totalLimCPU, totalLimMem, totalLimGPU, totalLimPod int64
	for _, p := range pods {
		for _, c := range p.Spec.Containers {
			totalReqCPU += c.Resources.Requests.Cpu().MilliValue()
			totalReqMem += c.Resources.Requests.Memory().Value()
			totalReqGPU += GetNvidiaGPU(&c.Resources.Requests).MilliValue()
			totalReqPod += c.Resources.Requests.Pods().Value()

			totalLimCPU += c.Resources.Limits.Cpu().MilliValue()
			totalLimMem += c.Resources.Limits.Memory().Value()
			totalLimGPU += GetNvidiaGPU(&c.Resources.Limits).MilliValue()
			totalLimPod += c.Resources.Limits.Pods().MilliValue()
		}
	}
	node.Requests.CPU = totalReqCPU
	node.Requests.Memory = totalReqMem
	node.Requests.NvidiaGPU = totalReqGPU
	node.Requests.POD = totalReqPod

	node.Limits.CPU = totalLimCPU
	node.Limits.Memory = totalLimMem
	node.Limits.NvidiaGPU = totalLimGPU
	node.Limits.POD = totalLimPod
}

func (nts *NodeSync) UpsertNode(node *entity.Node) error {
	logger.Infof("%+v", node)
	update := bson.M{"$set": node}
	q := bson.M{"name": node.Name}
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

func (nts *NodeSync) StartPrune(nodeEvent chan entity.Node) {
	logger.Infof("Node information prune periodic time is set to %d minutes", nts.t)
	ticker := time.NewTicker(time.Duration(nts.t) * time.Minute)
	for {
		select {
		case <-ticker.C:
			nts.Prune()
		case ne := <-nodeEvent:
			logger.Info("Receive a delete event")
			err := nts.RemoveNodeByName(ne.Name)
			if err != nil {
				logger.Error("Remove node error:", err)
			}

		}
	}
}

func (nts *NodeSync) ResourceUpdater(nodeEvent chan entity.Node) {
	logger.Infof("Pod resource update periodic time is set to %d minutes", nts.t)
	ticker := time.NewTicker(time.Duration(nts.t) * time.Minute)
	for {
		select {
		case <-ticker.C:
			logger.Info("fetching all nodes")
			nodes := nts.FetchNodes()
			logger.Infof("found %d nodes", len(nodes))
			for _, n := range nodes {
				pods := nts.FetchPodsByNode(n.Name)
				nodeEntity := LoadNodeEntity(&n)
				UpdateResourceInfo(&nodeEntity, pods)
				err := nts.UpsertNode(&nodeEntity)
				if err != nil {
					logger.Error("Upsert node error:", err)
				}
			}
		case ne := <-nodeEvent:
			logger.Info("Receive an add/update event")
			pods := nts.FetchPodsByNode(ne.Name)
			UpdateResourceInfo(&ne, pods)
			err := nts.UpsertNode(&ne)
			if err != nil {
				logger.Error("Upsert node error:", err)
			}

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
