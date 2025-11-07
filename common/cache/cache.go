package cache

import "github.com/puzpuzpuz/xsync/v4"

/*
   Created by zyx
   Date Time: 2025/9/18
   File: cache.go
*/

var cache = xsync.NewMap[string, any]()

type Cache struct{}
type ICache interface {
	Set(key string, value any)
	Get(key string) any
	Del(key string)
	Has(key string) bool
}

func Service() ICache {
	return &Cache{}
}

func (c *Cache) Set(key string, value any) {
	cache.Store(key, value)
}

func (c *Cache) Get(key string) any {
	if ret, ok := cache.Load(key); ok {
		return ret
	}
	return nil
}

func (c *Cache) Del(key string) {
	cache.Delete(key)
}

func (c *Cache) Has(key string) bool {
	_, ok := cache.Load(key)
	return ok
}
