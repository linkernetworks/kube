package kudis

import (
	"fmt"
	"time"

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
		stop:             make(chan bool),
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

func (s *PodEventSubscription) Stop() error {
	s.stop <- true
	return nil
}

func (s *PodEventSubscription) Start() error {
	var dt = s.DeploymentTarget.(*deployment.KubeDeploymentTarget)
	s.running = true
	s.watcher = dt.WatchPodEvents(s.PodName)
	go s.startStream()
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

func (s *PodEventSubscription) startStream() {
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
		case e, ok := <-s.watcher.C:
			if ok {
				// publish to redis with the topic
				conn.PublishAndSetJSON(topic, s.newEvent(e))
			} else {
				break STREAM
			}
		}
	}
}
