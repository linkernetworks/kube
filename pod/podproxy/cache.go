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

const DefaultProxyCachePrefix = "proxy:cache:address:"

func NewProxyCache(r *redis.Service, expirySeconds int) *ProxyCache {
	return &ProxyCache{DefaultProxyCachePrefix, r, expirySeconds}
}

func (c *ProxyCache) setCacheAddress(conn *redis.Connection, cacheKey string, address string) error {
	if _, err := conn.SetWithExpire(cacheKey, address, c.ExpirySeconds); err != nil {
		return fmt.Errorf("Failed to update proxy cache: %v", err)
	}
	return nil
}

func (c *ProxyCache) SetAddress(deploymentID string, address string) error {
	conn := c.Redis.GetConnection()
	cacheKey := c.Prefix + deploymentID
	return c.setCacheAddress(conn, cacheKey, address)
}

// GetAddress uses the redis connection to get the address
func (c *ProxyCache) GetAddress(deploymentID string, fetch AddressFetcher) (address string, err error) {
	// Get the document and its pod info cache from redis
	cacheKey := c.Prefix + deploymentID
	conn := c.Redis.GetConnection()
	address, err = conn.GetString(cacheKey)

	if err == redigo.ErrNil {
		retaddr, err := fetch()
		if err != nil {
			return "", err
		}
		if err := c.setCacheAddress(conn, cacheKey, address); err != nil {
			return "", err
		}
		return retaddr, nil
	} else if err != nil {
		return "", err
	}

	return address, err
}
