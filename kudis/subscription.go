package kudis

type Subscription interface {
	Topic() string

	Start() error
	Stop() error
	IsRunning() bool
}
