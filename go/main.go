package main

import (
	"ManifoldTradingBot/CompaniesMarketCapApi"
	"ManifoldTradingBot/CpmmMarketUtils"
	"ManifoldTradingBot/ManifoldApi"
	"log"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Market struct {
	manifoldId                         string
	stockSymbol                        string
	marketCapToManifoldValueMultiplier float64
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var marketsDb = []Market{
		{manifoldId: "aZn4kn9dIv5wjQSbVzdk", stockSymbol: "AAPL", marketCapToManifoldValueMultiplier: 0.000000001},
		{manifoldId: "qy4Pujoc7k2G03cb7Vnh", stockSymbol: "AMZN", marketCapToManifoldValueMultiplier: 0.000000001},
		{manifoldId: "RnzTxpnUSsbfPG8Ec6BO", stockSymbol: "GOOG", marketCapToManifoldValueMultiplier: 0.000000001},
		{manifoldId: "1IBrgJ6IlwBIaJ7xdQ5c", stockSymbol: "MSFT", marketCapToManifoldValueMultiplier: 0.000000001},
	}

	for {
		var wg sync.WaitGroup

		log.Println("Starting betting round...")

		for _, m := range marketsDb {
			wg.Add(1)
			go func(market Market) {
				runLogicForMarket(market)
				wg.Done()
			}(m)
		}

		wg.Wait()

		log.Println("Betting round done! Sleeping until next one")

		time.Sleep(time.Hour)
	}
}

func runLogicForMarket(market Market) {
	var betRequest = calculateBetForMarket(market)
	if betRequest.Amount >= 1 {
		var placedBet = ManifoldApi.PlaceBet(betRequest)
		log.Printf("Placed bet #%v: %+v\n", placedBet.BetID, betRequest)
		if !placedBet.IsFilled {
			// Cancel instantly if it had limitProb
			ManifoldApi.CancelBet(placedBet.BetID)
		}
	}
}

func calculateBetForMarket(market Market) ManifoldApi.PlaceBetRequest {
	var marketCap, err = CompaniesMarketCapApi.GetCompanyMarketCap(market.stockSymbol)

	if err != nil {
		log.Fatal(err)
	}

	var manifoldMarket = ManifoldApi.GetMarket(market.manifoldId)
	var expectedMarketValue float64 = float64(marketCap) * market.marketCapToManifoldValueMultiplier
	var expectedMarketProbability = CpmmMarketUtils.ConvertValueToProbability(manifoldMarket, expectedMarketValue)

	var outcome, amount = CpmmMarketUtils.CalculatePseudoNumericMarketplaceBet(manifoldMarket, expectedMarketValue, nil)

	if amount >= 1 {
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
