package kudis

import (
	batchV1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetJob(clientset *kubernetes.Clientset, namespace, jobName string) (*batchV1.Job, error) {
	opts := metav1.GetOptions{}
	return clientset.BatchV1().Jobs(namespace).Get(jobName, opts)
}
