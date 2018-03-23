package messages

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalPodLogSubscriptionRequestJSON(t *testing.T) {
	c := `
	{
		"Target": "default",
		"ContainerName": "john-pod-container",
		"PodName": "john-pod-id",
		"TailLines": 5
	}
	`
	req := PodLogSubscriptionRequest{}
	err := jsonpb.UnmarshalString(c, &req)
	if err != nil {
		t.Fatal(err)
	}

	Target := req.GetTarget()
	assert.Equal(t, "default", Target)

	PodName := req.GetPodName()
	assert.Equal(t, "john-pod-id", PodName)

	ContainerName := req.GetContainerName()
	assert.Equal(t, "john-pod-container", ContainerName)

	TailLines := req.GetTailLines()
	assert.EqualValues(t, 5, TailLines)
}

func TestUnmarshalJobLogSubscriptionRequestJSON(t *testing.T) {
	c := `
	{
		"Target": "default",
		"ContainerName": "john-job-container",
		"JobName": "john-job-id",
		"TailLines": 15
	}
	`
	req := JobLogSubscriptionRequest{}
	err := jsonpb.UnmarshalString(c, &req)
	if err != nil {
		t.Fatal(err)
	}

	Target := req.GetTarget()
	assert.Equal(t, "default", Target)

	JobName := req.GetJobName()
	assert.Equal(t, "john-job-id", JobName)

	ContainerName := req.GetContainerName()
	assert.Equal(t, "john-job-container", ContainerName)

	TailLines := req.GetTailLines()
	assert.EqualValues(t, 15, TailLines)
}
