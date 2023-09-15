package modulestock

import (
	"ManifoldTradingBot/CompaniesMarketCapApi"
	"ManifoldTradingBot/CpmmMarketUtils"
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"os"
	"sync"
	"time"
)

type stockMarket struct {
	manifoldId                         string
	stockSymbol                        string
	marketCapToManifoldValueMultiplier float64
}

// 1 would be the minimum bet amount allowed by manifold, we do 3 to preserve mana due to placing bets via API fees
var minimumAmountForBet int64 = 3

func Run() {
	log.Println("Stock module enabled!")

	var marketsDb = []stockMarket{
		{manifoldId: "aZn4kn9dIv5wjQSbVzdk", stockSymbol: "AAPL", marketCapToManifoldValueMultiplier: 0.000000001},
		{manifoldId: "qy4Pujoc7k2G03cb7Vnh", stockSymbol: "AMZN", marketCapToManifoldValueMultiplier: 0.000000001},
		{manifoldId: "RnzTxpnUSsbfPG8Ec6BO", stockSymbol: "GOOG", marketCapToManifoldValueMultiplier: 0.000000001},
		{manifoldId: "1IBrgJ6IlwBIaJ7xdQ5c", stockSymbol: "MSFT", marketCapToManifoldValueMultiplier: 0.000000001},
	}

	var waitBeforeFirstBet = os.Getenv("STOCK_MODULE_WAIT_BEFORE_FIRST_BET") != "false"

	var iteration = 0
	for {
		iteration++
		if waitBeforeFirstBet || iteration > 1 {
			log.Println("Sleeping for an hour until next stock betting round...")
			time.Sleep(time.Hour)
		}

		var wg sync.WaitGroup

		log.Println("Starting stock betting round...")

		for _, m := range marketsDb {
			wg.Add(1)
			go func(market stockMarket) {
				runLogicForMarket(market)
				wg.Done()
			}(m)
		}

		wg.Wait()

		log.Println("Betting stock round done!")
	}
}

func runLogicForMarket(market stockMarket) {
	var betRequest = calculateBetForMarket(market)
	if betRequest.Amount >= minimumAmountForBet {
		var placedBet, err = ManifoldApi.PlaceInstantlyCancelledLimitOrder(betRequest)
		if err != nil {
			log.Printf("Error placing bet. Request: #%+v.\nError message: %v\n", betRequest, err)
		} else {
			log.Printf("Placed bet. Request: #%+v.\nResponse: %+v\n", betRequest, placedBet)
		}
	}
}

func calculateBetForMarket(market stockMarket) ManifoldApi.PlaceBetRequest {
	var marketCap, err = CompaniesMarketCapApi.GetCompanyMarketCap(market.stockSymbol)

	if err != nil {
		log.Fatal(err)
	}

	var manifoldMarket = ManifoldApi.GetMarket(market.manifoldId)
	var expectedMarketValue float64 = float64(marketCap) * market.marketCapToManifoldValueMultiplier
	var expectedMarketProbability = CpmmMarketUtils.ConvertValueToProbability(manifoldMarket, expectedMarketValue)

	var outcome, amount = CpmmMarketUtils.CalculatePseudoNumericMarketplaceBet(manifoldMarket, expectedMarketValue, nil)

	if amount >= minimumAmountForBet {
		var limitOrdersSummary = ManifoldApi.GetOpenLimitOrdersSummary(market.manifoldId)

		outcome, amount = CpmmMarketUtils.CalculatePseudoNumericMarketplaceBet(manifoldMarket, expectedMarketValue, limitOrdersSummary)
	}

	var betRequest = ManifoldApi.PlaceBetRequest{
		ContractId: market.manifoldId,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  expectedMarketProbability,
	}

	return betRequest
}
