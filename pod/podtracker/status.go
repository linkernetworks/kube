package podtracker

import (
	"errors"

	"github.com/linkernetworks/kube/kubemon"
	"github.com/linkernetworks/kube/pod/podutil"
	"github.com/linkernetworks/logger"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodStatusMessage struct {
	Phase   v1.PodPhase
	Error   error
	Message string
	Pod     *v1.Pod
}

type PodStatusTracker struct {
	Clientset *kubernetes.Clientset
	stop      chan struct{}
}

// TrackUntilCompletion track the pod completion status until the pod reach the completion status.
func (t *PodStatusTracker) TrackUntilCompletion(namespace string, selector fields.Selector) chan PodStatusMessage {
	t.stop = make(chan struct{})

	var o = make(chan PodStatusMessage)

	var handlePodChange = func(pod *v1.Pod) bool {
		logger.Infof("Received pod update: name=%s phase=%s", pod.Name, pod.Status.Phase)
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodUnknown {
			// send completion status message without error
			o <- PodStatusMessage{pod.Status.Phase, nil, pod.Status.Message, pod}
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
					o <- PodStatusMessage{pod.Status.Phase, errors.New(reason), cs.State.Waiting.Message, pod}
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
				t.Stop()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			if handlePodChange(pod) {
				close(o)
				t.Stop()
			}
		},
	})

	go controller.Run(t.stop)

	return o
}

func (t *PodStatusTracker) Stop() {
	if t.stop != nil {
		var e struct{}
		t.stop <- e
		close(t.stop)
		t.stop = nil
	}
}
