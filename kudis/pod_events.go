package kudis

import (
	"fmt"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	v1 "k8s.io/api/core/v1"
)

type PodEventSubscription struct {
	redis   *redis.Service
	watcher *deployment.KubernetesWatcher
	running bool
	stop    chan bool

	DeploymentTarget deployment.DeploymentTarget
	Target           string
	PodName          string
}

func NewPodEventSubscription(rds *redis.Service, target string, dt deployment.DeploymentTarget, podName string) *PodEventSubscription {
	return &PodEventSubscription{
		redis:            rds,
		Target:           target,
		DeploymentTarget: dt,
		PodName:          podName,
	}
}

func (s *PodEventSubscription) Topic() string {
	return fmt.Sprintf("target:%s:pod:%s:events", s.Target, s.PodName)
}

func (s *PodEventSubscription) IsRunning() bool {
	return s.running
}

func (p *PodEventSubscription) NumSubscribers() (int, error) {
	topic := p.Topic()
	nums, err := p.redis.GetNumSub(topic)
	if err != nil {
		return -1, err
	}
	return nums[topic], nil
}

func (s *PodEventSubscription) Stop() error {
	s.stop <- true
	return nil
}

func (s *PodEventSubscription) Start() error {
	var dt = s.DeploymentTarget.(*deployment.KubeDeploymentTarget)
	s.running = true
	s.watcher = dt.WatchPodEvents(s.PodName)
	s.stop = make(chan bool)
	go s.stream()
	return nil
}

func (p *PodEventSubscription) newEvent(e *v1.Event) *event.RecordEvent {
	return &event.RecordEvent{
		Type: "record.insert",
		Insert: &event.RecordInsertEvent{
			Document: "pod.events",
			Record: map[string]interface{}{
				"target":  p.Target,
				"pod":     p.PodName,
				"reason":  e.Reason,
				"message": e.Message,
				"kind":    e.Kind,
				"type":    e.Type,
				"event":   e,
			},
		},
	}
}

func (s *PodEventSubscription) stream() {
	var topic = s.Topic()
STREAM:
	for {
		select {
		case <-s.stop:
			break STREAM
		case e, ok := <-s.watcher.C:
			if ok {
				// publish to redis with the topic
				s.redis.PublishAndSetJSON(topic, s.newEvent(e))
			} else {
				break STREAM
			}
		}
	}
}
