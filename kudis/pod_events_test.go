package kudis

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPodEventSubscriptionTopic(t *testing.T) {
	subscription := PodEventSubscription{
		Target:  "default",
		PodName: "DPID",
	}
	assert.Equal(t, "target:default:pod:DPID:events", subscription.Topic())
}
