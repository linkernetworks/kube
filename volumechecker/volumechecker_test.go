package volumechecker

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/types/container"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

const (
	namespace         = "default"
	testingConfigPath = "../../../config/testing.json"
)

func TestMountSuccess(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.CreateClientset()

	id := bson.NewObjectId().Hex()
	volume := []container.Volume{}
	//Deploy a Check POD
	pod := NewAvailablePod(id, volume)
	assert.NotNil(t, pod)

	newPod, err := clientset.CoreV1().Pods(namespace).Create(&pod)
	assert.NoError(t, err)
	//Wait the POD
	//Create a channel here
	o := make(chan *v1.Pod)
	stop := make(chan struct{})
	_, controller := kubemon.WatchPods(clientset, namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}
			o <- pod
		},
	})
	go controller.Run(stop)

	err = WaitAvailiablePod(o, newPod.ObjectMeta.Name, 20)
	var e struct{}
	stop <- e
	assert.NoError(t, err)

	clientset.CoreV1().Pods(namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
}

func TestMountFail(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.CreateClientset()

	id := bson.NewObjectId().Hex()
	volume := []container.Volume{
		{
			ClaimName: "unexist",
			VolumeMount: container.VolumeMount{
				Name:      "unexist",
				MountPath: "aaa",
			},
		},
	}
	//Deploy a Check POD
	pod := NewAvailablePod(id, volume)
	assert.NotNil(t, pod)

	newPod, err := clientset.CoreV1().Pods(namespace).Create(&pod)
	assert.NoError(t, err)
	//Wait the POD
	//Create a channel here
	o := make(chan *v1.Pod)
	stop := make(chan struct{})
	_, controller := kubemon.WatchPods(clientset, namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}
			o <- pod
		},
	})
	go controller.Run(stop)

	err = WaitAvailiablePod(o, newPod.ObjectMeta.Name, 10)
	var e struct{}
	stop <- e

	assert.Error(t, err)
	assert.Equal(t, err, ErrMountUnAvailable)
	clientset.CoreV1().Pods(namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
}
