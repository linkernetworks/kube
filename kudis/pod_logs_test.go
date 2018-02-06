package kudis

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	redissvc "bitbucket.org/linkernetworks/aurora/src/service/redis"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestPodSubscriptionTopic(t *testing.T) {
	jSubscription := PodLogSubscription{
		Target:        "default",
		PodName:       "DPID",
		ContainerName: "log-collector",
	}
	assert.Equal(t, "target:default:pod:DPID:container:log-collector:logs", jSubscription.Topic())
}

func TestPodSubscriptionNewUpdateEvent(t *testing.T) {
	subc := PodLogSubscription{
		Target:        "default",
		PodName:       "DPID",
		ContainerName: "johnlin",
	}
	assert.Equal(t, "target:default:pod:DPID:container:johnlin:logs", subc.Topic())

	message := "log message"
	pEvent := subc.newUpdateEvent(message)
	assert.Equal(t, "record.update", pEvent.Type)
	assert.Equal(t, "pod.container.logs", pEvent.ContainerLog.Document)
	assert.Equal(t, "default", pEvent.ContainerLog.Target)
	assert.Equal(t, "DPID", pEvent.ContainerLog.DeploymentId)
	assert.Equal(t, "johnlin", pEvent.ContainerLog.ContainerName)
	assert.Equal(t, message, pEvent.ContainerLog.Log)
}

func TestPodSubscriptionGetNumSub(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	rd := redissvc.New(cf.Redis)

	s := PodLogSubscription{
		redis:         rd,
		Target:        "default",
		PodName:       "job-00xx00",
		ContainerName: "log-collector",
	}
	topic := s.Topic()

	rdc := rd.Pool.Get()
	psc := redis.PubSubConn{Conn: rdc}
	err := psc.Subscribe(topic)
	assert.NoError(t, err)

	num, err := s.NumSubscribers()
	assert.NoError(t, err)
	assert.Equal(t, 1, num)

	err = psc.Unsubscribe(topic)
	assert.NoError(t, err)

	num, err = s.NumSubscribers()
	assert.NoError(t, err)
	assert.Equal(t, 0, num)
}
