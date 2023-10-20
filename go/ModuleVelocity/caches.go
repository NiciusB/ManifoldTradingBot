package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
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

	// [0-1]
	var skillEstimate = 0.5 +
		utils.MapNumber(apiUser.ProfitCached.AllTime, -5_000, 40_000, -0.1, 0.3) +
		utils.MapNumber(apiUser.ProfitCached.Monthly, -2_000, 10_000, -0.1, 0.2) +
		utils.MapNumber(time.Since(time.UnixMilli(apiUser.CreatedTime)).Hours(), 0, 24*30, -0.2, 0)

	return cachedUser{
		CreatedTime:         apiUser.CreatedTime,
		ProfitCachedAllTime: apiUser.ProfitCached.AllTime,
		SkillEstimate:       skillEstimate,
	}
}, time.Hour*24*5, time.Minute*30)

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
var marketVelocityCache = CreateGenericCache("marketVelocity-v3", func(marketId string) float64 {
	var betsForMarket = ManifoldApi.GetAllBetsForMarket(marketId, time.Now().UnixMilli()-1000*60*60*24*31)

	var uniqueBettorsInLastMonth = mapset.NewSet[string]()
	var uniqueBettorsInLastWeek = mapset.NewSet[string]()
	var uniqueBettorsInLastDay = mapset.NewSet[string]()
	for _, marketBet := range betsForMarket {
		if marketBet.CreatedTime > time.Now().UnixMilli()-1000*60*60*24 {
			uniqueBettorsInLastDay.Add(marketBet.UserID)
		}
		if marketBet.CreatedTime > time.Now().UnixMilli()-1000*60*60*24*7 {
			uniqueBettorsInLastWeek.Add(marketBet.UserID)
		}

		// No need to check for CreatedTime since the GetAllBetsForMarket already has that limit
		uniqueBettorsInLastMonth.Add(marketBet.UserID)
	}

	// Minimum requirements
	if uniqueBettorsInLastMonth.Cardinality() < 4 || uniqueBettorsInLastWeek.Cardinality() < 4 {
		return 0
	}

	// [0-1], score for how much the market moves
	return 0 +
		utils.MapNumber(float64(uniqueBettorsInLastDay.Cardinality()), 0, 500, 0, 0.2) +
		utils.MapNumber(float64(uniqueBettorsInLastWeek.Cardinality()), 0, 5000, 0, 0.4) +
		utils.MapNumber(float64(uniqueBettorsInLastMonth.Cardinality()), 0, 400, 0, 0.4)
}, time.Minute*30, time.Minute*2)
