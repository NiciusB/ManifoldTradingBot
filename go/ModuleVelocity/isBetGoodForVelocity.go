package modulevelocity

import (
	"math"
	"slices"
	"time"
)

var MIN_PROB_SWING = 0.13

var bannedUserIDs = []string{
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

func isBetGoodForVelocity(bet SupabaseBet) bool {
	if bet.IsAPI {
		// Ignore bots, mainly to prevent infinite loops of one reacting to another
		return false
	}

	if bet.IsRedemption {
		// Ignore redemptions
		return false
	}

	if bet.Amount == 0 {
		// Ignore unfilled limit orders
		return false
	}

	if bet.AnswerID != "undefined" && bet.AnswerID != "" {
		// Ignore non-binary markets, until the API supports betting yes/no on those
		return false
	}

	if bet.ProbAfter > 0.9 || bet.ProbAfter < 0.1 {
		// Ignore extreme probabilities since limit orders do not have enough granularity
		return false
	}

	var probDiff = math.Abs(bet.ProbBefore - bet.ProbAfter)
	if probDiff < MIN_PROB_SWING {
		// Ignore small prob changes
		return false
	}

	if slices.Contains(bannedUserIDs, bet.UserID) {
		// Ignore banned user ids
		return false
	}

	var cachedMarket = marketsCache.Get(bet.ContractID)
	if cachedMarket.CreatorID == bet.UserID {
		// Ignore bets by market creator
		return false
	}

	var myPosition = myMarketPositionCache.Get(bet.ContractID)
	if myPosition != nil && myPosition.Invested > 200 {
		// Ignore markets where I am too invested. This could be increased in the future to allow larger positions
		return false
	}

	var cachedUser = usersCache.Get(bet.UserID)
	var isNewAccount = cachedUser.CreatedTime > time.Now().UnixMilli()-1000*60*60*24*3
	if isNewAccount && cachedUser.ProfitCachedAllTime > 1000 {
		// Ignore new accounts with large profits
		return false
	}

	var marketVelocity = marketVelocityCache.Get(bet.ContractID)
	if !marketVelocity {
		// Ignore markets with low volatility. This check could be improved in the future
		return false
	}

	// Return from variable to prevent Go complaining about previous if being redundant
	var returnValue = true
	return returnValue
}
