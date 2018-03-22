package kudis

import (
	"fmt"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createKubernetesDummyJob(name string) *batchV1.Job {
	return &batchV1.Job{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: batchV1.JobSpec{
			Template: coreV1.PodTemplateSpec{
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Name:  name,
							Image: "hello-world:latest",
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
}

func deployKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) (*batchV1.Job, error) {
	return clientset.BatchV1().Jobs(namespace).Create(job)
}

func deleteKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) error {
	opts := metaV1.NewDeleteOptions(0)
	return clientset.BatchV1().Jobs(namespace).Delete(job.GetName(), opts)
}

func GetPodByJobName(clientset *kubernetes.Clientset, namespace, jobName string) (*coreV1.Pod, error) {
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
	// since we might get many pods but we always return the latest one
	return &list.Items[0], nil
}
