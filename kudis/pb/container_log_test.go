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

	pTarget := req.GetTarget()
	assert.Equal(t, "default", pTarget)

	pPodName := req.GetPodName()
	assert.Equal(t, "john-pod-id", pPodName)

	pContainerName := req.GetContainerName()
	assert.Equal(t, "john-pod-container", pContainerName)

	pTailLines := req.GetTailLines()
	assert.EqualValues(t, 5, pTailLines)
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

	pTarget := req.GetTarget()
	assert.Equal(t, "default", pTarget)

	pJobName := req.GetJobName()
	assert.Equal(t, "john-job-id", pJobName)

	pContainerName := req.GetContainerName()
	assert.Equal(t, "john-job-container", pContainerName)

	pTailLines := req.GetTailLines()
	assert.EqualValues(t, 15, pTailLines)
}
