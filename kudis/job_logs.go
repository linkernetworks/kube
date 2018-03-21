package kudis

import (
	"fmt"
	_ "log"
	"regexp"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	dtypes "bitbucket.org/linkernetworks/aurora/src/deployment/types"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
)

var JobLogRegExp = regexp.MustCompile("target:(?P<Target>[a-z_-]+):job:(?P<Job>[a-z0-9_-]+):container:(?P<Container>[a-z0-9_-]+):logs")

type JobLogSubscription struct {
	PodLogSubscription
	JobName string
}

func NewJobLogSubscription(rds *redis.Service, target string, dt deployment.DeploymentTarget, jobName string, containerName string, tl int64) *JobLogSubscription {
	return &JobLogSubscription{
		PodLogSubscription: PodLogSubscription{
			redis:            rds,
			stop:             make(chan bool),
			Target:           target,
			DeploymentTarget: dt,
			ContainerName:    containerName,
			tailLines:        tl,
		},
		JobName: jobName,
	}
}

func (s *JobLogSubscription) Regexp() *regexp.Regexp {
	return JobLogRegExp
}

func (s *JobLogSubscription) Topic() string {
	return fmt.Sprintf("target:%s:job:%s:container:%s:logs", s.Target, s.JobName, s.ContainerName)
}

func (s *JobLogSubscription) newEvent(text string) *event.RecordEvent {
	return &event.RecordEvent{
		Type: "record.insert",
		Insert: &event.RecordInsertEvent{
			Document: "job.container.logs",
			Record: map[string]interface{}{
				"target":    s.Target,
				"job":       s.JobName,
				"container": s.ContainerName,
				"log":       text,
			},
		},
	}
}

func (s *JobLogSubscription) Start() error {
	// FIXME we need to use k8s job to find the pod name
	// kdt := s.DeploymentTarget.(*deployment.KubeDeploymentTarget)
	// clientset := kdt.GetClientset()

	// job, err := GetJob(clientset, s.Target, s.JobName)
	// if err != nil {
	// 	return err
	// }

	// log.Printf("Generate Name %v", job)
	// magic to get pod name from a job
	// s.PodName = job.Spec.Template.GetName()

	// the pod id of the job
	deployment := dtypes.Deployment{ID: s.PodName}

	// listen the container logs from the log channel
	watcher, err := s.DeploymentTarget.GetContainerLogStream(&deployment, s.ContainerName, s.tailLines)
	if err != nil {
		return err
	}

	s.stream = watcher.C
	s.watcher = watcher
	s.running = true

	go s.startStream()
	return nil
}
