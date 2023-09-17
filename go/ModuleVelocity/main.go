package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"math"
	"sync"
	"time"
)

var roundLock sync.Mutex

var marketsCache = CreateGenericCache(ManifoldApi.GetMarket, time.Minute*30)
var usersCache = CreateGenericCache(ManifoldApi.GetUser, time.Hour)
var marketPositionsCache = CreateGenericCache(ManifoldApi.GetMarketPositions, time.Minute*15)

func Run() {
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

	for _, bet := range allBets {
		var outcome string
		var cachedMarket = marketsCache.Get(bet.ContractID)
		if bet.ProbBefore > cachedMarket.Probability {
			outcome = "YES"
		} else {
			outcome = "NO"
		}

		var amount int64 = 10

		var alpha = 0.7
		var limitProb = math.Round((bet.ProbBefore*(1-alpha)+cachedMarket.Probability*alpha)*100) / 100

		var betRequest = ManifoldApi.PlaceBetRequest{
			ContractId: bet.ContractID,
			Outcome:    outcome,
			Amount:     amount,
			LimitProb:  limitProb,
		}
		log.Printf("Placing velocity bet. Bet: %+v\n Request: %+v.\n", bet, betRequest)

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
