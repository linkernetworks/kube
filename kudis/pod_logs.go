package kudis

import (
	"fmt"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	dtypes "bitbucket.org/linkernetworks/aurora/src/deployment/types"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
)

type PodLogSubscription struct {
	redis            *redis.Service
	DeploymentTarget deployment.DeploymentTarget
	running          bool

	Target        string
	PodName       string
	ContainerName string
	Log           string

	tailLines int64

	logStream deployment.ContainerLogStream
}

func NewPodLogSubscription(rds *redis.Service, target string, dt deployment.DeploymentTarget, podName string, containerName string, tl int64) *PodLogSubscription {
	return &PodLogSubscription{
		redis:            rds,
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

func (p *PodLogSubscription) NumSubscribers() (int, error) {
	topic := p.Topic()
	nums, err := p.redis.GetNumSub(topic)
	if err != nil {
		return -1, err
	}
	return nums[topic], nil
}

func (s *PodLogSubscription) Stop() error {
	return nil
}

func (s *PodLogSubscription) Start() error {
	// the pod id of the job
	deployment := dtypes.Deployment{ID: s.PodName}

	// listen the container logs from the log channel
	logC, err := s.DeploymentTarget.GetContainerLogStream(&deployment, s.ContainerName, s.tailLines)
	if err != nil {
		return err
	}

	s.logStream = logC
	s.running = true

	go s.stream()
	return nil
}

func (s *PodLogSubscription) stream() {
	logC := s.logStream
	topic := s.Topic()
	frames := []int{}
STREAM:
	for {
		select {
		case <-time.Tick(time.Second * 10):
			n, err := s.NumSubscribers()
			if err != nil {
				// redis connections error
				logger.Errorf("failed to get the number of redis subscriptions: error=%v", err)
				continue
			}

			logger.Debugf("topic:%s number of the subscribers: %d", topic, n)
			frames = append(frames, n)

			for len(frames) > 2 {
				if reduce(frames) == 0 {
					logger.Info("No redis subscription. stop streaming...")
					break STREAM
				}
				frames = frames[1:]
			}

		case lc, ok := <-logC:
			if ok {
				// publish to redis with the topic
				s.redis.PublishAndSetJSON(topic, s.newEvent(lc.Line))
			} else {
				// receive log EOF
				logger.Infof("topic:%s EOF", topic)
				break STREAM
			}
		}
	}
	s.running = false
}
