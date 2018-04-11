package kudis

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodsByJobName(clientset *kubernetes.Clientset, namespace, jobName string) ([]v1.Pod, error) {
	label := "job-name=" + jobName

	opts := metaV1.ListOptions{
		LabelSelector: label,
	}
	list, err := clientset.CoreV1().Pods(namespace).List(opts)
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("could not find job for pod in namespace %s with label: %v", namespace, label)
	}

	return list.Items, nil
}
