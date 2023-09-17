package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"math"
	"sync"
	"time"
)

var roundLock sync.Mutex

var myUserId string

func Run() {
	myUserId = ManifoldApi.GetMe().ID

	log.Println("Velocity module enabled!")

	markNewBetsAllAsSeen() // Only process bets created from now on

	for {
		time.Sleep(time.Millisecond * 500)
		runVelocityRound()
	}
}

func runVelocityRound() {
	roundLock.Lock()
	defer roundLock.Unlock()

	var allBets = getNewGoodForVelocityBets()

	type groupedMarketBet struct {
		prevProb float64
	}
	var groupedMarketBets = make(map[string]*groupedMarketBet)
	for _, bet := range allBets {
		if groupedMarketBets[bet.ContractID] == nil {
			groupedMarketBets[bet.ContractID] = &groupedMarketBet{
				prevProb: bet.ProbBefore,
			}
		} else {
			var savedPrev = groupedMarketBets[bet.ContractID].prevProb
			var myPrev = bet.ProbBefore
			if math.Abs(0.5-savedPrev) < math.Abs(0.5-myPrev) {
				// Save the most extreme prob
				groupedMarketBets[bet.ContractID].prevProb = myPrev
			}
		}
	}

	for marketId, groupedBet := range groupedMarketBets {
		var outcome string
		var cachedMarket = marketsCache.Get(marketId)
		if groupedBet.prevProb > cachedMarket.Probability {
			outcome = "YES"
		} else {
			outcome = "NO"
		}

		// 10 might not be enough to offset api betting fees, we might need to increase in the future
		var amount int64 = 10

		var alpha = 0.7
		var limitProb = math.Round((groupedBet.prevProb*(1-alpha)+cachedMarket.Probability*alpha)*100) / 100

		var betRequest = ManifoldApi.PlaceBetRequest{
			ContractId: marketId,
			Outcome:    outcome,
			Amount:     amount,
			LimitProb:  limitProb,
		}
		log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, groupedBet, betRequest)

		/*
			var placedBet, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
			if err != nil {
				log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
			} else {
				log.Printf("Placed bet. Request: #%+v.\nResponse: %+v\n", betRequest, placedBet)
			}
		*/
	}
}
