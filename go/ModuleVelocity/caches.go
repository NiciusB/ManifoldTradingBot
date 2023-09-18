package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"sync"
	"time"
)

// Markets cache
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
}, time.Hour*98)

// Users cache
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
}, time.Hour*48)

// My market position cache
var myMarketPositionCache = CreateGenericCache(func(marketId string) *ManifoldApi.MarketPosition {
	return ManifoldApi.GetMarketPositionForUser(marketId, myUserId)
}, time.Hour*92)

// Market velocity cache
var marketVelocityCache = CreateGenericCache(func(marketId string) bool {
	var marketPositions []ManifoldApi.MarketPosition
	var betsForMarket []ManifoldApi.Bet
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		marketPositions = ManifoldApi.GetMarketPositions(marketId)
		wg.Done()
	}()
	go func() {
		betsForMarket = ManifoldApi.GetAllBetsForMarket(marketId)
		wg.Done()
	}()
	wg.Wait()

	var betsInLast24Hours = 0
	for _, marketBet := range betsForMarket {
		if marketBet.CreatedTime > time.Now().UnixMilli()-1000*60*60*24 {
			betsInLast24Hours++
		}
	}

	// Returns true if the market has enough velocity for betting
	return betsInLast24Hours >= 3 && len(marketPositions) >= 4
}, time.Minute*15)
