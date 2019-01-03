package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/subiz/partitioner/worker"
	"hash/fnv"
	"sync"
)

// Cache is a thread-safe fixed size LRU cache.
// This cache handle cache consistency when worker's partitions is changed by
// deleting all key of other partitions
type Cache struct {
	*sync.Mutex
	cache *lru.Cache
}

// global hashing util, used to hash key to partition number
var ghash = fnv.New32a()

// NewCache create new cache with given size
// size is the maximum number of item in cache
func NewCache(size int, w *worker.Worker) *Cache {
	me := &Cache{Mutex: &sync.Mutex{}}
	if size <= 0 {
		size = 10000
	}
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}
	me.cache = cache
	w.OnUpdate(func(pars []int32) { // called when partition is changed
		if len(pars) == 0 {
			me.cache.Purge()
			return
		}
		// remove all key which is not in our partitions
		me.Lock()
		defer me.Unlock()
		keys := me.cache.Keys()
		mpars := make(map[uint32]bool)
		for _, p := range pars {
			mpars[uint32(p)] = true
		}

		for _, k := range keys {
			ghash.Write(k.([]byte))
			parnumber := ghash.Sum32() % uint32(len(mpars))
			ghash.Reset()
			if mpars[parnumber] == false {
				me.cache.Remove(k)
			}
		}
	})
	return me
}

// Get looks up a key's value from the cache
func (me *Cache) Get(key string) (value interface{}, ok bool) {
	me.Lock()
	defer me.Unlock()
	return me.cache.Get([]byte(key))
}

// Set adds a key-value pair to the cache
func (me *Cache) Set(key string, value interface{}) {
	me.Lock()
	defer me.Unlock()
	me.cache.Add([]byte(key), value)
}

// Delete removes the provided key from the cache
func (me *Cache) Delete(key string) {
	me.Lock()
	defer me.Unlock()
	me.cache.Remove([]byte(key))
}
