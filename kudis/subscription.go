package kudis

type Subscription interface {
	Topic() string
	NumSubscribers() (int, error)

	Start() error
	IsRunning() bool
}
