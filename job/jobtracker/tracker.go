package jobtracker

import (
	"sync"

	"bitbucket.org/linkernetworks/aurora/src/jobtranslator"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"github.com/linkernetworks/logger"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type JobTracker struct {
	clientset *kubernetes.Clientset
	namespace string
	podName   string
	stop      chan struct{}
}

type JobReceiver func(job *batchv1.Job) bool

func New(clientset *kubernetes.Clientset, namespace string, podName string) *JobTracker {
	return &JobTracker{clientset, namespace, podName, make(chan struct{})}
}

func matchJobName(obj interface{}, podName string) (*batchv1.Job, bool) {
	pod, ok := obj.(*batchv1.Job)
	if !ok {
		return nil, false
	}
	return pod, podName == pod.ObjectMeta.Name
}

// WaitForPhase wait for a pod to the specific phase
func (t *JobTracker) WaitForPhase(waitPhase Phase) {
	var m sync.Mutex
	var cv = sync.NewCond(&m)
	m.Lock()
	var handler = func(job *batchv1.Job) (stop bool) {
		logger.Infof("Waiting for job")
		if waitPhase == jobtranslator.TranslateKubernetesJobPhase(job) {
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

func (t *JobTracker) TrackAdd(callback JobReceiver) {
	_, controller := kubemon.WatchJobs(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			if pod, ok := matchJobName(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *JobTracker) TrackUpdate(callback JobReceiver) {
	_, controller := kubemon.WatchJobs(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchJobName(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *JobTracker) TrackDelete(callback JobReceiver) {
	_, controller := kubemon.WatchJobs(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchJobName(obj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *JobTracker) Stop() {
	if nil != t.stop {
		var e struct{}
		t.stop <- e
		close(t.stop)
		t.stop = nil
	}
}
