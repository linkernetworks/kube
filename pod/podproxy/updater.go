package podproxy

import (
	"fmt"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
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
func (u *DocumentProxyInfoUpdater) TrackAndSync(doc SpawnableDocument) (*podtracker.PodTracker, error) {
	podName := doc.DeploymentID()

	podTracker := podtracker.New(u.Clientset, u.Namespace, podName)

	podTracker.Track(func(pod *v1.Pod) bool {
		phase := pod.Status.Phase
		logger.Infof("Tracking notebook pod=%s phase=%s", podName, phase)

		switch phase {
		case "Pending":
			u.SyncWithPod(doc, pod)
			// Check all containers status in a pod. can't be ErrImagePull or ImagePullBackOff
			for _, c := range pod.Status.ContainerStatuses {
				if c.State.Waiting != nil {
					waitingReason := c.State.Waiting.Reason
					if waitingReason == "ErrImagePull" || waitingReason == "ImagePullBackOff" {
						logger.Errorf("Container is waiting. Reason %s\n", waitingReason)

						// stop tracking
						return true
					}
				}
			}

		// Stop the tracker if the status is completion status.
		// Terminating won't be catched
		case "Running", "Failed", "Succeeded", "Unknown", "Terminating":
			u.SyncWithPod(doc, pod)
			return true
		}

		return false
	})
	return podTracker, nil
}

func (u *DocumentProxyInfoUpdater) Sync(doc SpawnableDocument) error {
	pod, err := u.getPod(doc)

	if err != nil && kerrors.IsNotFound(err) {

		return u.Reset(doc, nil)

	} else if err != nil {

		u.Reset(doc, err)
		return err
	}

	return u.SyncWithPod(doc, pod)
}

func (u *DocumentProxyInfoUpdater) Reset(doc SpawnableDocument, kerr error) (err error) {
	var q = bson.M{"_id": doc.GetID()}
	var m = bson.M{
		"$set": bson.M{
			"backend.connected": false,
			"backend.error":     kerr,
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
	backend, err := NewProxyBackendFromPodStatus(pod, p.PortName)
	if err != nil {
		return err
	}

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

	p.emit(doc, doc.NewUpdateEvent(bson.M{
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

// SelectPodContainerPort selects the container port from the given port by the port name
// This method is called by NewProxyBackendFromPodStatus
// TODO: can be moved to kubernetes/pod/util
func SelectPodContainerPort(pod *v1.Pod, portname string) (containerPort int32, found bool) {
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == portname {
				containerPort = port.ContainerPort
				found = true
				return
			}
		}
	}
	return containerPort, found
}

// TODO: can be moved to kubernetes/pod/util
// NewProxyBackendFromPodStatus creates the proxy backend struct from the pod object.
func NewProxyBackendFromPodStatus(pod *v1.Pod, portname string) (*entity.ProxyBackend, error) {
	port, ok := SelectPodContainerPort(pod, portname)
	if !ok {
		return nil, fmt.Errorf("portname %s not found", portname)
	}
	return &entity.ProxyBackend{
		IP:        pod.Status.PodIP,
		Port:      int(port),
		Connected: pod.Status.PodIP != "",
	}, nil
}
