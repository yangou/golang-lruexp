package lruexp

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("AsyncCache", func() {
	BeforeEach(func() {})

	It("fetches the object if not cached", func() {
		cache, err := NewAsyncCache(10, 5*time.Second, 100*time.Millisecond, false, nil)
		Ω(err).Should(BeNil())

		one, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 1, nil
		}, nil)
		Ω(one).Should(BeEquivalentTo(1))
		Ω(err).Should(BeNil())

		erred := false
		errFunc := func(error) {
			erred = true
		}
		two, err := cache.FetchWithFunc("test2", 0, func() (interface{}, error) {
			return 2, errors.New("test error")
		}, errFunc)
		Ω(two).Should(BeNil())
		Ω(err).Should(BeNil())
		Ω(erred).Should(BeTrue())

		three, err := cache.FetchWithFunc("test3", 0, func() (interface{}, error) {
			return nil, nil
		}, nil)
		Ω(three).Should(BeNil())
		Ω(err).Should(BeNil())

		Ω(cache.cache.Len()).Should(Equal(1))

		one, err = cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 0, nil
		}, nil)
		Ω(one).Should(BeEquivalentTo(1))
		Ω(err).Should(BeNil())
	})

	It("caches nil if configured", func() {
		cache, err := NewAsyncCache(10, 5*time.Second, 100*time.Millisecond, true, nil)
		Ω(err).Should(BeNil())

		one, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return nil, nil
		}, nil)
		Ω(one).Should(BeNil())
		Ω(err).Should(BeNil())

		Ω(cache.cache.Len()).Should(Equal(1))

		one, err = cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 1, nil
		}, nil)
		Ω(one).Should(BeNil())
		Ω(err).Should(BeNil())
	})

	It("evicts the expired object", func() {
		cache, err := NewAsyncCache(10, 5*time.Second, 100*time.Millisecond, false, nil)
		Ω(err).Should(BeNil())

		one, err := cache.FetchWithFunc("test1", 1*time.Second, func() (interface{}, error) {
			return 1, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := cache.FetchWithFunc("test1", 5*time.Second, func() (interface{}, error) {
			return 2, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(1))

		time.Sleep(2 * time.Second) // make sure `one` expires

		// using default expiry - 5 seconds from line:71
		two, err = cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 2, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(1)) // should return the stale object, but this should have triggered update

		time.Sleep(1 * time.Second) // make sure update is done

		three, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 3, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(2))
	})

	It("evicts when reaches limit of cache", func() {
		cache, err := NewAsyncCache(1, 5*time.Second, 100*time.Millisecond, false, nil)
		Ω(err).Should(BeNil())

		one, err := cache.FetchWithFunc("test1", 10*time.Second, func() (interface{}, error) {
			return 1, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := cache.FetchWithFunc("test2", 10*time.Second, func() (interface{}, error) {
			return 2, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		three, err := cache.FetchWithFunc("test1", 10*time.Second, func() (interface{}, error) {
			return 3, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(3))
	})

	It("doesn't evict object if set as non-expirable", func() {
		cache, err := NewAsyncCache(10, 0, 100*time.Millisecond, false, nil)
		Ω(err).Should(BeNil())

		cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 1, nil
		}, nil)

		cache.FetchWithFunc("test2", -1, func() (interface{}, error) {
			return 2, nil
		}, nil)

		time.Sleep(1 * time.Second)

		one, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 3, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := cache.FetchWithFunc("test2", 0, func() (interface{}, error) {
			return 4, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		cache, err = NewAsyncCache(10, 1*time.Second, 100*time.Millisecond, false, nil)
		Ω(err).Should(BeNil())

		cache.FetchWithFunc("test3", -1, func() (interface{}, error) {
			return 3, nil
		}, nil)

		time.Sleep(2 * time.Second)

		three, err := cache.FetchWithFunc("test3", 0, func() (interface{}, error) {
			return 5, nil
		}, nil)
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(3))
	})

})
