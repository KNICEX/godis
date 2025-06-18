package dict

type Consumer func(key string, val any) (continues bool)

type Dict interface {
	Get(key string) (val any, exist bool)
	Len() int
	Put(key string, val any) int
	PutIfAbsent(key string, val any) int
	PutIfExists(key string, val any) int
	Remove(key string) (val any, result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
	DictScan(cursor int, count int, pattern string) ([][]byte, int)
}
