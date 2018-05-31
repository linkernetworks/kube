package testutils

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateKubernetesDummyJob(ns string, name string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    ns,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
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

func DeployKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	return clientset.BatchV1().Jobs(namespace).Create(job)
}

func DeleteKubenetesJob(clientset *kubernetes.Clientset, namespace string, job *batchv1.Job) error {
	opts := metav1.NewDeleteOptions(0)
	return clientset.BatchV1().Jobs(namespace).Delete(job.GetName(), opts)
}

func WaitUntilJobComplete(clientset *kubernetes.Clientset, namespace string, job *batchv1.Job) error {
	for {
		j, err := clientset.BatchV1().Jobs(namespace).Get(job.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Succeeded > 0 || j.Status.Failed > 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}
