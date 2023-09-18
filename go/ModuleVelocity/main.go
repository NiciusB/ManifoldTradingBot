package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"log"
	"math"
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
				go warmupCachesForBet(bet)
				processBet(bet)
			}
		}
	})
}

func warmupCachesForBet(bet SupabaseBet) {
	go marketsCache.Get(bet.ContractID)
	go myMarketPositionCache.Get(bet.ContractID)
	go usersCache.Get(bet.UserID)
	go marketVelocityCache.Get(bet.ContractID)
}

func processBet(bet SupabaseBet) {
	if isBetGoodForVelocity(bet) {
		placeBet(&groupedMarketBet{
			marketId:   bet.ContractID,
			probBefore: bet.ProbBefore,
			probAfter:  bet.ProbAfter,
		})
	}
}

type groupedMarketBet struct {
	marketId   string
	probBefore float64
	probAfter  float64
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

	var cachedMarket, err = marketsCache.Get(groupedBet.marketId)
	if err != nil {
		log.Printf("Error getting market from cache. Error message: %v\n", err)
		return
	}

	log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, groupedBet, betRequest)

	var doNotActuallyPlaceBets = os.Getenv("VELOCITY_MODULE_DO_NOT_ACTUALLY_PLACE_BETS") == "true"
	if !doNotActuallyPlaceBets {
		var _, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
		if err != nil {
			log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
		}
	}

	// Refresh cache for my market position on this market
	myMarketPositionCache.Delete(groupedBet.marketId)
	myMarketPositionCache.Get(groupedBet.marketId)
}
