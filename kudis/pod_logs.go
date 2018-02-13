package kudis

import (
	"fmt"
	"regexp"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	dtypes "bitbucket.org/linkernetworks/aurora/src/deployment/types"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
)

var PodLogRegExp = regexp.MustCompile("target:(?P<Target>[a-z_-]+):pod:(?P<Pod>[a-z_-]+):container:(?P<Container>[a-z_-]+):logs")

type PodLogSubscription struct {
	redis            *redis.Service
	DeploymentTarget deployment.DeploymentTarget
	running          bool
	stop             chan bool

	Target        string
	PodName       string
	ContainerName string
	Log           string

	tailLines int64

	stream  deployment.ContainerLogStream
	watcher deployment.Watcher
}

func NewPodLogSubscription(rds *redis.Service, target string, dt deployment.DeploymentTarget, podName string, containerName string, tl int64) *PodLogSubscription {
	return &PodLogSubscription{
		redis:            rds,
		stop:             make(chan bool),
		Target:           target,
		DeploymentTarget: dt,
		PodName:          podName,
		ContainerName:    containerName,
		tailLines:        tl,
	}
}

func (s *PodLogSubscription) IsRunning() bool {
	return s.running
}

func (s *PodLogSubscription) Regexp() *regexp.Regexp {
	return PodLogRegExp
}

func (s *PodLogSubscription) Topic() string {
	return fmt.Sprintf("target:%s:pod:%s:container:%s:logs", s.Target, s.PodName, s.ContainerName)
}

func (p *PodLogSubscription) newEvent(text string) *event.RecordEvent {
	return &event.RecordEvent{
		Type: "record.insert",
		Insert: &event.RecordInsertEvent{
			Document: "pod.container.logs",
			Record: map[string]interface{}{
				"target":    p.Target,
				"pod":       p.PodName,
				"container": p.ContainerName,
				"log":       text,
			},
		},
	}
}

func (s *PodLogSubscription) Stop() error {
	return nil
}

func (s *PodLogSubscription) Start() error {
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

func (s *PodLogSubscription) startStream() {
	var topic = s.Topic()
	var conn = s.redis.GetConnection()
	defer conn.Close()

	// due to the defer order, the keepalive will be stopped first before the
	// connection is closed.
	var keepalive = conn.KeepAlive(10 * time.Second)
	defer keepalive.Stop()
STREAM:
	for {
		select {
		case <-s.stop:
			s.watcher.Stop()
			break STREAM
		case lc, ok := <-s.stream:
			if ok {
				// publish to redis with the topic
				conn.PublishAndSetJSON(topic, s.newEvent(lc.Line))
			} else {
				// receive log EOF
				logger.Infof("topic:%s EOF", topic)
				break STREAM
			}
		}
	}
	s.running = false
}
