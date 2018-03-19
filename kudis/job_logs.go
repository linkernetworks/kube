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
	return &JobLogSubscription{
		PodLogSubscription: PodLogSubscription{
			DeploymentTarget: dt,
			Target:           target,
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

func (p *JobLogSubscription) newEvent(text string) *event.RecordEvent {
	return &event.RecordEvent{
		Type: "record.insert",
		Insert: &event.RecordInsertEvent{
			Document: "job.container.logs",
			Record: map[string]interface{}{
				"target":    p.Target,
				"job":       p.JobName,
				"container": p.ContainerName,
				"log":       text,
			},
		},
	}
}
