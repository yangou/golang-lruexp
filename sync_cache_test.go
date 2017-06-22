package lruexp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = BeforeSuite(func() {})

var _ = AfterSuite(func() {})

var _ = Describe("SyncCache", func() {
	BeforeEach(func() {})

	It("evicts the expired object", func() {
		cache, err := NewSyncCache(10, 5*time.Second, 100*time.Millisecond, false)
		Ω(err).Should(BeNil())

		one, err := cache.FetchWithFunc("test1", 1*time.Second, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := cache.FetchWithFunc("test1", 5*time.Second, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(1))

		time.Sleep(2 * time.Second) // make sure `one` expires

		// using default expiry - 5 seconds from line:24
		two, err = cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		three, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 3, nil
		})
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(2))
	})

	It("evicts when reaches limit of cache", func() {
		cache, err := NewSyncCache(1, 5*time.Second, 100*time.Millisecond, false)
		Ω(err).Should(BeNil())

		one, err := cache.FetchWithFunc("test1", 10*time.Second, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := cache.FetchWithFunc("test2", 10*time.Second, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		three, err := cache.FetchWithFunc("test1", 10*time.Second, func() (interface{}, error) {
			return 3, nil
		})
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(3))
	})

	It("doesn't evict object if set as non-expirable", func() {
		cache, err := NewSyncCache(10, 0, 100*time.Millisecond, false)
		Ω(err).Should(BeNil())

		cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 1, nil
		})

		cache.FetchWithFunc("test2", -1, func() (interface{}, error) {
			return 2, nil
		})

		time.Sleep(1 * time.Second)

		one, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 3, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := cache.FetchWithFunc("test2", 0, func() (interface{}, error) {
			return 4, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		cache, err = NewSyncCache(10, 1*time.Second, 100*time.Millisecond, false)
		Ω(err).Should(BeNil())

		cache.FetchWithFunc("test3", -1, func() (interface{}, error) {
			return 3, nil
		})

		time.Sleep(2 * time.Second)

		three, err := cache.FetchWithFunc("test3", 0, func() (interface{}, error) {
			return 5, nil
		})
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(3))
	})

	It("caches nil object if configured", func() {
		cache, err := NewSyncCache(10, 1*time.Second, 100*time.Millisecond, true)
		Ω(err).Should(BeNil())

		cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return nil, nil
		})

		one, err := cache.FetchWithFunc("test1", 0, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeNil())

		cache, err = NewSyncCache(10, 1*time.Second, 100*time.Millisecond, false)
		Ω(err).Should(BeNil())

		cache.FetchWithFunc("test2", 0, func() (interface{}, error) {
			return nil, nil
		})

		two, err := cache.FetchWithFunc("test2", 0, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))
	})
})
