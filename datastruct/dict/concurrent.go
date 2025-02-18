package dict

import "sync"

type ConcurrentDict struct {
	table      []*shard
	count      int
	shardCount int
}

type shard struct {
	m  map[string]any
	mu sync.RWMutex
}

func NewConcurrentDict(shardCount int) *ConcurrentDict {
	if shardCount <= 0 {
		shardCount = 1
	}
	cd := &ConcurrentDict{
		shardCount: shardCount,
	}
	return cd
}
