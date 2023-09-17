package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"time"
)

// Create caches to use on other files
type cachedMarket struct {
	CreatorID string
	URL       string
}

var marketsCache = CreateGenericCache(func(marketId string) cachedMarket {
	var apiMarket = ManifoldApi.GetMarket(marketId)
	return cachedMarket{
		CreatorID: apiMarket.CreatorID,
		URL:       apiMarket.URL,
	}
}, time.Hour*24)

var marketPositionsCache = CreateGenericCache(ManifoldApi.GetMarketPositions, time.Minute*30)

var betsForMarketCache = CreateGenericCache(func(marketId string) []ManifoldApi.Bet {
	return ManifoldApi.GetAllBetsForMarket(marketId)
}, time.Minute*15)

type cachedUser struct {
	CreatedTime         int64
	ProfitCachedAllTime float64
}

var usersCache = CreateGenericCache(func(userId string) cachedUser {
	var apiUser = ManifoldApi.GetUser(userId)
	return cachedUser{
		CreatedTime:         apiUser.CreatedTime,
		ProfitCachedAllTime: apiUser.ProfitCached.AllTime,
	}
}, time.Hour*8)
