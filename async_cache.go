package lruexp

import (
	"github.com/hashicorp/golang-lru"
	"math/rand"
	"sync"
	"time"
)

type AsyncCache struct {
	expiry    time.Duration
	randomExp time.Duration
	cacheNil  bool
	onError   func(error)

	keys  map[string]chan struct{}
	m     sync.Mutex
	ch    chan *task
	cache *lru.ARCCache
}

type task struct {
	key     string
	expiry  time.Duration
	f       func() (interface{}, error)
	onError func(error)
}

func NewAsyncCache(size int, expiry, randomExp time.Duration, cacheNil bool, onError func(error)) (*AsyncCache, error) {
	if expiry < 0 {
		return nil, negativeExpiryErr
	}

	cache, err := lru.NewARC(size)
	if err != nil {
		return nil, err
	}

	return (&AsyncCache{
		expiry:    expiry,
		randomExp: randomExp,
		cacheNil:  cacheNil,
		onError:   onError,
		keys:      make(map[string]chan struct{}),
		ch:        make(chan *task, size),
		cache:     cache,
	}).run(), nil
}

func (c *AsyncCache) run() *AsyncCache {
	go func() {
		for t := range c.ch {
			go c.doTask(t)
		}
	}()
	return c
}

func (c *AsyncCache) doTask(t *task) {
	defer func() {
		c.m.Lock()
		defer func() { c.m.Unlock() }()
		ch := c.keys[t.key]
		delete(c.keys, t.key)
		close(ch)
	}()

	obj, err := t.f()
	if err != nil {
		onError := c.onError
		if t.onError != nil {
			onError = t.onError
		}
		if onError != nil {
			onError(err)
		}
		return
	}

	if obj == nil && !c.cacheNil {
		return
	}

	_entry := &entry{object: obj}
	if c.expiry == 0 {
		if t.expiry > 0 {
			_entry.expAt = time.Now().Add(t.expiry).Add(time.Duration(rand.Int63n(int64(c.randomExp))) * time.Nanosecond)
		}
	} else {
		if t.expiry > 0 {
			_entry.expAt = time.Now().Add(t.expiry).Add(time.Duration(rand.Int63n(int64(c.randomExp))) * time.Nanosecond)
		} else if t.expiry == 0 {
			_entry.expAt = time.Now().Add(c.expiry).Add(time.Duration(rand.Int63n(int64(c.randomExp))) * time.Nanosecond)
		}
	}

	c.cache.Add(t.key, _entry)
}

func (c *AsyncCache) enqueue(key string, expiry time.Duration, f func() (interface{}, error), onError func(error)) chan struct{} {
	c.m.Lock()
	defer c.m.Unlock()

	if ch, found := c.keys[key]; !found {
		ch := make(chan struct{})
		c.keys[key] = ch
		c.ch <- &task{
			key:     key,
			expiry:  expiry,
			f:       f,
			onError: onError,
		}
		return ch
	} else {
		return ch
	}
}

func (c *AsyncCache) FetchWithFunc(key string, expiry time.Duration, f func() (interface{}, error), onError func(error)) (interface{}, error) {
	if f == nil {
		return nil, nilFuncErr
	}

	entryItf, found := c.cache.Get(key)
	if found {
		_entry := entryItf.(*entry)
		if !_entry.expAt.IsZero() && _entry.expAt.Before(time.Now()) {
			c.enqueue(key, expiry, f, onError)
		}
		return _entry.object, nil
	} else {
		<-c.enqueue(key, expiry, f, onError)

		entryItf, found := c.cache.Get(key)
		if !found {
			return nil, nil
		} else {
			_entry := entryItf.(*entry)
			return _entry.object, nil
		}
	}
}
