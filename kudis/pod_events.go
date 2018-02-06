package kudis

import (
	"fmt"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	v1 "k8s.io/api/core/v1"
)

type PodEventSubscription struct {
	redis            *redis.Service
	DeploymentTarget deployment.DeploymentTarget
	running          bool
	Target           string
	PodName          string
}

func (s *PodEventSubscription) Topic() string {
	return fmt.Sprintf("target:%s:pod:%s:events:%s", s.Target, s.PodName)
}

func (s *PodEventSubscription) Start() error {
	var dt = s.DeploymentTarget.(*deployment.KubeDeploymentTarget)
	var watcher = dt.WatchPodEvents(s.PodName)
	s.running = true
	go s.stream(watcher)
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

func (s *PodEventSubscription) stream(watcher *deployment.KubernetesWatcher) {
	var topic = s.Topic()
STREAM:
	for {
		select {
		case e, ok := <-watcher.C:
			if ok {
				// publish to redis with the topic
				s.redis.PublishAndSetJSON(topic, s.newEvent(e))
			} else {
				break STREAM
			}
		}
	}
}
