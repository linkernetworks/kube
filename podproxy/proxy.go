package podproxy

type ProxyInfoProvider interface {
	Host() string
	Port() string
	BaseURL() string
}
