package hotconfig

import (
	"context"
	"errors"
	"sync"
	"time"
)

// a fetcher to fetch new version of config
type Fetcher interface {
	Fetch(ctx context.Context) (interface{}, error)
}

// a simple fetcher that calls a func
type FetcherFunc func(ctx context.Context) (interface{}, error)

func (f FetcherFunc) Fetch(ctx context.Context) (interface{}, error) {
	return f(ctx)
}

// a hot-updatable config struct
type Config struct {
	lock          *sync.RWMutex
	config        interface{}
	lastUpdated   time.Time
	lastErrorTime time.Time
	lastError     error
	Fetcher       Fetcher
}

// Get the config
func (c *Config) Config() (interface{}, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.lastUpdated.IsZero() {
		return nil, errors.New("hotconfig: config has not been initialized (error occured during initial fetch)")
	}
	return c.config, nil
}

// Get the config, or return nil if not properly initialized
func (c *Config) ConfigOrNil() interface{} {
	ret, err := c.Config()
	if err != nil {
		return nil
	}
	return ret
}

// Update config to given value
func (c *Config) update(config interface{}) {
	c.lock.Lock()
	c.config = config
	c.lastUpdated = time.Now()
	c.lock.Unlock()
}

// Update the config via Fetcher
func (c *Config) Update(ctx context.Context) error {
	config, err := c.Fetcher.Fetch(ctx)
	if err != nil {
		c.lock.Lock()
		c.lastError = err
		c.lastErrorTime = time.Now()
		c.lock.Unlock()
		return err
	}
	c.update(config)
	return nil
}

// Get the last-update timestamp
func (c *Config) LastUpdated() time.Time {
	c.lock.RLock()
	ret := c.lastUpdated
	c.lock.RUnlock()
	return ret
}

// Periodically update a config with set interval. Launch this in a goroutine.
// Kill it by cancelling context
func (c *Config) StartPeriodicUpdate(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
			c.Update(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// Create new hot-reloadable config
func NewConfig(ctx context.Context, fetcher Fetcher) *Config {
	config := &Config{Fetcher: fetcher, lock: &sync.RWMutex{}}
	config.Update(ctx)
	return config
}
