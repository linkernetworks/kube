package podproxy

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	redigo "github.com/garyburd/redigo/redis"

	"github.com/stretchr/testify/assert"
)

func TestCacheSet(t *testing.T) {
	cf := config.MustRead("../../../../config/testing.json")
	rds := redis.New(cf.Redis)

	cache := NewDefaultProxyCache(rds)
	conn := rds.GetConnection()
	defer conn.Close()
	err := cache.set(conn, "testing-foo", "10.0.0.1")
	assert.NoError(t, err)
	defer cache.unset(conn, "testing-foo")

	val, err := cache.get(conn, "testing-foo")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", val)
}

func TestCacheUnset(t *testing.T) {
	cf := config.MustRead("../../../../config/testing.json")
	rds := redis.New(cf.Redis)

	cache := NewDefaultProxyCache(rds)
	conn := rds.GetConnection()
	defer conn.Close()
	err := cache.set(conn, "testing-foo", "10.0.0.1")
	assert.NoError(t, err)
	cache.unset(conn, "testing-foo")

	val, err := cache.get(conn, "testing-foo")
	assert.Error(t, err)
	assert.Equal(t, redigo.ErrNil, err)
	assert.Equal(t, "", val)
}

func TestCacheSetAddress(t *testing.T) {
	cf := config.MustRead("../../../../config/testing.json")
	rds := redis.New(cf.Redis)

	cache := NewDefaultProxyCache(rds)
	conn := rds.GetConnection()
	defer conn.Close()

	cache.RemoveAddress("testing-foo-get")

	err := cache.SetAddress("testing-foo-set", "11.22.33.55")
	defer cache.RemoveAddress("testing-foo-get")
	assert.NoError(t, err)

	val, err := cache.GetAddressWith("testing-foo-set", func() (string, error) {
		return "11.22.33.44", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "11.22.33.55", val)
}

func TestCacheGetAddress(t *testing.T) {
	cf := config.MustRead("../../../../config/testing.json")
	rds := redis.New(cf.Redis)

	cache := NewDefaultProxyCache(rds)
	conn := rds.GetConnection()
	defer conn.Close()

	cache.RemoveAddress("testing-foo-get")

	var call = false
	val, err := cache.GetAddressWith("testing-foo-get", func() (string, error) {
		call = true
		return "11.22.33.44", nil
	})
	defer cache.RemoveAddress("testing-foo-get")
	assert.NoError(t, err)
	assert.Equal(t, "11.22.33.44", val)
	assert.True(t, call)

	call = false
	val, err = cache.GetAddressWith("testing-foo-get", func() (string, error) {
		call = true
		return "11.22.33.44", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "11.22.33.44", val)
	assert.False(t, call)
}
