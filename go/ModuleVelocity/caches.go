package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

type GenericCache[T interface{}] struct {
	cache      *cache.Cache[string, T]
	getItem    func(id string) T
	expiration time.Duration
	lock       *utils.StringKeyLock
}

func CreateGenericCache[T interface{}](getItem func(id string) T, expiration time.Duration) GenericCache[T] {
	var c = cache.New[string, T]()

	return GenericCache[T]{
		cache:      c,
		getItem:    getItem,
		expiration: expiration,
		lock:       utils.NewStringKeyLock(),
	}

}

func (c *GenericCache[T]) DeleteCache(id string) {
	c.cache.Delete(id)
}

func (c *GenericCache[T]) Get(id string) T {
	c.lock.Lock(id)
	defer c.lock.Unlock(id)

	var cachedMarket, ok = c.cache.Get(id)
	if ok {
		return cachedMarket
	}

	var freshItem = c.getItem(id)

	c.cache.Set(id, freshItem, cache.WithExpiration(c.expiration))

	return freshItem
}

// Create caches to us eone on other files
var marketsCache = CreateGenericCache(ManifoldApi.GetMarket, time.Minute*30)
var betsForMarketCache = CreateGenericCache(getBetsForMarket, time.Minute*5)
var usersCache = CreateGenericCache(ManifoldApi.GetUser, time.Minute*30)
var marketPositionsCache = CreateGenericCache(ManifoldApi.GetMarketPositions, time.Minute*5)

func getBetsForMarket(marketId string) []ManifoldApi.Bet {
	return ManifoldApi.GetBets(marketId, "")
}
