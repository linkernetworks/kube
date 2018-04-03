package podtracker

import (
	"errors"
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

// WaitForPhase wait for a pod to the specific phase
func (t *PodTracker) WaitForPhase(waitPhase v1.PodPhase) {
	var m sync.Mutex
	var cv = sync.NewCond(&m)
	m.Lock()

	var handler = func(pod *v1.Pod) (stop bool) {
		logger.Infof("Checking pod phase pod=%s current=%s expect=%s", t.podName, pod.Status.Phase, waitPhase)
		if waitPhase == pod.Status.Phase {
			m.Lock()
			cv.Broadcast()
			m.Unlock()
			stop = true
		}
		return stop
	}

	t.TrackAdd(handler)
	t.TrackUpdate(handler)
	cv.Wait()
	m.Unlock()
}

func (t *PodTracker) TrackAdd(callback PodReceiver) {
	_, controller := kubemon.WatchPods(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			if pod, ok := matchPodName(newObj, t.podName); ok {
				logger.Debugf("Received pod add: %s %s", pod.Name, pod.Status.Phase)
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *PodTracker) TrackUpdate(callback PodReceiver) {
	_, controller := kubemon.WatchPods(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchPodName(newObj, t.podName); ok {
				logger.Debugf("Received pod update: %s %s", pod.Name, pod.Status.Phase)
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *PodTracker) TrackDelete(callback PodReceiver) {
	_, controller := kubemon.WatchPods(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchPodName(obj, t.podName); ok {
				logger.Debugf("Received pod delete: %s %s", pod.Name, pod.Status.Phase)
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *PodTracker) Stop() {
	if nil != t.stop {
		var e struct{}
		t.stop <- e
		close(t.stop)
		t.stop = nil
	}
}

type PodStatusMessage struct {
	Phase   v1.PodPhase
	Error   error
	Message string
}

type PodStatusTracker struct {
	Clientset *kubernetes.Clientset
}

// TrackUntilCompletion track the pod completion status until the pod reach the completion status.
func (t *PodStatusTracker) TrackUntilCompletion(namespace string, selector fields.Selector) (chan PodStatusMessage, chan struct{}) {
	var e struct{}

	var stop = make(chan struct{})
	var o = make(chan PodStatusMessage)

	var handlePodChange = func(pod *v1.Pod) bool {
		logger.Infof("Received pod update: name=%s phase=%s", pod.Name, pod.Status.Phase)
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodUnknown {
			// send completion status message without error
			o <- PodStatusMessage{pod.Status.Phase, nil, pod.Status.Message}
			return true

		} else if pod.Status.Phase == v1.PodPending {
			statuses := podutil.FindWaitingContainerStatuses(pod)
			for _, cs := range statuses {
				reason := cs.State.Waiting.Reason
				switch reason {
				// the reasons below are copied from kubernetes/pkg/kubelet/images/types.go
				case "ErrImageInspect",
					"ErrImagePullBackOff",
					"ErrImagePull",
					"ErrImageNeverPull",
					"RegistryUnavailable",
					"ErrInvalidImageName",

					// from kubernetes/pkg/kubelet/container/sync_result.go
					"CrashLoopBackOff":
					o <- PodStatusMessage{pod.Status.Phase, errors.New(reason), cs.State.Waiting.Message}
					return true
				}
			}
		}
		return false
	}

	_, controller := kubemon.WatchPods(t.Clientset, namespace, selector, cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			pod := newObj.(*v1.Pod)
			if handlePodChange(pod) {
				close(o)
				stop <- e
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			if handlePodChange(pod) {
				close(o)
				stop <- e
			}
		},
	})

	go controller.Run(stop)

	return o, stop
}
