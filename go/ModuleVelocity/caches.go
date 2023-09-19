package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"sync"
	"time"
)

// Markets cache
type cachedMarket struct {
	CreatorID string
	URL       string
	Pool      ManifoldApi.MarketPool
}

var marketsCache = CreateGenericCache("markets-v2", func(marketId string) cachedMarket {
	var apiMarket = ManifoldApi.GetMarket(marketId)
	return cachedMarket{
		CreatorID: apiMarket.CreatorID,
		URL:       apiMarket.URL,
		Pool:      apiMarket.Pool,
	}
}, time.Hour*24*14, time.Hour*24*14)

// Users cache
type cachedUser struct {
	CreatedTime         int64
	ProfitCachedAllTime float64
	SkillEstimate       float64 // [0-1], our own formula that estimates skill
}

var usersCache = CreateGenericCache("users-v3", func(userId string) cachedUser {
	var apiUser = ManifoldApi.GetUser(userId)

	var skillEstimate = 0.5 + utils.MapNumber(apiUser.ProfitCached.AllTime, -2_000, 20_000, -0.1, 0.3) + utils.MapNumber(apiUser.ProfitCached.Monthly, -2_000, 10_000, -0.1, 0.2)

	return cachedUser{
		CreatedTime:         apiUser.CreatedTime,
		ProfitCachedAllTime: apiUser.ProfitCached.AllTime,
		SkillEstimate:       skillEstimate,
	}
}, time.Hour*24*5, time.Minute*15)

// My market position cache
type cachedMarketPosition struct {
	Invested     float64
	HasYesShares bool
}

var myMarketPositionCache = CreateGenericCache("myMarketPosition-v2", func(marketId string) cachedMarketPosition {
	var apiMarketPosition = ManifoldApi.GetMarketPositionForUser(marketId, myUserId)
	if apiMarketPosition == nil {
		return cachedMarketPosition{
			Invested:     0,
			HasYesShares: true,
		}
	}

	return cachedMarketPosition{
		Invested:     apiMarketPosition.Invested,
		HasYesShares: apiMarketPosition.HasYesShares,
	}
}, time.Hour*24*5, time.Hour)

// Market velocity cache
var marketVelocityCache = CreateGenericCache("marketVelocity-v1", func(marketId string) bool {
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
	return betsInLast24Hours >= 6 && len(marketPositions) >= 4
}, time.Hour*2, time.Minute)
