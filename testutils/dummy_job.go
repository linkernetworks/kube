package testutils

import (
	"time"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateKubernetesDummyJob(name string) *batchV1.Job {
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

func DeployKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) (*batchV1.Job, error) {
	return clientset.BatchV1().Jobs(namespace).Create(job)
}

func DeleteKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) error {
	opts := metaV1.NewDeleteOptions(0)
	return clientset.BatchV1().Jobs(namespace).Delete(job.GetName(), opts)
}

func WaitUntilJobComplete(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) error {
	for {
		j, err := clientset.BatchV1().Jobs(namespace).Get(job.GetName(), metaV1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Succeeded > 0 || j.Status.Failed > 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func WaitUntilJobStart(clientset *kubernetes.Clientset, namespace string, job *batchV1.Job) error {
	for {
		j, err := clientset.BatchV1().Jobs(namespace).Get(job.GetName(), metaV1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Active > 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}
