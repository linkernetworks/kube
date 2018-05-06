package summary

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	pb "bitbucket.org/linkernetworks/aurora/src/jobserver/pb"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNumGPUFromPod(t *testing.T) {
	job := entity.Job{
		Request: &pb.JobRequest{
			Training: &pb.JobTraining{
				Resources: map[string]*pb.Resource{
					"test": {
						Request: &pb.ResourceRequest{
							Gpu: "10",
						},
					},
				},
			},
		},
	}
	num := GetNumGPUFromJob(job)
	assert.Equal(t, 10, num)

	jobWith2Containers := entity.Job{
		Request: &pb.JobRequest{
			Training: &pb.JobTraining{
				Resources: map[string]*pb.Resource{
					"test": {
						Request: &pb.ResourceRequest{
							Gpu: "10",
						},
					},
					"test2": {
						Request: &pb.ResourceRequest{
							Gpu: "20",
						},
					},
				},
			},
		},
	}

	num = GetNumGPUFromJob(jobWith2Containers)
	assert.Equal(t, 30, num)
}
