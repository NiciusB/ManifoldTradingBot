package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"log"
	"math"
	"math/rand"
	"os"
)

var myUserId string

func Run() {
	myUserId = ManifoldApi.GetMe().ID

	log.Println("Velocity module enabled!")

	utils.AddSupabaseWebsocketEventListener(func(event utils.SupabaseEvent) {
		if event.Event == "postgres_changes" {
			var payload, err = parseSupabasePostgresChangePayload(event.Payload)
			if err != nil {
				log.Printf("Error while decoding postgres_changes: %+\n", err)
			} else {
				var bet = payload.Data.Record.Data
				processBet(bet)
			}
		}
	})
}

func processBet(bet SupabaseBet) {
	if !isBetGoodForVelocity(bet) {
		return
	}

	// [0.7, 0.8)
	var alpha = rand.Float64()*0.1 + 0.7
	var limitProb = math.Round((bet.ProbBefore*(1-alpha)+bet.ProbAfter*alpha)*100) / 100

	if limitProb < math.Min(bet.ProbAfter, bet.ProbBefore) && limitProb > math.Max(bet.ProbAfter, bet.ProbBefore) {
		// We do not have enough granularity on the limit, ignore
		return
	}

	var outcome string
	if bet.ProbBefore > bet.ProbAfter {
		outcome = "YES"
	} else {
		outcome = "NO"
	}

	// [10, 30]. Might not be enough to offset api betting fees, we might need to increase in the future
	var amount int64 = rand.Int63n(20+1) + 10

	var betRequest = ManifoldApi.PlaceBetRequest{
		ContractId: bet.ContractID,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  limitProb,
	}

	var doNotActuallyPlaceBets = os.Getenv("VELOCITY_MODULE_DO_NOT_ACTUALLY_PLACE_BETS") == "true"

	go func() {
		var cachedMarket, _ = marketsCache.Get(bet.ContractID)
		if doNotActuallyPlaceBets {
			log.Printf("Would've placed velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, bet, betRequest)
		} else {
			log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, bet, betRequest)
		}
	}()

	if !doNotActuallyPlaceBets {
		var _, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
		if err != nil {
			log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
		}
	}

	// Refresh cache for my market position on this market
	myMarketPositionCache.Renew(bet.ContractID)
}
