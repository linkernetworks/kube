package outcluster

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DiscoverVisibleNode(clientset *kubernetes.Clientset) (*v1.Node, string, error) {
	nodesList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, "", err
	}
	for _, n := range nodesList.Items {
		for _, addr := range n.Status.Addresses {
			if addr.Type == "ExternalIP" && addr.Address != "" {
				return &n, addr.Address, nil
			}
		}
	}
	return nil, "", nil
}
