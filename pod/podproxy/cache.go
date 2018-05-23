package podproxy

import (
	"fmt"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/linkernetworks/redis"
)

type ProxyCache struct {
	Prefix        string
	Redis         *redis.Service
	ExpirySeconds int
}

type AddressFetcher func() (address string, err error)

const DefaultPrefix = "podproxy:"

func NewDefaultProxyCache(rds *redis.Service) *ProxyCache {
	return &ProxyCache{
		Prefix: DefaultPrefix,
		// TODO: load from config
		ExpirySeconds: 60 * 10, // 10 minutes
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

func (c *ProxyCache) SetAddress(id string, address string) error {
	var key = id + ":address"
	var conn = c.Redis.GetConnection()
	return c.set(conn, key, address)
}

func (c *ProxyCache) RemoveAddress(id string) error {
	var key = id + ":address"
	var conn = c.Redis.GetConnection()
	defer conn.Close()
	return c.unset(conn, key)
}

// GetAddressWith uses the redis connection to get the address
func (c *ProxyCache) GetAddressWith(id string, fetch AddressFetcher) (address string, err error) {
	var key = id + ":address"

	// Get the document and its pod info cache from redis
	var conn = c.Redis.GetConnection()
	defer conn.Close()
	address, err = c.get(conn, key)

	if err == redigo.ErrNil {
		retaddr, err := fetch()
		if err != nil {
			return retaddr, err
		} else if len(retaddr) == 0 {
			return "", fmt.Errorf("Empty pod IP address")
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
