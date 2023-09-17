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
		time.Sleep(time.Millisecond * 50)
		runVelocityRound()
	}
}

type groupedMarketBet struct {
	marketId string
	prevProb float64
}

func runVelocityRound() {
	roundLock.Lock()
	defer roundLock.Unlock()

	var allBets = getNewGoodForVelocityBets()

	// Group by market id
	var groupedMarketBets = make(map[string]*groupedMarketBet)
	for _, bet := range allBets {
		if groupedMarketBets[bet.ContractID] == nil {
			groupedMarketBets[bet.ContractID] = &groupedMarketBet{
				marketId: bet.ContractID,
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
	var cachedMarket = marketsCache.Get(groupedBet.marketId)
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
		ContractId: groupedBet.marketId,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  limitProb,
	}
	log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, groupedBet, betRequest)

	var _, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
	if err != nil {
		log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
	}
}
