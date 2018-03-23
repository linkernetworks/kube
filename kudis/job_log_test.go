package kudis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobLogSubscriptionTopic(t *testing.T) {
	subs := JobLogSubscription{
		PodLogSubscription: PodLogSubscription{
			Target:        "default",
			ContainerName: "job-container-name",
		},
		JobName: "jobName",
	}
	assert.Equal(t, "target:default:job:jobName:container:job-container-name:logs", subs.Topic())
}

func TestJobSubscriptionNewUpdateEvent(t *testing.T) {
	subc := JobLogSubscription{
		PodLogSubscription: PodLogSubscription{
			Target:        "default",
			ContainerName: "job-container-name",
		},
		JobName: "jobName",
	}
	assert.Equal(t, "target:default:job:jobName:container:job-container-name:logs", subc.Topic())

	message := "log message"
	pEvent := subc.newEvent(message)
	assert.Equal(t, "record.insert", pEvent.Type)
	assert.Equal(t, "job.container.logs", pEvent.Insert.Document)
	assert.Equal(t, "default", pEvent.Insert.Record["target"])
	assert.Equal(t, "jobName", pEvent.Insert.Record["job"])
	assert.Equal(t, "job-container-name", pEvent.Insert.Record["container"])
	assert.Equal(t, message, pEvent.Insert.Record["log"])
}

func TestMatchRegexpJobLog(t *testing.T) {
	errorMatchJobLogTopicArray := []string{
		"target:default:pod:nodesync-54d4995cdc-xt44h:container:nodesync:logs",
		"target:default:container:nodesync:logs",
		"pod:nodesync-54d4995cdc-xt44h:container:nodesync:logs",
	}

	correctMatchJobLogTopicArray := []string{
		"target:default:job:nodesync-54d4995cdc-xt44h:container:nodesync:logs",
	}

	for _, e := range errorMatchJobLogTopicArray {
		m := JobLogRegExp.MatchString(e)
		assert.False(t, m)
	}

	for _, c := range correctMatchJobLogTopicArray {
		m := JobLogRegExp.MatchString(c)
		assert.True(t, m)
	}
}
