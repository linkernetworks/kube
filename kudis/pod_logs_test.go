package kudis

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	redissvc "bitbucket.org/linkernetworks/aurora/src/service/redis"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestPodLogsSubscriptionTopic(t *testing.T) {
	subs := PodLogSubscription{
		Target:        "default",
		PodName:       "DPID",
		ContainerName: "log-collector",
	}
	assert.Equal(t, "target:default:pod:DPID:container:log-collector:logs", subs.Topic())
}

func TestPodSubscriptionNewUpdateEvent(t *testing.T) {
	subc := PodLogSubscription{
		Target:        "default",
		PodName:       "DPID",
		ContainerName: "johnlin",
	}
	assert.Equal(t, "target:default:pod:DPID:container:johnlin:logs", subc.Topic())

	message := "log message"
	pEvent := subc.newEvent(message)
	assert.Equal(t, "record.insert", pEvent.Type)
	assert.Equal(t, "pod.container.logs", pEvent.Insert.Document)
	assert.Equal(t, "default", pEvent.Insert.Record["target"])
	assert.Equal(t, "DPID", pEvent.Insert.Record["pod"])
	assert.Equal(t, "johnlin", pEvent.Insert.Record["container"])
	assert.Equal(t, message, pEvent.Insert.Record["log"])
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
