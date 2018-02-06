package kudis

type Subscription interface {
	Topic() string
	NumSubscribers() (int, error)

	Start() error
	Stop() error
	IsRunning() bool
}
