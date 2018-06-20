package volumechecker

import (
	"os"
	"testing"

	"github.com/linkernetworks/kube/kubemon"
	kvolume "github.com/linkernetworks/kube/volume"
	"github.com/linkernetworks/config"
	"github.com/linkernetworks/service/kubernetes"
	"github.com/linkernetworks/types/container"

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
	if _, defined := os.LookupEnv("TEST_K8S_PVC"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.NewClientset()

	id := bson.NewObjectId().Hex()
	//Deploy a Check POD
	pod := NewVolumeCheckPod(id)
	assert.NotNil(t, pod)

	newPod, err := clientset.CoreV1().Pods(namespace).Create(&pod)
	defer clientset.CoreV1().Pods(namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
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

	err = Check(o, newPod.ObjectMeta.Name, 30)
	var e struct{}
	stop <- e
	assert.NoError(t, err)

}

func TestMountFail(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S_PVC"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.NewClientset()

	id := bson.NewObjectId().Hex()
	volumes := []container.Volume{
		{
			ClaimName: "nonexistent",
			VolumeMount: container.VolumeMount{
				Name:      "nonexistent",
				MountPath: "aaa",
			},
		},
	}
	//Deploy a Check POD
	pod := NewVolumeCheckPod(id)
	kvolume.AttachVolumesToPod(volumes, &pod)

	assert.NotNil(t, pod)

	newPod, err := clientset.CoreV1().Pods(namespace).Create(&pod)
	defer clientset.CoreV1().Pods(namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
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

	err = Check(o, newPod.ObjectMeta.Name, 10)
	var e struct{}
	stop <- e

	assert.Error(t, err)
	assert.Equal(t, err, ErrMountUnAvailable)
}
