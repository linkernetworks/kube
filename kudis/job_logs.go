package kudis

import (
	"fmt"
	"regexp"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
)

var JobLogRegExp = regexp.MustCompile("target:(?P<Target>[a-z_-]+):job:(?P<Job>[a-z0-9_-]+):container:(?P<Container>[a-z0-9_-]+):logs")

type JobLogSubscription struct {
	PodLogSubscription
	JobName string
}

func NewJobLogSubscription(rds *redis.Service, target string, dt deployment.DeploymentTarget, jobName string, containerName string, tl int64) *JobLogSubscription {
	kdt := dt.(*deployment.KubeDeploymentTarget)
	clientset := kdt.GetClientset()

	job, _ := GetJob(clientset, target, jobName)

	return &JobLogSubscription{
		PodLogSubscription: PodLogSubscription{
			redis:            rds,
			stop:             make(chan bool),
			Target:           target,
			DeploymentTarget: dt,
			PodName:          job.Spec.Template.ObjectMeta.GetName(),
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
