package messages

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalPodContainerLogRequestJSON(t *testing.T) {
	c := `
	{
		"Target": "default",
		"ContainerName": "john-container",
		"PodName": "john-pod-id",
		"TailLines": 5
	}
	`
	req := PodContainerLogRequest{}
	err := jsonpb.UnmarshalString(c, &req)
	if err != nil {
		t.Fatal(err)
	}

	pTarget := req.GetTarget()
	assert.Equal(t, "default", pTarget)

	pContainerName := req.GetContainerName()
	assert.Equal(t, "john-container", pContainerName)

}