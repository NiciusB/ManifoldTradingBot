package modulevelocity

import (
	"ManifoldTradingBot/ManifoldApi"
	"ManifoldTradingBot/utils"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

var myUserId string

type betPerformanceInfoType struct {
	originalBetCreatedAt time.Time
	receivedAt           time.Time
	cachesLoadedAt       time.Time
	velocityCheckedAt    time.Time
	betReqStartedAt      time.Time
	betPlacedAt          time.Time
}

func (info betPerformanceInfoType) String() string {
	return fmt.Sprintf("{originalBetCreatedAt: %s receivedAt: %s cachesLoadedAt: +%s velocityCheckedAt: +%s betReqStartedAt: +%s betPlacedAt: +%s}",
		info.originalBetCreatedAt.Format("2006-01-02 15:04:05.999999999 -0700"),
		info.receivedAt.Sub(info.originalBetCreatedAt).String(),
		info.cachesLoadedAt.Sub(info.originalBetCreatedAt).String(),
		info.velocityCheckedAt.Sub(info.originalBetCreatedAt).String(),
		info.betReqStartedAt.Sub(info.originalBetCreatedAt).String(),
		info.betPlacedAt.Sub(info.originalBetCreatedAt).String(),
	)
}

func Run() {
	myUserId = ManifoldApi.GetMe().ID

	err := utils.SendSupabaseWebsocketMessage(`{
		"event": "phx_join",
		"topic": "realtime:*",
		"payload": {
			"config": {
				"broadcast": {
					"self": false
				},
				"presence": {
					"key": ""
				},
				"postgres_changes": [
					{
					"table": "contract_bets",
					"event": "INSERT"
					}
				]
			}
		},
		"ref": null
		}`)
	if err != nil {
		log.Println("subscribing to supabase error:", err)
	}

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
	var betPerformanceInfo = betPerformanceInfoType{originalBetCreatedAt: time.UnixMilli(bet.CreatedTime), receivedAt: time.Now()}

	var loadedCaches = loadCachesForBet(bet)
	betPerformanceInfo.cachesLoadedAt = time.Now()

	var alpha = 0.8 - loadedCaches.betCreatorUser.SkillEstimate*0.3 // [0, 1] (right now: [0.5, 0.8]). The bigger, the more we correct
	var limitProb = math.Round((bet.ProbBefore*alpha+bet.ProbAfter*(1-alpha))*100) / 100

	if !isBetGoodForVelocity(bet, loadedCaches, limitProb) {
		// Bet is no good for velocity, ignore
		return
	}
	betPerformanceInfo.velocityCheckedAt = time.Now()

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

	if !doNotActuallyPlaceBets {
		betPerformanceInfo.betReqStartedAt = time.Now()
		var myPlacedBet, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
		if err == nil {
			betPerformanceInfo.betPlacedAt = time.UnixMilli(myPlacedBet.CreatedTime)
		} else {
			betPerformanceInfo.betPlacedAt = time.Now()
			log.Printf("Error placing bet on market: %v\nBet info: %+v\nOur bet: %+v\nBet performance: %v\nError message: %v\n", loadedCaches.market.URL, bet, betRequest, betPerformanceInfo, err)
			return
		}
	}

	if doNotActuallyPlaceBets {
		betPerformanceInfo.betReqStartedAt = time.Now()
		betPerformanceInfo.betPlacedAt = time.Now()
		log.Printf("Would've placed velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\nBet performance: %v\n", loadedCaches.market.URL, bet, betRequest, betPerformanceInfo)
	} else {
		log.Printf("Placed velocity bet on market: %v\nBet info: %+v\nOur bet: %+v\nBet performance: %v\n", loadedCaches.market.URL, bet, betRequest, betPerformanceInfo)
	}

	// Refresh cache for my market position on this market
	myMarketPositionCache.Renew(bet.ContractID)
}
