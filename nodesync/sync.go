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
	logger.Infof("found %d nodes", len(nodes))
	for _, n := range nodes {
		// fetch all pods on each node
		pods := nts.FetchPodsByNode(n)
		nodeEntity := LoadNodeEntity(&n)
		UpdateResourceInfo(&nodeEntity, pods)
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

			nodeEntity := LoadNodeEntity(n)
			logger.Info("[Event] nodes state added")
			err := nts.UpsertNode(&nodeEntity)
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

			nodeEntity := LoadNodeEntity(n)
			logger.Info("[Event] nodes state deleted")
			err := nts.RemoveNodeByName(nodeEntity.Name)
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

			nodeEntity := LoadNodeEntity(n)
			logger.Info("[Event] nodes state updated")
			err := nts.UpsertNode(&nodeEntity)
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
	go nts.StartUpdatePodResource()
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
			totalReqMem += c.Resources.Requests.Memory().MilliValue()
			totalReqGPU += c.Resources.Requests.NvidiaGPU().MilliValue()
			totalReqPod += c.Resources.Requests.Pods().Value()

			totalLimCPU += c.Resources.Limits.Cpu().MilliValue()
			totalLimMem += c.Resources.Limits.Memory().MilliValue()
			totalLimGPU += c.Resources.Limits.NvidiaGPU().MilliValue()
			totalLimPod += c.Resources.Limits.Pods().MilliValue()
		}
	}
	node.Requests.CPU = totalReqCPU
	node.Requests.Memory = totalReqMem
	node.Requests.NvidiaGPU = totalReqCPU
	node.Requests.POD = totalReqCPU

	node.Limits.CPU = totalLimCPU
	node.Limits.Memory = totalLimMem
	node.Limits.NvidiaGPU = totalLimGPU
	node.Limits.POD = totalLimPod
	// logger.Info(node)
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

func (nts *NodeSync) FetchNodes() []corev1.Node {
	var nodes []corev1.Node
	nodeList, _ := kubemon.GetNodes(nts.clientset)
	for _, no := range nodeList.Items {
		// logger.Info(no.GetName())
		nodes = append(nodes, no)
	}
	return nodes
}

func (nts *NodeSync) FetchPodsByNode(no corev1.Node) []corev1.Pod {
	var pods []corev1.Pod
	podList, _ := kubemon.GetPods(nts.clientset, corev1.NamespaceAll)
	for _, po := range podList.Items {
		if po.Status.Phase == "Running" && po.Spec.NodeName == no.GetName() {
			// logger.Info(no.GetName() + ">>>>" + po.GetName())
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
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			logger.Info("[Polling] start pruning nodes...")
			nts.Prune()
		}
	}
}

func (nts *NodeSync) StartUpdatePodResource() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			logger.Info("[Polling] start updating pod resources...")
			nodes := nts.FetchNodes()
			for _, n := range nodes {
				pods := nts.FetchPodsByNode(n)
				nodeEntity := LoadNodeEntity(&n)
				UpdateResourceInfo(&nodeEntity, pods)
				err := nts.UpsertNode(&nodeEntity)
				if err != nil {
					logger.Error("Upsert node error:", err)
				}
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
