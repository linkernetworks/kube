package kudis

import (
	"fmt"
	"regexp"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	v1 "k8s.io/api/core/v1"
)

var PodEventRegExp = regexp.MustCompile("target:(?P<Target>[a-z_-]+):pod:(?P<Pod>[a-z0-9_-]+):events")

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

func (s *PodEventSubscription) Regexp() *regexp.Regexp {
	return PodEventRegExp
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

	watcher := dt.WatchPodEvents(s.PodName)

	s.watcher = watcher
	s.running = true

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
STREAM:
	for {
		select {
		case <-s.stop:
			s.watcher.Stop()
			break STREAM
		case e, ok := <-s.watcher.C:
			if ok {
				logger.Debugf("received event: %v", e)
				var conn = s.redis.GetConnection()
				// publish to redis with the topic
				conn.PublishAndSetJSON(topic, s.newEvent(e))
				conn.Close()
			} else {
				break STREAM
			}
		}
	}
	logger.Debug("event stream is closed")
	s.running = false
}
