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
				processBet(bet)
			}
		}
	})
}

func processBet(bet SupabaseBet) {
	var loadedCaches = loadCachesForBet(bet)

	var alpha = 0.8 - loadedCaches.betCreatorUser.SkillEstimate*0.3 // [0, 1] (right now: [0.5, 0.8]). The bigger, the more we correct
	var limitProb = math.Round((bet.ProbBefore*alpha+bet.ProbAfter*(1-alpha))*100) / 100

	if !isBetGoodForVelocity(bet, loadedCaches, limitProb) {
		// Bet is no good for velocity, ignore
		return
	}

	var outcome = utils.Ternary(bet.ProbBefore > bet.ProbAfter, "YES", "NO")

	// [8, 30]. Might not be enough to offset api betting fees, we might need to increase in the future
	var amount int64 = int64(math.Round(utils.MapNumber(loadedCaches.betCreatorUser.SkillEstimate, 1, 0, 8, 30)))

	var betRequest = ManifoldApi.PlaceBetRequest{
		ContractId: bet.ContractID,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  limitProb,
	}

	var doNotActuallyPlaceBets = os.Getenv("VELOCITY_MODULE_DO_NOT_ACTUALLY_PLACE_BETS") == "true"

	go func() {
		if doNotActuallyPlaceBets {
			log.Printf("Would've placed velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", loadedCaches.market.URL, bet, betRequest)
		} else {
			log.Printf("Placing velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\n", loadedCaches.market.URL, bet, betRequest)
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
