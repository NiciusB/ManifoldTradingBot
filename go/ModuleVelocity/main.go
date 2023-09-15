package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"math"
	"slices"
	"sort"
	"time"
)

var minProbSwing = 0.2
var requestSecondsInterval int64 = 10
var bannedUserIDs = []string{
	"skGf6ln62qPPMIaxpRZG8JH9wbJ3",
	"ilJdhpLzZZSUgzueJOs2cbRnJn82",
	"jO7sUhIDTQbAJ3w86akzncTlpRG2",
	"w1knZ6yBvEhRThYPEYTwlmGv7N33",
	"BhNkw088bMNwIFF2Aq5Gg9NTPzz1",
	"dNgcgrHGn8ZB30hyDARNtbjvGPm1",
	"kzjQhRJ4GINn5umiq2ee1QvaMcE2",
	"P7PV13rynzOHyxm8AiXIN568bmF2",
	"jOl1FMKpFbXkoaDGp2qlakUxAiJ3",
	"MxdyEeVgrFMTDDsPbXwAe9W1CLs2",
	"IEVDP2LTpgMYaka38r1TVZcabWS2",
	"4JuXgDx47xPagH5mcLDqLzUSN5g2",
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

func Run() {
	log.Println("Velocity module enabled!")

	for {
		time.Sleep(time.Second * time.Duration(requestSecondsInterval))

		log.Println("Starting velocity betting round...")

		// Get top 10 best recent bets
		var bestRecentBets = getBestRecentBets()
		bestRecentBets = bestRecentBets[:min(10, len(bestRecentBets))]

		var bestScore = math.Inf(-1)
		var bestBet *ManifoldApi.Bet

		for _, bet := range bestRecentBets {
			var market = ManifoldApi.GetMarket(bet.ContractID)
			var probDiff = math.Abs(bet.ProbBefore - market.Probability)
			if probDiff < minProbSwing {
				continue
			}
			var score = probDiff
			if score > bestScore {
				var copiedBet = bet
				bestBet = &copiedBet
				bestScore = score
			}
		}

		if bestBet != nil {
			log.Printf("Best bet: #%+v.\n", bestBet)
			var outcome string
			if bestBet.Outcome == "YES" {
				outcome = "NO"
			} else {
				outcome = "YES"
			}

			var amount int64 = 10

			var limitProb = math.Round((bestBet.ProbBefore*0.3+bestBet.ProbAfter*0.7)*100) / 100

			var betRequest = ManifoldApi.PlaceBetRequest{
				ContractId: bestBet.ContractID,
				Outcome:    outcome,
				Amount:     amount,
				LimitProb:  limitProb,
			}

			var placedBet, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
			if err != nil {
				log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
			} else {
				log.Printf("Placed bet. Request: #%+v.\nResponse: %+v\n", betRequest, placedBet)
			}
		} else {
			log.Printf("Did not found any suitable velocity bet")
		}
	}
}

func getBestRecentBets() []ManifoldApi.Bet {
	type scoredBet struct {
		bet   ManifoldApi.Bet
		score float64
	}

	var bets = ManifoldApi.GetBets("", "")

	var scoredBets []scoredBet
	for _, bet := range bets {
		if bet.IsAPI {
			// Ignore bots
			continue
		}

		if bet.ProbAfter >= 0.9 || bet.ProbAfter <= 0.1 {
			// Ignore extreme probabilities
			continue
		}

		var betTime = bet.CreatedTime / 1000
		var timeNow = time.Now().Unix()

		if timeNow-betTime > requestSecondsInterval {
			// Ignore old bets
			continue
		}

		if slices.Contains(bannedUserIDs, bet.UserID) {
			// Do not act on this user
			continue
		}

		var probDiff = math.Abs(bet.ProbBefore - bet.ProbAfter)
		if probDiff < minProbSwing {
			continue
		}

		var score = probDiff
		var scoredBet = scoredBet{
			bet:   bet,
			score: score,
		}
		scoredBets = append(scoredBets, scoredBet)
	}

	sort.SliceStable(scoredBets, func(i, j int) bool {
		return scoredBets[i].score > scoredBets[j].score
	})

	var result []ManifoldApi.Bet
	for _, bet := range scoredBets {
		result = append(result, bet.bet)
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
