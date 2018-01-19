package kudis

import (
	"fmt"
	"net"
	"sync"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	pb "bitbucket.org/linkernetworks/aurora/src/kubernetes/kudis/pb"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Kudis struct {
	redisService      *redis.Service
	deploymentTargets deployment.DeploymentTargetMap
	running           bool
	grpcServer        *grpc.Server
	listener          net.Listener

	subscriptions sync.Map
}

func New(rds *redis.Service, dts deployment.DeploymentTargetMap) *Kudis {
	return &Kudis{
		redisService:      rds,
		deploymentTargets: dts,
	}
}

func (k *Kudis) GetDeploymentTarget(target string) (dt deployment.DeploymentTarget, err error) {
	var ok bool = false
	dt, ok = k.deploymentTargets[target]
	if !ok {
		return nil, fmt.Errorf("deployment target '%s' is not defined.", dt)
	}
	return dt, nil
}

func (k *Kudis) SubscribePod(ctx context.Context, req *pb.PodContainerLogRequest) (*pb.SubscriptionResponse, error) {
	target := req.GetTarget()
	dt, err := k.GetDeploymentTarget(target)

	if err != nil {
		logrus.Errorln(err)
		return &pb.SubscriptionResponse{
			Success: false,
			Reason:  err.Error(),
		}, err
	}

	subscription := NewPodLogSubscription(
		k.redisService, target, dt,
		req.GetPodName(),
		req.GetContainerName(),
		req.GetTailLines(),
	)
	topic := subscription.Topic()

	// if the topic is already been subscribed then return subscribed
	if prevsub, ok := k.subscriptions.LoadOrStore(topic, subscription); ok {
		if prevsub.(Subscription).IsRunning() {
			return &pb.SubscriptionResponse{
				Success: true,
				Reason:  "The subscription is already running.",
			}, nil
		}

		subscription, ok = prevsub.(*PodLogSubscription)
		if !ok {
			return &pb.SubscriptionResponse{
				Success: false,
				Reason:  "Failed to cast type to PodLogSubscription",
			}, nil
		}
	}

	logrus.Infof("Starting subscription: topic=%s", topic)
	if err := subscription.Start(); err != nil {
		return &pb.SubscriptionResponse{
			Success: false,
			Reason:  err.Error(),
		}, err
	}
	k.subscriptions.Store(topic, subscription)
	return &pb.SubscriptionResponse{Success: true}, nil
}

func (k *Kudis) Start(bind string) error {
	// initalize a grpc server
	k.grpcServer = grpc.NewServer()

	// register the protobuf with the gRPC server and the server implementation
	pb.RegisterSubscriptionServiceServer(k.grpcServer, k)

	logrus.Infof("gRPC server listening on %s", bind)
	c, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}
	k.listener = c
	k.running = true
	return k.grpcServer.Serve(k.listener)
}

func (k *Kudis) Stop() error {
	if !k.running {
		return nil
	}
	k.grpcServer.GracefulStop()
	k.running = false
	return k.listener.Close()
}
