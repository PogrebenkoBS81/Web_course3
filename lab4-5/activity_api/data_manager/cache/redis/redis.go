package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const serviceName = "Redis"

// Redis - redis manager
type Redis struct {
	Addr     string
	Password string
	DB       int

	redis  *redis.Client
	ctx    context.Context
	mtx    sync.RWMutex
	logger logrus.FieldLogger
}

// NewRedisManager - returns new redis manager instance.
func NewRedisManager(addr, password string, db int, ctx context.Context, logger logrus.FieldLogger) *Redis {
	return &Redis{
		Addr:     addr,
		Password: password,
		DB:       db,
		ctx:      ctx,
		logger:   logger.WithField("module", "Redis"),
	}
}

// Restart - restarts Redis instance, can be user when service is down.
func (r *Redis) Restart() (err error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	entry := r.logger.WithField("func", "Restart")
	entry.Info("Restarting Redis connection...")

	if err = r.close(); err != nil {
		r.logger.Errorf("close redis connection error")
	}

	if err = r.open(); err != nil {
		r.logger.Errorf("open redis connection error")
	}

	return
}

// Open - returns new redis manager instance.
func (r *Redis) Open() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.open()
}

// open - open helper. Will be user in Restart and in Open.
func (r *Redis) open() error {
	entry := r.logger.WithField("func", "open")
	entry.Info("Opening new Redis connection...")

	r.redis = redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})

	r.logger.Info("Pinging new Redis connection...")

	if _, err := r.redis.Ping(r.ctx).Result(); err != nil {
		return fmt.Errorf("redis Ping(r.ctx): %w", err)
	}

	return nil
}

// Close - closes redis connection and sets it to nil.
func (r *Redis) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.close()
}

// close - close helper. Will be user in Restart and in Close.
func (r *Redis) close() error {
	r.logger.WithField("func", "close").Info("Closing Redis connection...")

	if r.redis == nil {
		return errors.New("redis connection doesn't exist")
	}

	if err := r.redis.Close(); err != nil {
		return fmt.Errorf("redis r.redis.Close() error: %w", err)
	}

	r.redis = nil

	return nil
}

// OK - checks redis connection.
func (r *Redis) OK() error {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	r.logger.WithField("func", "OK").Debug("Checking Redis connection...")

	// TODO: if OK() will be used not only in pinger - create OK flag to reduce ping overhead.
	if _, err := r.redis.Ping(r.ctx).Result(); err != nil {
		return fmt.Errorf("redis r.redis.Ping(): %w", err)
	}

	return nil
}

// Describe - returns Redis name, so it's possible to identify service behind interface.
func (r *Redis) Describe() string {
	r.logger.WithField("func", "Describe").Debug("Returning Redis description...")

	return serviceName
}

// Set - sets given key:value pair to Redis.
func (r *Redis) Set(key string, value interface{}, expiration time.Duration) error {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	r.logger.WithField("func", "Set").Debug("Setting new key to Redis:", key)
	rtCreated, err := r.redis.Set(r.ctx, key, value, expiration).Result()

	if err != nil {
		return fmt.Errorf("redis Set(): %w", err)
	}

	if rtCreated == "0" {
		return errors.New("redis Set(): no record inserted")
	}

	return nil
}

// Get - returns value by key from redis.
func (r *Redis) Get(key string) (string, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	r.logger.WithField("func", "Get").Debug("Getting key from Redis:", key)
	data, err := r.redis.Get(r.ctx, key).Result()

	if err != nil {
		return "", fmt.Errorf("redis r.redis.Get(): %w", err)
	}

	return data, nil
}

// Del - deletes value by key from redis.
func (r *Redis) Del(keys ...string) error {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	r.logger.WithField("func", "Del").Debug("Deleting keys from Redis:", keys)
	deletedRt, err := r.redis.Del(r.ctx, keys...).Result()

	if err != nil {
		return fmt.Errorf("redis r.redis.Del(): %w", err)
	}

	if int(deletedRt) != len(keys) {
		return fmt.Errorf("redis Del(): error deleting, expected deletions %d, got %d", len(keys), deletedRt)
	}

	return nil
}
