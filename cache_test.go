package lruexp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = BeforeSuite(func() {})

var _ = AfterSuite(func() {})

var _ = Describe("Cache", func() {
	BeforeEach(func() {})

	It("evicts the expired object", func() {
		name := strconv.FormatInt(time.Now().UnixNano(), 10)
		Ω(New(name, 10, 5*time.Second, 100, false)).Should(BeNil())
		one, err := FetchWithFunc(name, "test1", 1*time.Second, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := FetchWithFunc(name, "test1", 5*time.Second, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(1))

		time.Sleep(2 * time.Second) // make sure `one` expires

		// using default expiry - 5 seconds from line:25
		two, err = FetchWithFunc(name, "test1", 0, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		three, err := FetchWithFunc(name, "test1", 0, func() (interface{}, error) {
			return 3, nil
		})
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(2))
	})

	It("evicts when reaches limit of cache", func() {
		name := strconv.FormatInt(time.Now().UnixNano(), 10)
		Ω(New(name, 1, 5*time.Second, 100, false)).Should(BeNil())
		one, err := FetchWithFunc(name, "key1", 10*time.Second, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := FetchWithFunc(name, "key2", 10*time.Second, func() (interface{}, error) {
			return 2, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		three, err := FetchWithFunc(name, "key1", 10*time.Second, func() (interface{}, error) {
			return 3, nil
		})
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(3))
	})

	It("doesn't evict object if set as non-expirable", func() {
		name := strconv.FormatInt(time.Now().UnixNano(), 10)
		Ω(New(name, 10, 0, 100, false)).Should(BeNil())

		FetchWithFunc(name, "key1", 0, func() (interface{}, error) {
			return 1, nil
		})

		FetchWithFunc(name, "key2", -1, func() (interface{}, error) {
			return 2, nil
		})

		time.Sleep(1 * time.Second)

		one, err := FetchWithFunc(name, "key1", 0, func() (interface{}, error) {
			return 3, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))

		two, err := FetchWithFunc(name, "key2", 0, func() (interface{}, error) {
			return 4, nil
		})
		Ω(err).Should(BeNil())
		Ω(two).Should(BeEquivalentTo(2))

		name = strconv.FormatInt(time.Now().UnixNano(), 10)
		Ω(New(name, 10, 1*time.Millisecond, 100, false)).Should(BeNil())

		FetchWithFunc(name, "key3", -1, func() (interface{}, error) {
			return 3, nil
		})

		time.Sleep(1 * time.Second)

		three, err := FetchWithFunc(name, "key3", 0, func() (interface{}, error) {
			return 5, nil
		})
		Ω(err).Should(BeNil())
		Ω(three).Should(BeEquivalentTo(3))
	})

	It("caches nil object if configured", func() {
		name := strconv.FormatInt(time.Now().UnixNano(), 10)
		Ω(New(name, 10, 1*time.Second, 100, true)).Should(BeNil())

		FetchWithFunc(name, "name1", 0, func() (interface{}, error) {
			return nil, nil
		})

		one, err := FetchWithFunc(name, "name1", 0, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeNil())

		name = strconv.FormatInt(time.Now().UnixNano(), 10)
		Ω(New(name, 10, 1*time.Second, 100, false)).Should(BeNil())

		FetchWithFunc(name, "name2", 0, func() (interface{}, error) {
			return nil, nil
		})

		one, err = FetchWithFunc(name, "name2", 0, func() (interface{}, error) {
			return 1, nil
		})
		Ω(err).Should(BeNil())
		Ω(one).Should(BeEquivalentTo(1))
	})
})
