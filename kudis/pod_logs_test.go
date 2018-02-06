package kudis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPodLogSubscriptionTopic(t *testing.T) {
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
