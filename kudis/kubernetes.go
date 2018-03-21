package kudis

import (
	"log"

	batchV1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetJob(clientset *kubernetes.Clientset, namespace, jobName string) (*batchV1.Job, error) {
	opts := metav1.GetOptions{}
	job, err := clientset.BatchV1().Jobs(namespace).Get(jobName, opts)
	if errors.IsNotFound(err) {
		log.Printf("Pod not found\n")
		return nil, err
	} else if err != nil {
		log.Printf("clientset error: %+v", err.Error())
		return nil, err
	}
	return job, nil
}
