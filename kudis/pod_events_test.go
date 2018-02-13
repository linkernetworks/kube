package kudis

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MatchStringMap(re *regexp.Regexp, matches []string) (matchNames map[string]string) {
	matchNames = make(map[string]string)
	names := re.SubexpNames()
	for idx, name := range names {
		matchNames[name] = matches[idx]
	}
	return matchNames
}

func TestPodEventRegExp(t *testing.T) {
	subscription := PodEventSubscription{
		Target:  "default",
		PodName: "podname",
	}
	matches := PodEventRegExp.FindStringSubmatch(subscription.Topic())
	assert.Len(t, matches, 3)

	var matchNames = MatchStringMap(PodEventRegExp, matches)
	t.Logf("matches: %+v", matchNames)
}

func TestPodEventSubscriptionTopic(t *testing.T) {
	subscription := PodEventSubscription{
		Target:  "default",
		PodName: "podname",
	}
	assert.Equal(t, "target:default:pod:podname:events", subscription.Topic())
}
