package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"math"
	"sync"
	"time"
)

var myUserId string

func Run() {
	myUserId = ManifoldApi.GetMe().ID

	log.Println("Velocity module enabled!")

	markNewBetsAllAsSeen() // Only process bets created from now on

	for {
		time.Sleep(time.Millisecond * 99)
		if ManifoldApi.GetThroughputFillPercentage() < 0.5 {
			go runVelocityRound()
		}
	}
}

type groupedMarketBet struct {
	marketId   string
	probBefore float64
	probAfter  float64
}

func runVelocityRound() {
	var allBets = getNewGoodForVelocityBets()

	// Group by market id
	var groupedMarketBets = make(map[string]*groupedMarketBet)
	for _, bet := range allBets {
		if groupedMarketBets[bet.ContractID] == nil {
			groupedMarketBets[bet.ContractID] = &groupedMarketBet{
				marketId:   bet.ContractID,
				probBefore: bet.ProbBefore,
				probAfter:  bet.ProbAfter,
			}
		} else {
			// Save the most extreme probs
			if math.Abs(0.5-groupedMarketBets[bet.ContractID].probBefore) < math.Abs(0.5-bet.ProbBefore) {
				groupedMarketBets[bet.ContractID].probBefore = bet.ProbBefore
			}
			if math.Abs(0.5-groupedMarketBets[bet.ContractID].probAfter) < math.Abs(0.5-bet.ProbAfter) {
				groupedMarketBets[bet.ContractID].probAfter = bet.ProbAfter
			}
		}
	}

	// Place bets in parallel
	var wg sync.WaitGroup
	for _, groupedBet := range groupedMarketBets {
		wg.Add(1)
		go func(groupedBet *groupedMarketBet) {
			defer wg.Done()
			placeBet(groupedBet)
		}(groupedBet)
	}
	wg.Wait()
}

func placeBet(groupedBet *groupedMarketBet) {
	var outcome string
	if groupedBet.probBefore > groupedBet.probAfter {
		outcome = "YES"
	} else {
		outcome = "NO"
	}

	// 10 might not be enough to offset api betting fees, we might need to increase in the future
	var amount int64 = 10

	var alpha = 0.85
	var limitProb = math.Round((groupedBet.probBefore*(1-alpha)+groupedBet.probAfter*alpha)*100) / 100

	var betRequest = ManifoldApi.PlaceBetRequest{
		ContractId: groupedBet.marketId,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  limitProb,
	}

	var cachedMarket = marketsCache.Get(groupedBet.marketId)
	log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, groupedBet, betRequest)

	var _, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
	if err != nil {
		log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
	}

	marketPositionsCache.DeleteCache(groupedBet.marketId)
}
