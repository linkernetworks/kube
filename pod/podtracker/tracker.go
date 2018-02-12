package podtracker

import (
	"sync"

	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podutil"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodTracker struct {
	clientset *kubernetes.Clientset
	namespace string
	podName   string
	stop      chan struct{}
}

type PodReceiver func(pod *v1.Pod) bool

func New(clientset *kubernetes.Clientset, namespace string, podName string) *PodTracker {
	return &PodTracker{clientset, namespace, podName, make(chan struct{})}
}

func matchPodName(obj interface{}, podName string) (*v1.Pod, bool) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil, false
	}
	return pod, podName == pod.ObjectMeta.Name
}

// WaitFor wait for a pod to the specific phase
func (t *PodTracker) WaitFor(waitPhase v1.PodPhase) *sync.Cond {
	cv := &sync.Cond{}
	t.Track(func(pod *v1.Pod) (stop bool) {
		logger.Infof("Wait for pod=%s phase=%s wait=%s", t.podName, pod.Status.Phase)ci, waitPhase)
		if waitPhase == pod.Status.Phase {
			cv.Signal()
			stop = true
		}
		return stop
	})
	return cv
}

func (t *PodTracker) Track(callback PodReceiver) {
	_, controller := kubemon.WatchPods(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			if pod, ok := matchPodName(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchPodName(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},

		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchPodName(obj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *PodTracker) Stop() {
	var e struct{}
	t.stop <- e
}
