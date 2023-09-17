package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"math"
	"slices"
	"sort"
	"sync"
	"time"
)

type seenBetsHistoryType struct {
	lastCreatedTime           int64
	seenBetsOnLastCreatedTime []string
	lock                      sync.Mutex
}

var seenBetsHistory seenBetsHistoryType

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

var MIN_PROB_SWING = 0.1

func markNewBetsAllAsSeen() {
	seenBetsHistory.lastCreatedTime = time.Now().UnixMilli()
}

func getNewGoodForVelocityBets() []ManifoldApi.Bet {
	type scoredBet struct {
		bet   ManifoldApi.Bet
		score float64
	}

	var bets = ManifoldApi.GetBets("", "")

	var scoredBets []scoredBet
	var wg sync.WaitGroup
	for _, bet := range bets {
		wg.Add(1)
		go func(bet ManifoldApi.Bet) {
			defer wg.Done()
			if !isBetGoodForVelocity(bet) {
				return
			}

			var probDiff = math.Abs(bet.ProbBefore - bet.ProbAfter)
			var score = probDiff
			var scoredBet = scoredBet{
				bet:   bet,
				score: score,
			}
			scoredBets = append(scoredBets, scoredBet)

		}(bet)
	}
	wg.Wait()

	// Update seenBetsHistory
	seenBetsHistory.lock.Lock()
	for _, bet := range bets {
		if bet.CreatedTime > seenBetsHistory.lastCreatedTime {
			seenBetsHistory.lastCreatedTime = bet.CreatedTime
		}
	}
	seenBetsHistory.seenBetsOnLastCreatedTime = []string{}
	for _, bet := range bets {
		if bet.CreatedTime == seenBetsHistory.lastCreatedTime {
			seenBetsHistory.seenBetsOnLastCreatedTime = append(seenBetsHistory.seenBetsOnLastCreatedTime, bet.ID)
		}
	}
	seenBetsHistory.lock.Unlock()

	sort.SliceStable(scoredBets, func(i, j int) bool {
		return scoredBets[i].score > scoredBets[j].score
	})

	var result []ManifoldApi.Bet
	for _, bet := range scoredBets {
		result = append(result, bet.bet)
	}
	return result
}

func isBetGoodForVelocity(bet ManifoldApi.Bet) bool {
	if bet.IsAPI {
		// Ignore bots
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

	if bet.CreatedTime < seenBetsHistory.lastCreatedTime || slices.Contains(seenBetsHistory.seenBetsOnLastCreatedTime, bet.ID) {
		// Ignore already seen bets
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

	var cachedMarketPositions = marketPositionsCache.Get(bet.ContractID)
	if len(cachedMarketPositions) < 4 {
		// Ignore markets with less than 4 market positions
		return false
	}

	var myPosition *ManifoldApi.MarketPosition
	for _, pos := range cachedMarketPositions {
		if pos.UserID == myUserId {
			myPosition = &pos
		}
	}
	if myPosition != nil && myPosition.Invested > 200 {
		// Ignore markets where I am too invested. This could be increased in the future to allow larger positions
		return false
	}

	// TODO: Ignore markets with low volatility

	var cachedUser = usersCache.Get(bet.UserID)
	var isNewAccount = cachedUser.CreatedTime > time.Now().UnixMilli()-1000*60*60*24*3
	if isNewAccount && cachedUser.ProfitCached.AllTime > 1000 {
		// Ignore new accounts with large profits
		return false
	}

	// Check market again, but without cache: since we passed all previous checks this is now seriously eligible for betting and requires fresh data
	marketsCache.DeleteCache(bet.ContractID)
	cachedMarket = marketsCache.Get(bet.ContractID)
	probDiff = math.Abs(bet.ProbBefore - cachedMarket.Probability)
	if probDiff < MIN_PROB_SWING && true {
		return false
	}

	return true
}
