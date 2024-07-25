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

type betInfo struct {
	marketUrl          string
	betPerformanceInfo betPerformanceInfoType
	bet                utils.ManifoldWebsocketBet
	betRequest         ManifoldApi.PlaceBetRequest
	alpha              float64
}

func (info betPerformanceInfoType) String() string {
	return fmt.Sprintf("{originalBetCreatedAt: %s receivedAt: %s cachesLoadedAt: %s velocityCheckedAt: %s betReqStartedAt: %s betPlacedAt: %s}",
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

	err := utils.SendManifoldApiWebsocketMessage(`{
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

	utils.AddManifoldWebsocketEventListener(func(event utils.ManifoldSocketEvent) {
		if event.Topic == "global/new-bet" {
			var data, err = utils.ParseManifoldNewBetEventPayload(event.Data)
			if err != nil {
				log.Printf("Error while decoding manifold event data: %+\n", err)
			} else {
				for _, bet := range data.Bets {
					go processBet(bet)
				}
			}
		}
	})
}

func processBet(bet utils.ManifoldWebsocketBet) {
	var betPerformanceInfo = betPerformanceInfoType{originalBetCreatedAt: time.UnixMilli(bet.CreatedTime), receivedAt: time.Now()}

	var loadedCaches = loadCachesForBet(bet)
	betPerformanceInfo.cachesLoadedAt = time.Now()

	// [0, 1] The bigger, the more we correct
	var alpha = 0.2 +
		utils.MapNumber(loadedCaches.betCreatorUser.SkillEstimate, 1, 0, 0, 0.4) +
		utils.MapNumber(loadedCaches.marketVelocity, 0, 1, -0.2, 0.3)

	var beforeOdds = utils.ProbToOdds(bet.ProbBefore)
	var afterOdds = utils.ProbToOdds(bet.ProbAfter)
	var correctedOdds = beforeOdds*alpha + afterOdds*(1-alpha)
	var correctedProb = utils.OddsToProb(correctedOdds)
	var limitProb = math.Round(correctedProb*100) / 100 // round for manifold's limit order accuracy

	var outcome = utils.Ternary(bet.ProbBefore > bet.ProbAfter, "YES", "NO")

	// [10, 50]
	var amount = int64(math.Round(
		10 +
			utils.MapNumber(loadedCaches.betCreatorUser.SkillEstimate, 1, 0, 0, 15) +
			utils.MapNumber(loadedCaches.marketVelocity, 0, 1, 0, 25),
	))

	var betRequest = ManifoldApi.PlaceBetRequest{
		ContractId: bet.ContractID,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  limitProb,
	}

	if !isBetGoodForVelocity(bet, loadedCaches, betRequest) {
		// Bet is no good for velocity, ignore
		return
	}
	betPerformanceInfo.velocityCheckedAt = time.Now()

	var doNotActuallyPlaceBets = os.Getenv("VELOCITY_MODULE_DO_NOT_ACTUALLY_PLACE_BETS") == "true"

	if !doNotActuallyPlaceBets {
		betPerformanceInfo.betReqStartedAt = time.Now()
		var myPlacedBet, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
		if err == nil {
			betPerformanceInfo.betPlacedAt = time.UnixMilli(myPlacedBet.CreatedTime)
		} else {
			betPerformanceInfo.betPlacedAt = time.Now()
			var info = betInfo{
				marketUrl:          loadedCaches.market.URL,
				betPerformanceInfo: betPerformanceInfo,
				bet:                bet,
				betRequest:         betRequest,
				alpha:              alpha,
			}
			log.Printf("Error placing bet %+v\nError message: %v\n", info, err)
			return
		}
	}

	if doNotActuallyPlaceBets {
		betPerformanceInfo.betReqStartedAt = time.Now()
		betPerformanceInfo.betPlacedAt = time.Now()
		var info = betInfo{
			marketUrl:          loadedCaches.market.URL,
			betPerformanceInfo: betPerformanceInfo,
			bet:                bet,
			betRequest:         betRequest,
			alpha:              alpha,
		}
		log.Printf("Would've placed velocity bet: %+v\n", info)
	} else {
		var info = betInfo{
			marketUrl:          loadedCaches.market.URL,
			betPerformanceInfo: betPerformanceInfo,
			bet:                bet,
			betRequest:         betRequest,
			alpha:              alpha,
		}
		log.Printf("Placed velocity bet: %+v\n", info)
	}

	// Refresh cache for my market position on this market
	myMarketPositionCache.Renew(bet.ContractID)
}
