package cache_mock

import (
	"errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const serviceName = "Cache Mock"

// CacheMock - mock cache service for testing
type CacheMock struct {
	fields map[string]string
	mtx    sync.RWMutex // RWMutex is used to improve performance
	logger logrus.FieldLogger
}

// NewRedisManager - returns new redis manager instance.
func NewCacheMock(logger logrus.FieldLogger) *CacheMock {
	return &CacheMock{
		logger: logger.WithField("module", "CacheMock"),
	}
}

// Restart - restart mock
func (m *CacheMock) Restart() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	entry := m.logger.WithField("func", "Restart")
	entry.Debug("Restarting cache mock...")

	if err := m.close(); err != nil {
		m.logger.Errorf("close redis connection error")
	}

	if err := m.open(); err != nil {
		m.logger.Errorf("open redis connection error")
	}

	return nil
}

// Get - returns value from CacheMock
func (m *CacheMock) Get(key string) (string, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	// Same as Redis checks, don't need it, but let it be
	if m.fields == nil {
		return "", errors.New("mock 'connection' doesn't exist")
	}

	m.logger.WithField("func", "Get").Debugf("Getting value %s from cache mock", key)

	if data, ok := m.fields[key]; ok {
		return data, nil
	}

	return "", errors.New("field doesn't exist")
}

// Set - sets given key:value pair to cache mock.
func (m *CacheMock) Set(key string, value interface{}, _ time.Duration) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// Same as Redis checks, don't need it, but let it be
	if m.fields == nil {
		return errors.New("mock 'connection' doesn't exist")
	}

	data, ok := value.(string)
	if !ok {
		return errors.New("non string key, interface key does not implemented")
	}

	m.logger.WithField("func", "Set").Debugf("Setting value %s by key %s to cache mock", data, key)
	// Redis overrides key, as I know, so I`ll do the same
	m.fields[key] = data

	return nil
}

// Del - deletes value by key from cache mock.
func (m *CacheMock) Del(keys ...string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.logger.WithField("func", "Del").Debugf("Deleting values %v from cache mock", keys)

	// Same as Redis checks, don't need it, but let it be
	if m.fields == nil {
		return errors.New("mock 'connection' doesn't exist")
	}

	for _, key := range keys {
		delete(m.fields, key)
	}

	return nil
}

// Describe - returns cache mock name, so it's possible to identify service behind interface.
func (m *CacheMock) Describe() string {
	m.logger.WithField("func", "Del").Debug("Getting cache mock description")

	return serviceName
}

// Open - 'opens' cache mock connection.
func (m *CacheMock) Open() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return m.open()
}

// open - cache mock Open helper .
func (m *CacheMock) open() error {
	if m.fields != nil {
		return errors.New("'connection' already exists")
	}

	m.fields = make(map[string]string)

	return nil
}

// Close - closes redis connection and sets it to nil.
func (m *CacheMock) Close() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return m.close()
}

// close - cache mock Close helper .
func (m *CacheMock) close() error {
	if m.fields == nil {
		return errors.New("'connection' doesn't exist")
	}

	m.fields = nil

	return nil
}

// OK - check connection mock.
func (m *CacheMock) OK() error {
	if m.fields == nil {
		return errors.New("'connection' doesn't exist")
	}

	return nil
}
