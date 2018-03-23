package kudis

import (
	"fmt"
	"net"
	"sync"

	"bitbucket.org/linkernetworks/aurora/src/deployment"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	pb "bitbucket.org/linkernetworks/aurora/src/kubernetes/kudis/pb"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Server struct {
	redisService      *redis.Service
	deploymentTargets deployment.DeploymentTargetMap
	running           bool
	grpcServer        *grpc.Server
	listener          net.Listener

	subscriptions sync.Map
	frames        sync.Map
}

// NewServer Kudis server
func NewServer(rds *redis.Service, dts deployment.DeploymentTargetMap) *Server {
	return &Server{
		redisService:      rds,
		deploymentTargets: dts,
	}
}

func (k *Server) GetDeploymentTarget(target string) (dt deployment.DeploymentTarget, err error) {
	var ok bool = false
	dt, ok = k.deploymentTargets[target]
	if !ok {
		return nil, fmt.Errorf("deployment target '%s' is not defined.", dt)
	}
	return dt, nil
}

func (k *Server) SubscribePodEvent(ctx context.Context, req *pb.PodEventSubscriptionRequest) (*pb.SubscriptionResponse, error) {
	target := req.GetTarget()
	dt, err := k.GetDeploymentTarget(target)
	if err != nil {
		return &pb.SubscriptionResponse{
			Success: false,
			Reason:  err.Error(),
		}, err
	}
	var subscription Subscription = NewPodEventSubscription(k.redisService, target, dt, req.GetPodName())
	success, reason, err := k.Subscribe(subscription)
	return &pb.SubscriptionResponse{Success: success, Reason: reason}, err

}

func (k *Server) SubscribePodLogs(ctx context.Context, req *pb.PodLogSubscriptionRequest) (*pb.SubscriptionResponse, error) {
	target := req.GetTarget()
	dt, err := k.GetDeploymentTarget(target)
	if err != nil {
		return &pb.SubscriptionResponse{
			Success: false,
			Reason:  err.Error(),
		}, err
	}

	var subscription Subscription = NewPodLogSubscription(
		k.redisService, target, dt,
		req.GetPodName(),
		req.GetContainerName(),
		req.GetTailLines(),
	)

	success, reason, err := k.Subscribe(subscription)
	return &pb.SubscriptionResponse{Success: success, Reason: reason}, err
}

func (k *Server) SubscribeJobLogs(ctx context.Context, req *pb.JobLogSubscriptionRequest) (*pb.SubscriptionResponse, error) {
	target := req.GetTarget()
	dt, err := k.GetDeploymentTarget(target)
	if err != nil {
		return &pb.SubscriptionResponse{
			Success: false,
			Reason:  err.Error(),
		}, err
	}

	var subscription Subscription = NewJobLogSubscription(
		k.redisService, target, dt,
		req.GetJobName(),
		req.GetContainerName(),
		req.GetTailLines(),
	)

	success, reason, err := k.Subscribe(subscription)
	return &pb.SubscriptionResponse{Success: success, Reason: reason}, err
}

func (k *Server) Subscribe(subscription Subscription) (success bool, reason string, err error) {
	if prevsub, ok := k.LoadSubscription(subscription); ok {
		if prevsub.IsRunning() {
			return true, "The subscription is already running.", nil
		}

		// load the pod log subscription object so that we can restart it again
		subscription = prevsub
	}

	if err := k.StartSubscription(subscription); err != nil {
		return false, err.Error(), err
	}

	return true, "topic subscribed successfully", nil
}

func (k *Server) LoadSubscription(subscription Subscription) (Subscription, bool) {
	// if the topic is already been subscribed then return subscribed
	val, ok := k.subscriptions.LoadOrStore(subscription.Topic(), subscription)
	return val.(Subscription), ok
}

func (k *Server) StartSubscription(subscription Subscription) error {
	var topic = subscription.Topic()
	logrus.Infof("Starting subscription: topic=%s", topic)
	if err := subscription.Start(); err != nil {
		return err
	}
	k.subscriptions.Store(topic, subscription)
	return nil
}

func (k *Server) QueryNumSubscribers(s Subscription) (int, error) {
	topic := s.Topic()

	c := k.redisService.GetConnection()
	defer c.Close()
	nums, err := c.PubSub().NumSub(topic)

	if err != nil {
		return -1, err
	}
	return nums[topic], nil
}

func (k *Server) CleanUp() error {

	k.subscriptions.Range(func(key interface{}, val interface{}) bool {
		var s = val.(Subscription)
		var topic = s.Topic()

		var frames = []int{}
		if val, ok := k.frames.LoadOrStore(topic, frames); ok {
			frames = val.([]int)
		}

		var n, err = k.QueryNumSubscribers(s)
		if err != nil {
			// redis connections error
			logger.Errorf("failed to get the number of redis subscriptions: error=%v", err)
			return true
		}

		logger.Debugf("topic:%s number of the subscribers: %d", topic, n)

		frames = append(frames, n)
		for len(frames) > 2 {
			if reduce(frames) == 0 {
				logger.Info("No redis subscription. stop streaming...")

				// load the subscription and stop the streaming
				if val, ok := k.subscriptions.Load(topic); ok {
					if sub, ok := val.(Subscription); ok {
						if sub.IsRunning() {
							if err := val.(Subscription).Stop(); err != nil {
								logger.Errorf("Failed to stop subscription: %v", err)
							}
						}

						k.subscriptions.Delete(topic)
						k.frames.Delete(topic)
						// iterate to the next subscription
						return true
					}
				}

			}
			frames = frames[1:]
		}

		k.frames.Store(topic, frames)
		return true
	})

	return nil
}

// Start starts the server
func (k *Server) Start(bind string) error {
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

func (k *Server) Stop() error {
	if !k.running {
		return nil
	}
	k.grpcServer.GracefulStop()
	k.running = false
	return k.listener.Close()
}
