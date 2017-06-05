package lruexp

import (
	"fmt"
	"github.com/hashicorp/golang-lru"
	"math/rand"
	"sync"
	"time"
)

var caches = map[string]*cache{}
var lock sync.RWMutex

var negativeExpiryErr = fmt.Errorf("can't expire object with negative default expiry")

type cache struct {
	e        time.Duration
	r        int64
	c        *lru.ARCCache
	cacheNil bool
}

type entry struct {
	object interface{}
	expAt  time.Time
}

func New(name string, size int, expiry time.Duration, randomSec int64, cacheNil bool) error {
	if expiry < 0 {
		return negativeExpiryErr
	}

	c, err := lru.NewARC(size)
	if err != nil {
		return err
	}

	_cache := &cache{
		e:        expiry,
		r:        randomSec,
		c:        c,
		cacheNil: cacheNil,
	}

	lock.Lock()
	defer lock.Unlock()
	caches[name] = _cache
	return nil
}

func FetchWithFunc(name, key string, expiry time.Duration, f func() (interface{}, error)) (interface{}, error) {
	lock.RLock()
	_cache, found := caches[name]
	if !found {
		lock.RUnlock()
		return nil, fmt.Errorf("cache %s is not initialized", name)
	}
	lock.RUnlock()

	entryItf, found := _cache.c.Get(key)
	if found {
		_entry := entryItf.(*entry)
		if _entry.expAt.IsZero() || _entry.expAt.After(time.Now()) {
			return _entry.object, nil
		}
	}

	obj, err := f()
	if err != nil {
		return nil, err
	}
	if obj == nil && !_cache.cacheNil {
		return nil, nil
	}

	_entry := &entry{object: obj}

	if _cache.e == 0 {
		if expiry > 0 {
			_entry.expAt = time.Now().Add(expiry).Add(time.Duration(rand.Int63n(_cache.r)) * time.Millisecond)
		}
	} else {
		if expiry > 0 {
			_entry.expAt = time.Now().Add(expiry).Add(time.Duration(rand.Int63n(_cache.r)) * time.Millisecond)
		} else if expiry == 0 {
			_entry.expAt = time.Now().Add(_cache.e).Add(time.Duration(rand.Int63n(_cache.r)) * time.Millisecond)
		}
	}

	_cache.c.Add(key, _entry)
	return obj, nil
}
