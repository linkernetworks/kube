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

func (c *ProxyCache) setCacheAddress(conn *redis.Connection, cacheKey string, address string) error {
	if _, err := conn.SetWithExpire(cacheKey, address, c.ExpirySeconds); err != nil {
		return fmt.Errorf("Failed to update proxy cache: %v", err)
	}
	return nil
}

func (c *ProxyCache) SetAddress(key string, address string) error {
	conn := c.Redis.GetConnection()
	cacheKey := c.Prefix + key
	return c.setCacheAddress(conn, cacheKey, address)
}

func (c *ProxyCache) RemoveAddress(docID string) error {
	cacheKey := c.Prefix + docID
	conn := c.Redis.GetConnection()
	defer conn.Close()
	_, err := conn.Delete(cacheKey)
	return err
}

// GetAddress uses the redis connection to get the address
func (c *ProxyCache) GetAddress(docID string, fetch AddressFetcher) (address string, err error) {
	// Get the document and its pod info cache from redis
	cacheKey := c.Prefix + docID
	conn := c.Redis.GetConnection()
	defer conn.Close()
	address, err = conn.GetString(cacheKey)

	if err == redigo.ErrNil {
		retaddr, err := fetch()
		if err != nil {
			return retaddr, err
		}
		address = retaddr
		if err := c.setCacheAddress(conn, cacheKey, retaddr); err != nil {
			return retaddr, err
		}
		return retaddr, nil
	} else if err != nil {
		return "", err
	}

	return address, err
}
