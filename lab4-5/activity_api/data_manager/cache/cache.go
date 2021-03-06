// Can't create common "core" here like in SQL,
// due to cache client doesn't have common connection interface
// (and i (cant/don't want to thinks how to) create it)
package cache

import (
	"activity_api/data_manager/cache/cache_mock"
	"activity_api/data_manager/cache/redis"
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	ICacheMock = iota
	Redis      = iota
	// Memcache
	// ...
)

// ICacheManager - cache manager interface fr AAService.
type ICacheManager interface {
	Get(key string) (string, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Del(keys ...string) error
	Describe() string

	Restart() error
	Open() error
	Close() error
	OK() error
}

// ICacheConfig - common config for cache services.
type ICacheConfig struct {
	Address  string
	Password string
	DB       int
}

// NewCacheManager - returns new cache manager.
func NewCacheManager(
	cacheType int,
	cacheConfig *ICacheConfig,
	ctx context.Context,
	logger logrus.FieldLogger,
) ICacheManager {
	switch cacheType {
	case ICacheMock:
		return cache_mock.NewCacheMock(logger)
	case Redis:
		return redis.NewRedisManager(cacheConfig.Address, cacheConfig.Password, cacheConfig.DB, ctx, logger)
	default:
		logger.WithField("func", "NewCacheManager").
			Warnf("Unsupported cacheType: %d, using default: Redis", Redis)

		return redis.NewRedisManager(cacheConfig.Address, cacheConfig.Password, cacheConfig.DB, ctx, logger)
	}
}
