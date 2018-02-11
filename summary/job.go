package summary

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"gopkg.in/mgo.v2/bson"
	_ "log"
	"strconv"
)

type JobGPUUsageSummary []JobGPUUsage

type JobGPUUsage struct {
	Job    *entity.Job
	NumGPU int
}

func QueryCurrentGpuUsageByUser(context *mongo.Session, uid bson.ObjectId) (JobGPUUsageSummary, error) {
	var usageSummary []JobGPUUsage
	var jobs []entity.Job
	err := context.C(entity.JobCollectionName).Find(bson.M{"CreatedBy": uid}).All(&jobs)
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		usage := JobGPUUsage{
			Job:    &job,
			NumGPU: GetNumGPUFromJob(job),
		}
		usageSummary = append(usageSummary, usage)
	}
	return usageSummary, nil
}

func GetNumGPUFromJob(job entity.Job) int {
	sum := 0
	// all containers in a pod
	for _, r := range job.Request.Training.Resources {
		num, _ := strconv.Atoi(r.Request.GetGpu())
		sum += num
	}
	return sum
}
