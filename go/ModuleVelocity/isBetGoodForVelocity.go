package modulevelocity

import (
	"log"
	"math"
	"slices"
	"sync"
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

func isBetGoodForVelocity(bet SupabaseBet) bool {
	var cachedMarket *cachedMarket
	var myPosition *cachedMarketPosition
	var cachedUser *cachedUser
	var marketVelocity *bool
	var wg sync.WaitGroup

	// Load in advance all needed data, even for obviously not needed markets, since it warms up the cache
	wg.Add(4)
	go func() {
		var err error
		cachedMarket, err = marketsCache.Get(bet.ContractID)
		if err != nil {
			log.Fatalln(err)
		}
		wg.Done()
	}()
	go func() {
		var err error
		myPosition, err = myMarketPositionCache.Get(bet.ContractID)
		if err != nil {
			log.Fatalln(err)
		}
		wg.Done()
	}()
	go func() {
		var err error
		cachedUser, err = usersCache.Get(bet.UserID)
		if err != nil {
			log.Fatalln(err)
		}
		wg.Done()
	}()
	go func() {
		var err error
		marketVelocity, err = marketVelocityCache.Get(bet.ContractID)
		if err != nil {
			log.Fatalln(err)
		}
		wg.Done()
	}()
	wg.Wait()

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

	var probDiff = math.Abs(bet.ProbBefore - bet.ProbAfter)
	var poolSize = cachedMarket.Pool.NO + cachedMarket.Pool.YES // 100 is the current minimum, 1_000 is decently sized, >10_000 is a big market, >100_000 is larger than LK-99
	var poolSizeFactor = math.Min(poolSize, 30_000) / 30_000    // From 0 to 1, 0 being pool is small, 1 being pool is huge
	var minProbSwing = 0.15 - poolSizeFactor*0.14               // 0.15 base, down to 0.01 depending on poolSize
	if probDiff < minProbSwing {
		// Ignore small prob changes
		return false
	}

	if slices.Contains(BANNED_USER_IDS, bet.UserID) {
		// Ignore banned user ids
		return false
	}

	if cachedMarket.CreatorID == bet.UserID {
		// Ignore bets by market creator
		return false
	}

	var outcomeWeWillWantToBuy string
	if bet.ProbBefore > bet.ProbAfter {
		outcomeWeWillWantToBuy = "YES"
	} else {
		outcomeWeWillWantToBuy = "NO"
	}

	if myPosition != nil && myPosition.Invested > 200 && ((outcomeWeWillWantToBuy == "YES" && myPosition.HasYesShares) || (outcomeWeWillWantToBuy == "NO" && !myPosition.HasYesShares)) {
		// Ignore markets where I am too invested on one side. This could be increased in the future to allow larger positions
		return false
	}

	var isNewAccount = cachedUser.CreatedTime > time.Now().UnixMilli()-1000*60*60*24*3
	if isNewAccount && cachedUser.ProfitCachedAllTime > 1300 {
		// Ignore new accounts with large profits
		return false
	}

	if !*marketVelocity {
		// Ignore markets with low volatility. This check could be improved in the future
		return false
	}

	// Return from variable to prevent Go complaining about previous if being redundant
	var returnValue = true
	return returnValue
}
