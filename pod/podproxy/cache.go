package podproxy

import (
	"fmt"

	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	redigo "github.com/garyburd/redigo/redis"
)

type ProxyCache struct {
	Prefix        string
	Redis         *redis.Service
	ExpirySeconds int
}

type AddressFetcher func() (address string, err error)

func NewDefaultProxyCache(rds *redis.Service) *ProxyCache {
	return &ProxyCache{
		Prefix:        "podproxy:",
		ExpirySeconds: 60 * 10,
		Redis:         rds,
	}
}

func (c *ProxyCache) set(conn *redis.Connection, key string, val string) error {
	if _, err := conn.SetWithExpire(c.Prefix+key, val, c.ExpirySeconds); err != nil {
		return fmt.Errorf("Failed to update proxy cache: %v", err)
	}
	return nil
}

func (c *ProxyCache) unset(conn *redis.Connection, key string) error {
	_, err := conn.Delete(c.Prefix + key)
	return err
}

func (c *ProxyCache) get(conn *redis.Connection, key string) (val string, err error) {
	val, err = conn.GetString(c.Prefix + key)
	return val, err
}

func (c *ProxyCache) SetAddress(key string, address string) error {
	var conn = c.Redis.GetConnection()
	return c.set(conn, key, address)
}

func (c *ProxyCache) RemoveAddress(key string) error {
	var conn = c.Redis.GetConnection()
	defer conn.Close()
	return c.unset(conn, key)
}

// GetAddress uses the redis connection to get the address
func (c *ProxyCache) GetAddress(key string, fetch AddressFetcher) (address string, err error) {
	// Get the document and its pod info cache from redis
	var conn = c.Redis.GetConnection()
	defer conn.Close()
	address, err = c.get(conn, key)

	if err == redigo.ErrNil {
		retaddr, err := fetch()
		if err != nil {
			return retaddr, err
		}
		address = retaddr

		if err := c.set(conn, key, retaddr); err != nil {
			return retaddr, err
		}
		return retaddr, nil
	} else if err != nil {
		return "", err
	}

	return address, err
}
