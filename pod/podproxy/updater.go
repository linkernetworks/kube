package podproxy

import (
	"errors"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podutil"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrPortNotFound = errors.New("Container port not found.")

type ProxyInfoProvider interface {
	Host() string
	Port() string
	BaseURL() string
}

func NewPodInfo(pod *v1.Pod) *entity.PodInfo {
	return &entity.PodInfo{
		Phase:     pod.Status.Phase,
		Message:   pod.Status.Message,
		Reason:    pod.Status.Reason,
		StartTime: pod.Status.StartTime,
	}
}

type SpawnableDocument interface {
	types.DeploymentIDProvider
	GetID() bson.ObjectId
	Topic() string
	NewUpdateEvent(info bson.M) *event.RecordEvent
}

type DocumentProxyInfoUpdater struct {
	Clientset *kubernetes.Clientset
	Namespace string

	Redis   *redis.Service
	Session *mongo.Session

	// Which mongo collection to update
	CollectionName string

	// The PortName of the Pod
	PortName string
}

func (u *DocumentProxyInfoUpdater) getPod(doc SpawnableDocument) (*v1.Pod, error) {
	return u.Clientset.CoreV1().Pods(u.Namespace).Get(doc.DeploymentID(), metav1.GetOptions{})
}

// TrackAndSync tracks the pod of the owner document and returns a pod tracker
// The following comments are copied from the kubernetes repository:
//
//     PodPending means the pod has been accepted by the system, but one or more of the containers
//     has not been started. This includes time before being bound to a node, as well as time spent
//     pulling images onto the host.
//
//    		PodPending PodPhase = "Pending"
//
//     PodRunning means the pod has been bound to a node and all of the containers have been started.
//     At least one container is still running or is in the process of being restarted.
//
//    		PodRunning PodPhase = "Running"
//
//     PodSucceeded means that all containers in the pod have voluntarily terminated
//     with a container exit code of 0, and the system is not going to restart any of these containers.
//
//    		PodSucceeded PodPhase = "Succeeded"
//
//     PodFailed means that all containers in the pod have terminated, and at least one container has
//     terminated in a failure (exited with a non-zero exit code or was stopped by the system).
//
//    		PodFailed PodPhase = "Failed"
//
//     PodUnknown means that for some reason the state of the pod could not be obtained, typically due
//     to an error in communicating with the host of the pod.
//
//    		PodUnknown PodPhase = "Unknown"
//
// See package "k8s.io/kubernetes/pkg/apis/core/types.go" for more details.

func (u *DocumentProxyInfoUpdater) TrackAndSync(doc SpawnableDocument) (*podtracker.PodTracker, error) {
	podName := doc.DeploymentID()

	tracker := podtracker.New(u.Clientset, u.Namespace, podName)

	tracker.Track(func(pod *v1.Pod) (stop bool) {
		phase := pod.Status.Phase
		logger.Infof("Tracking notebook pod=%s phase=%s", podName, phase)

		switch phase {
		case v1.PodPending:
			if err := u.SyncWithPod(doc, pod); err != nil {
				logger.Errorf("Failed to sync document: error=%v", err)
			}

			// Check all containers status in a pod. can't be ErrImagePull or ImagePullBackOff
			cslist := podutil.FindWaitingContainerStatuses(pod)
			for _, cs := range cslist {
				reason := cs.State.Waiting.Reason
				switch reason {
				case "PodInitializing", "ContainerCreating":
					// Skip the standard states
					logger.Infof("Container %s state is %s", cs.ContainerID, reason)

				case "ErrImageInspect",
					"ErrImagePullBackOff",
					"ErrImagePull",
					"ErrImageNeverPull",
					"RegistryUnavailable",
					"ErrInvalidImageName":
					logger.Errorf("Container %s is waiting. Reason=%s", cs.ContainerID, reason)

					// stop tracking
					stop = true
					return stop
				}
			}

		// Stop the tracker if the status is completion status.
		// Terminating won't be catched
		case v1.PodRunning, v1.PodFailed, v1.PodSucceeded, v1.PodUnknown:
			if err := u.SyncWithPod(doc, pod); err != nil {
				logger.Errorf("Failed to sync document: error=%v", err)
			}

			stop = true
			return stop
		}

		stop = false
		return stop
	})
	return tracker, nil
}

func (u *DocumentProxyInfoUpdater) Sync(doc SpawnableDocument) error {
	pod, err := u.getPod(doc)

	if err != nil && kerrors.IsNotFound(err) {

		return u.Reset(doc)

	} else if err != nil {

		u.Reset(doc)
		return err
	}

	return u.SyncWithPod(doc, pod)
}

func (u *DocumentProxyInfoUpdater) Reset(doc SpawnableDocument) (err error) {
	var q = bson.M{"_id": doc.GetID()}
	var m = bson.M{
		"$set": bson.M{
			"backend.connected": false,
		},
		"$unset": bson.M{
			"backend.host": nil,
			"backend.port": nil,
			"pod":          nil,
		},
	}
	err = u.Session.C(u.CollectionName).Update(q, m)
	u.emit(doc, doc.NewUpdateEvent(bson.M{
		"backend.connected": false,
		"backend.host":      nil,
		"backend.port":      nil,
		"pod":               nil,
	}))
	return err
}

// SyncWith updates the given document's "backend" and "pod" field by the given
// pod object.
func (p *DocumentProxyInfoUpdater) SyncWithPod(doc SpawnableDocument, pod *v1.Pod) (err error) {

	port, ok := podutil.FindContainerPort(pod, p.PortName)
	if !ok {
		return ErrPortNotFound
	}

	backend := NewProxyBackendFromPod(pod, port)

	q := bson.M{"_id": doc.GetID()}
	m := bson.M{
		"$set": bson.M{
			"backend": backend,
			"pod":     NewPodInfo(pod),
		},
	}

	if err = p.Session.C(p.CollectionName).Update(q, m); err != nil {
		return err
	}

	cache := NewProxyCache(p.Redis, 60*10)
	cache.SetAddress(doc.GetID().Hex(), backend.Addr())

	p.emit(doc, doc.NewUpdateEvent(bson.M{
		"backend":           backend,
		"backend.connected": pod.Status.PodIP != "",
		"pod.phase":         pod.Status.Phase,
		"pod.message":       pod.Status.Message,
		"pod.reason":        pod.Status.Reason,
		"pod.startTime":     pod.Status.StartTime,
	}))
	return nil
}

func (p *DocumentProxyInfoUpdater) emit(doc SpawnableDocument, e *event.RecordEvent) {
	go p.Redis.PublishAndSetJSON(doc.Topic(), e)
}

// NewProxyBackendFromPod creates the proxy backend struct from the pod object.
func NewProxyBackendFromPod(pod *v1.Pod, port int32) *entity.ProxyBackend {
	return &entity.ProxyBackend{
		Connected: pod.Status.PodIP != "",
		Host:      pod.Status.PodIP,
		Port:      int(port),
	}
}
