package lruexp

import (
	"fmt"
	"github.com/hashicorp/golang-lru"
	"math/rand"
	"time"
)

var negativeExpiryErr = fmt.Errorf("can't expire object with negative default expiry")
var nilFuncErr = fmt.Errorf("can't handle nil func")

type entry struct {
	object interface{}
	expAt  time.Time
}

type SyncCache struct {
	expiry    time.Duration
	randomExp time.Duration
	cacheNil  bool
	cache     *lru.ARCCache
}

func NewSyncCache(size int, expiry, randomExp time.Duration, cacheNil bool) (*SyncCache, error) {
	if expiry < 0 {
		return nil, negativeExpiryErr
	}

	cache, err := lru.NewARC(size)
	if err != nil {
		return nil, err
	}

	if randomExp == 0 {
		randomExp = 1 * time.Nanosecond
	}

	return &SyncCache{
		expiry:    expiry,
		randomExp: randomExp,
		cacheNil:  cacheNil,
		cache:     cache,
	}, nil
}

func (c *SyncCache) FetchWithFunc(key string, expiry time.Duration, f func() (interface{}, error)) (interface{}, error) {
	if f == nil {
		return nil, nilFuncErr
	}

	entryItf, found := c.cache.Get(key)
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
	if obj == nil && !c.cacheNil {
		return nil, nil
	}

	_entry := &entry{object: obj}
	if c.expiry == 0 {
		if expiry > 0 {
			_entry.expAt = time.Now().Add(expiry).Add(time.Duration(rand.Int63n(int64(c.randomExp))) * time.Nanosecond)
		}
	} else {
		if expiry > 0 {
			_entry.expAt = time.Now().Add(expiry).Add(time.Duration(rand.Int63n(int64(c.randomExp))) * time.Nanosecond)
		} else if expiry == 0 {
			_entry.expAt = time.Now().Add(c.expiry).Add(time.Duration(rand.Int63n(int64(c.randomExp))) * time.Nanosecond)
		}
	}

	c.cache.Add(key, _entry)
	return obj, nil
}
