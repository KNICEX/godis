package dict

import (
	"godis/pkg/wildcard"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/samber/lo"
)

type ConcurrentDict struct {
	table     []*shard
	count     int64
	tableSize int
}

type shard struct {
	m  map[string]any
	mu sync.RWMutex
}

func computeCapacity(param int) int {
	if param <= 16 {
		return 16
	}

	n := param - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return math.MaxInt32
	} else {
		return n + 1
	}
}

func NewConcurrentDict(shardCount int) *ConcurrentDict {
	shardCount = computeCapacity(shardCount)
	table := make([]*shard, shardCount)
	for i := 0; i < shardCount; i++ {
		table[i] = &shard{
			m: make(map[string]any),
		}
	}
	return &ConcurrentDict{
		table:     table,
		count:     0,
		tableSize: shardCount,
	}
}

func (dict *ConcurrentDict) fnv32(key string) (uint32, error) {
	h := fnv.New32()
	_, err := h.Write([]byte(key))
	if err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

func (dict *ConcurrentDict) Len() int {
	return int(atomic.LoadInt64(&dict.count))
}

func (dict *ConcurrentDict) Get(key string) (val any, exist bool) {
	shard := dict.getShard(key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	val, exist = shard.m[key]
	return
}

func (dict *ConcurrentDict) GetWithoutLock(key string) (val any, exist bool) {
	shard := dict.getShard(key)

	val, exist = shard.m[key]
	return
}

func (dict *ConcurrentDict) Put(key string, val any) int {
	shard := dict.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exist := shard.m[key]; !exist {
		shard.m[key] = val
		return 0
	}
	dict.addCount()
	shard.m[key] = val
	return 1
}

func (dict *ConcurrentDict) PutWithoutLock(key string, val any) int {
	shard := dict.getShard(key)

	if _, exist := shard.m[key]; !exist {
		shard.m[key] = val
		return 0
	}
	dict.addCount()
	shard.m[key] = val
	return 1
}

func (dict *ConcurrentDict) PutIfAbsent(key string, val any) int {
	shard := dict.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exist := shard.m[key]; exist {
		return 0
	}
	shard.m[key] = val
	dict.addCount()
	return 1
}

func (dict *ConcurrentDict) PutIfAbsentWithoutLock(key string, val any) int {
	shard := dict.getShard(key)

	if _, exist := shard.m[key]; exist {
		return 0
	}
	shard.m[key] = val
	dict.addCount()
	return 1
}

func (dict *ConcurrentDict) PutIfExists(key string, val any) int {
	shard := dict.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exist := shard.m[key]; !exist {
		return 0
	}
	shard.m[key] = val
	return 1
}

func (dict *ConcurrentDict) PutIfExistsWithoutLock(key string, val any) int {
	shard := dict.getShard(key)
	if _, exist := shard.m[key]; !exist {
		return 0
	}
	shard.m[key] = val
	return 1
}

func (dict *ConcurrentDict) Remove(key string) (any, int) {
	shard := dict.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	val, exist := shard.m[key]
	if exist {
		delete(shard.m, key)
		dict.decreaseCount()
		return val, 1
	}
	return nil, 0
}

func (dict *ConcurrentDict) RemoveWithoutLock(key string) (val any, exist bool) {
	shard := dict.getShard(key)
	val, exist = shard.m[key]
	if exist {
		delete(shard.m, key)
		dict.decreaseCount()
	}
	return
}

func (dict *ConcurrentDict) ForEach(consumer Consumer) {
	for _, shard := range dict.table {
		shard.mu.RLock()
		f := func() bool {
			defer shard.mu.RUnlock()
			for key, val := range shard.m {
				if !consumer(key, val) {
					return false
				}
			}
			return true
		}
		if !f() {
			return
		}
	}
}

func (dict *ConcurrentDict) Keys() []string {
	keys := make([]string, 0, dict.Len())
	dict.ForEach(func(key string, _ any) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

func (shard *shard) RandomKey() string {
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	for key := range shard.m {
		return key
	}
	return "" // Return empty string if no key
}

func (dict *ConcurrentDict) RandomKeys(limit int) []string {
	size := dict.Len()
	if limit > size {
		return dict.Keys()
	}

	result := make([]string, 0, limit)
	for i := 0; i < limit; {
		s := dict.table[rand.Intn(dict.tableSize)]
		key := s.RandomKey()
		if key != "" {
			result = append(result, key)
			i++
		}
	}
	return result
}

func (dict *ConcurrentDict) RandomDistinctKeys(limit int) []string {
	size := dict.Len()
	if limit > size {
		return dict.Keys()
	}

	result := make(map[string]struct{}, limit)
	for len(result) < limit {
		s := dict.table[rand.Intn(dict.tableSize)]
		key := s.RandomKey()
		if key != "" {
			if _, exists := result[key]; !exists {
				result[key] = struct{}{}
			}
		}
	}

	arr := make([]string, 0, len(result))
	for key := range result {
		arr = append(arr, key)
	}
	return arr
}

func (dict *ConcurrentDict) Clear() {
	*dict = *NewConcurrentDict(dict.tableSize)
}

func (dict *ConcurrentDict) toLockIndices(keys []string, reverse bool) []int {
	indexMap := make(map[int]struct{})
	for _, key := range keys {
		index := dict.spreadKey(key)
		indexMap[index] = struct{}{}
	}

	indices := make([]int, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}

	sort.Slice(indices, func(i, j int) bool {
		if reverse {
			return indices[i] > indices[j]
		}
		return indices[i] < indices[j]
	})
	return indices
}

func (dict *ConcurrentDict) RWLocks(writeKeys []string, readKeys []string) {
	wIndices := dict.toLockIndices(writeKeys, false)
	rIndices := dict.toLockIndices(readKeys, false)

	for _, index := range wIndices {
		dict.table[index].mu.Lock()
	}
	for _, index := range rIndices {
		dict.table[index].mu.RLock()
	}
}

func (dict *ConcurrentDict) RWUnlocks(writeKeys []string, readKeys []string) {
	wIndices := dict.toLockIndices(writeKeys, true)
	rIndices := dict.toLockIndices(readKeys, true)

	for _, index := range rIndices {
		dict.table[index].mu.RUnlock()
	}
	for _, index := range wIndices {
		dict.table[index].mu.Unlock()
	}
}

// DictScan scans the dictionary for keys matching the given pattern.
// return 0 if all keys have been scanned,
// return -1 if the pattern is invalid,
// return the next cursor position if there are more keys to scan.
func (dict *ConcurrentDict) DictScan(cursor int, count int, pattern string) ([][]byte, int) {
	// size := dict.Len()
	result := make([][]byte, 0)

	if pattern == "*" {
		return lo.Map(dict.Keys(), func(key string, _ int) []byte {
			return []byte(key)
		}), 0
	}

	exp, err := wildcard.Compile(pattern)
	if err != nil {
		return result, -1
	}

	size := dict.tableSize
	shardIdx := cursor

	for shardIdx < size {
		shard := dict.table[shardIdx]
		shard.mu.RLock()

		if len(result)+len(shard.m) > count && shardIdx > cursor {
			// If we have enough results, return them
			shard.mu.RUnlock()
			return result, shardIdx
		}

		for key := range shard.m {
			if pattern == "*" || exp.Match(key) {
				result = append(result, []byte(key))
			}
		}

		shard.mu.RUnlock()
		shardIdx++
	}

	return result, 0
}

func (dict *ConcurrentDict) getShard(key string) *shard {
	return dict.table[dict.spreadKey(key)]
}

func (dict *ConcurrentDict) spreadKey(key string) int {
	hash, err := dict.fnv32(key)
	if err != nil {
		panic(err)
	}
	return dict.spread(hash)
}

func (dict *ConcurrentDict) spread(hashCode uint32) int {
	return int(hashCode % uint32(dict.tableSize))
}

func (dict *ConcurrentDict) addCount() int64 {
	return atomic.AddInt64(&dict.count, 1)
}

func (dict *ConcurrentDict) decreaseCount() int32 {
	return int32(atomic.AddInt64(&dict.count, -1))
}

func (dict *ConcurrentDict) addCountBy(n int64) int64 {
	return atomic.AddInt64(&dict.count, n)
}
