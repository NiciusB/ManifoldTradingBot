package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"math"
	"sort"
	"time"
)

var minProbSwing = 0.2
var requestSecondsInterval int64 = 5

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
