package outcluster

import (
	"log"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DiscoverVisibleNode(clientset *kubernetes.Clientset) (node *v1.Node, address string, err error) {
	node, address, err = DiscoverVisibleNodeByAddressType(clientset, "ExternalIP")
	if node == nil || address == "" || err != nil {
		return DiscoverVisibleNodeByAddressType(clientset, "InternalIP")
	}
	return
}

func DiscoverVisibleNodeByAddressType(clientset *kubernetes.Clientset, addressType string) (*v1.Node, string, error) {
	nodesList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, "", err
	}
	for _, n := range nodesList.Items {
		for _, addr := range n.Status.Addresses {
			if string(addr.Type) == addressType && addr.Address != "" {
				log.Printf(addr.Address)
				return &n, "localhost", nil
			}
		}
	}
	return nil, "", nil
}
