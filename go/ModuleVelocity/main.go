package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"log"
	"math"
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
				processBet(*payload)
			}
		}
	})
}

func processBet(payload postgresChangesPayload) {
	var bet = payload.Data.Record.Data

	if isBetGoodForVelocity(payload) {
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

	var cachedMarket = marketsCache.Get(groupedBet.marketId)
	log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", cachedMarket.URL, groupedBet, betRequest)

	/*var _, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
	if err != nil {
		log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
	}*/

	marketPositionsCache.DeleteCache(groupedBet.marketId)
}
