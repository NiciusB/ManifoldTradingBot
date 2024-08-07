package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"math"
	"slices"
	"time"
)

var BANNED_USER_IDS = []string{
	"w1knZ6yBvEhRThYPEYTwlmGv7N33",
	"BhNkw088bMNwIFF2Aq5Gg9NTPzz1",
	"dNgcgrHGn8ZB30hyDARNtbjvGPm1",
	"kzjQhRJ4GINn5umiq2ee1QvaMcE2",
	"P7PV13rynzOHyxm8AiXIN568bmF2",
	"jOl1FMKpFbXkoaDGp2qlakUxAiJ3",
	"MxdyEeVgrFMTDDsPbXwAe9W1CLs2",
	"IEVDP2LTpgMYaka38r1TVZcabWS2",
	"prSlKwvKkRfHCY43txO4pG1sFMT2",
	"XebdFvo6vqO5WGXTsWYVdSH3WNc2",
	"Y8xXwCCYe3cBCW5XeU8MxykuPAY2",
	"ymezf2YMJ9aaILxT95uWJj7gnx83",
	"Y96HJoD5tQaPgbKi5JEt5JuQJLN2",
	"ffwIBb255DhSsJRh3VWZ4RY2pxz2",
	"wjbOTRRJ7Ee5mjSMMYrtwoWuiCp2",
	"EFzCw6YhqTYCJpeWHUG6p9JsDy02",
	"UN5UGCJRQdfB3eQSnadiAxjkmRp2",
	"9B5QsPTDAAcWOBW8NJNS7YdUjpO2",
	"KIpsyUwgKmO1YXv2EJPXwGxaO533",
	"VI8Htwx9JYeKeT6cUnH66XvBAv73",
	"n820DjHGX9dKsrv0jHIJV8xmDgr2",
	"w07LrYnLg8XDHySwrKxmAYAnLJH2",
	"U7KQfJgJp1fa35k9EXpQCgvmmjh1",
	"rVaQiGT7qCRfAD9QDQQ8SHxvvuu2",
	"wuOtYy52f4Sx4JFfT85LpizVGsx1",
	"I8VZW5hGw9cfIeWs7oQJaNdFwhL2",
	"kydVkcfg7TU4zrrMBRx1Csipwkw2",
	"QQodxPUTIFdQWJiIzVUW2ztF43e2",
	"K2BeNvRj4beTBafzKLRCnxjgRlv1",
	"zgCIqq8AmRUYVu6AdQ9vVEJN8On1",
	"BB5ZIBNqNKddjaZQUnqkFCiDyTs2",
}

func isBetGoodForVelocity(
	bet utils.ManifoldWebsocketBet,
	loadedCaches loadedCachesType,
	betRequest ManifoldApi.PlaceBetRequest,
) bool {
	if bet.IsAPI {
		// Ignore bots, mainly to prevent infinite loops of one reacting to another
		return false
	}

	if bet.IsRedemption {
		// Ignore redemptions, as they are always the opposite of another bet
		return false
	}

	if bet.Amount == 0 || bet.ProbBefore == bet.ProbAfter {
		// Ignore unfilled limit orders
		return false
	}

	if time.UnixMilli(bet.CreatedTime).Before(time.Now().Add(time.Second * -5)) {
		// Ignore old bets. Sometimes supabase gives old bets for some reason
		return false
	}

	for _, fill := range bet.Fills {
		if fill.Timestamp != bet.CreatedTime {
			// Ignore fills of limit orders, other than the initial one
			return false
		}
	}

	if bet.AnswerID != "undefined" && bet.AnswerID != "" {
		// Ignore non-binary markets, until the API supports betting yes/no on those
		return false
	}

	if loadedCaches.betCreatorUser.ProfitCachedAllTime > 1000 && (betRequest.LimitProb >= 0.99 || betRequest.LimitProb <= 0.01) {
		// Ignore extreme values for decent users
		return false
	}
	if loadedCaches.betCreatorUser.ProfitCachedAllTime > 5000 && (betRequest.LimitProb >= 0.97 || betRequest.LimitProb <= 0.03) {
		// Ignore extreme values for good users
		return false
	}

	var smallestProb = math.Min(bet.ProbBefore, bet.ProbAfter)
	var largestProb = math.Max(bet.ProbBefore, bet.ProbAfter)
	if betRequest.LimitProb <= smallestProb || betRequest.LimitProb >= largestProb {
		// We do not have enough granularity on the limit order probabilities: We would bounce even more than the original probs
		// This only works for betting against the latest best, if we wanted to sometimes follow it, we would need to rework this check
		return false
	}

	var betLogOdds = math.Log(bet.ProbAfter / (1 - bet.ProbAfter))
	var betRequestLogOdds = math.Log(betRequest.LimitProb / (1 - betRequest.LimitProb))
	var logOddsDiff = math.Abs(betRequestLogOdds - betLogOdds)                // How much we would change the market log odds
	var poolSize = loadedCaches.market.Pool.NO + loadedCaches.market.Pool.YES // 100 is the current minimum, 1_000 is decently sized, >10_000 is a big market, >100_000 is larger than LK-99
	var poolSizeFactor = math.Min(poolSize, 50_000) / 50_000                  // From 0 to 1, 0 being pool is small, 1 being pool is huge
	var minLogOddsSwing = 0.4 - poolSizeFactor*0.39
	//log.Printf("[isBetGoodForVelocity] %v : ProbBefore %v, ProbAfter %v, limitProb %v, betRequestOutcome: %v, betLogOdds %v, betRequestLogOdds %v, logOddsDiff %v", loadedCaches.market.URL, bet.ProbBefore, bet.ProbAfter, betRequest.LimitProb, betRequest.Outcome, betLogOdds, betRequestLogOdds, logOddsDiff)
	if logOddsDiff < minLogOddsSwing {
		// Ignore small prob changes
		return false
	}

	if slices.Contains(BANNED_USER_IDS, bet.UserID) {
		// Ignore banned user ids
		return false
	}

	if loadedCaches.market.CreatorID == bet.UserID {
		// Ignore bets by market creator
		return false
	}

	if loadedCaches.myPosition.Invested > 60 && ((betRequest.Outcome == "YES" && loadedCaches.myPosition.HasYesShares) || (betRequest.Outcome == "NO" && !loadedCaches.myPosition.HasYesShares)) {
		// Ignore markets where I am too invested on one side. This could be increased in the future to allow larger positions
		return false
	}

	if loadedCaches.marketVelocity < 0.01 {
		// Ignore markets with low volatility
		return false
	}

	// Return from variable to prevent Go complaining about previous if being redundant
	var returnValue = true
	return returnValue
}
