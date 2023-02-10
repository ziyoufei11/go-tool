package runtime

import (
	"sync"
	"time"

	"go_tool/support/logger"
)

type runtimeCache struct {
	key    string
	val    string
	expire time.Time
}

type cache struct {
	data *sync.Map
}

var once sync.Once
var cacheInstance *cache

// Cache 运行时缓存
func Cache() *cache {
	once.Do(func() {
		cacheInstance = &cache{
			data: new(sync.Map),
		}
	})

	return cacheInstance
}

func (a *cache) SetCache(key, value string, expire time.Time) {
	logger.SugarLog.Infof("runtime cache saved %s, expireAt %s", key, expire)

	a.data.Store(key, &runtimeCache{
		key:    key,
		val:    value,
		expire: expire,
	})
}

func (a *cache) GetCache(key string) string {
	cacheData, ok := a.data.Load(key)
	if !ok {
		return ""
	}

	cc := cacheData.(*runtimeCache)
	if time.Now().After(cc.expire) {
		a.data.Delete(key)
		logger.SugarLog.Infof("runtime delete %s", key)
		return ""
	}

	logger.SugarLog.Infof("runtime get %s, %s", key, cc.val)
	return cc.val
}
