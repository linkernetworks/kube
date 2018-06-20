package jobtracker

import (
	"time"

	"github.com/linkernetworks/logger"

	batch "k8s.io/api/batch/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type JobStatusTracker struct {
	Clientset *kubernetes.Clientset
	stop      chan struct{}
}

type JobStatusMessage struct {
	Phase Phase
	Job   *batch.Job
}

func debugjob(job *batch.Job) {
	logger.Infof("completions=%d active=%d succeeded=%d failed=%d start at %s complete at %s", *job.Spec.Completions, job.Status.Active, job.Status.Succeeded, job.Status.Failed, job.Status.StartTime, job.Status.CompletionTime)
	if len(job.Status.Conditions) > 0 {
		logger.Infof("conditions:")
		for _, condition := range job.Status.Conditions {
			// type=Complete or Failed.
			// status=True, False, Unknown.
			logger.Infof("  type=%s reason=%s status=%s message=%s", condition.Type, condition.Reason, condition.Status, condition.Message)
		}
	}
}

func (t *JobStatusTracker) TrackUntilCompletion(namespace string, selector fields.Selector) chan JobStatusMessage {
	var o = make(chan JobStatusMessage)

	t.stop = make(chan struct{})

	var handleJobChange = func(job *batch.Job) bool {
		debugjob(job)

		completions := *job.Spec.Completions

		if job.Status.Succeeded == 0 && job.Status.Active == 0 && job.Status.Failed == 0 {
			o <- JobStatusMessage{Phase: PhasePending, Job: job}
			return false
		} else if job.Status.Succeeded == completions {
			o <- JobStatusMessage{Phase: PhaseSucceeded, Job: job}
			return true
		} else if job.Status.Failed > 0 {
			o <- JobStatusMessage{Phase: PhaseFailed, Job: job}
			return true
		} else if job.Status.Active > 0 {
			o <- JobStatusMessage{Phase: PhaseRunning, Job: job}
			return false
		} else {
			for _, condition := range job.Status.Conditions {
				if condition.Type == "Complete" {
					o <- JobStatusMessage{Phase: PhaseSucceeded, Job: job}
					return true
				} else if condition.Type == "Failed" {
					o <- JobStatusMessage{Phase: PhaseFailed, Job: job}
					return true
				}
			}

			logger.Errorf("unexpected job status: job: %+v", job)
			o <- JobStatusMessage{Phase: PhasePending, Job: job}
			return true
		}

		return false
	}

	watchlist := cache.NewListWatchFromClient(t.Clientset.BatchV1().RESTClient(), "jobs", namespace, selector)
	_, controller := cache.NewInformer(
		watchlist,
		&batch.Job{},
		time.Minute*3,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(newObj interface{}) {
				job := newObj.(*batch.Job)
				if handleJobChange(job) {
					close(o)
					t.Stop()
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				job := newObj.(*batch.Job)
				if handleJobChange(job) {
					close(o)
					t.Stop()
				}
			},
		})

	go controller.Run(t.stop)
	return o
}

func (t *JobStatusTracker) Stop() {
	if t.stop != nil {
		var e struct{}
		t.stop <- e
		close(t.stop)
		t.stop = nil
	}
}
