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

const ErrImagePullBackOff = "ImagePullBackOff"

// Unable to inspect image
const ErrImageInspect = "ImageInspectError"

// General image pull error
const ErrImagePull = "ErrImagePull"

// Required Image is absent on host and PullPolicy is NeverPullImage
const ErrImageNeverPull = "ErrImageNeverPull"

// Get http error when pulling image from registry
const RegistryUnavailable = "RegistryUnavailable"

// Unable to parse the image name.
const ErrInvalidImageName = "InvalidImageName"

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

func matchPod(obj interface{}, podName string) (*v1.Pod, bool) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil, false
	}
	return pod, podName == pod.ObjectMeta.Name
}

func (t *PodTracker) WaitFor(waitPhase v1.PodPhase) *sync.Cond {
	cv := &sync.Cond{}
	t.Track(func(pod *v1.Pod) (stop bool) {
		logger.Infof("Tracking pod=%s phase=%s", t.podName, pod.Status.Phase)

		switch pod.Status.Phase {
		case "Pending":
			// Check all containers status in a pod. when it failed to start we should stop tracking.
			cslist := podutil.FindWaitingContainerStatuses(pod)
			for _, cs := range cslist {
				// Possible values are: PodInitializing, ErrImagePull, ImagePullBackOff
				reason := cs.State.Waiting.Reason
				switch reason {
				case "PodInitializing", "ContainerCreating":
					// Skip the standard states

				case ErrImageInspect,
					ErrImagePullBackOff,
					ErrImagePull,
					ErrImageNeverPull,
					RegistryUnavailable,
					ErrInvalidImageName:
					logger.Errorf("Container %s is waiting. Reason=%s", cs.ContainerID, reason)

					// stop tracking
					stop = true
					return stop
				}
			}

		// Stop the tracker if the status is completion status.
		// Terminating won't be catched
		case "Running", "Failed", "Succeeded", "Unknown", "Terminating":
			stop = true
			return stop
		}

		stop = false
		return stop
	})
	return cv
}

func (t *PodTracker) Track(callback PodReceiver) {
	_, controller := kubemon.WatchPods(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			if pod, ok := matchPod(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchPod(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},

		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchPod(obj, t.podName); ok {
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
