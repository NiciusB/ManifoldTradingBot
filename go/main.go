package main

import (
	"ManifoldTradingBot/CompaniesMarketCapApi"
	"ManifoldTradingBot/CpmmMarketUtils"
	"ManifoldTradingBot/ManifoldApi"
	"log"
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
		for _, market := range marketsDb {
			go runLogicForMarket(market)
		}

		time.Sleep(time.Hour)
	}
}

func runLogicForMarket(market Market) {
	var betRequest = calculateBetForMarket(market)
	if betRequest.Amount >= 1 {
		var placedBet = ManifoldApi.PlaceBet(betRequest)
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

	var outcome, amount = CpmmMarketUtils.CalculatePseudoNumericMarketplaceBet(manifoldMarket, expectedMarketValue)

	if amount >= 1 {
		log.Printf("Found bet for %v, correcting from %v to %v using %v mana", manifoldMarket.Question, manifoldMarket.Probability, expectedMarketProbability, amount)

		var limitOrdersSummary = ManifoldApi.GetOpenLimitOrdersSummary(market.manifoldId)

		for limitOrderProbability, limitOrderAmount := range limitOrdersSummary {
			var shouldAddExtraAmount bool
			if outcome == "YES" {
				shouldAddExtraAmount = limitOrderProbability < expectedMarketProbability && limitOrderProbability > manifoldMarket.Probability
			}
			if outcome == "NO" {
				shouldAddExtraAmount = limitOrderProbability > expectedMarketProbability && limitOrderProbability < manifoldMarket.Probability
			}

			if shouldAddExtraAmount {
				amount += int64(limitOrderAmount) // This is incorrect, we should calculate how much to add somehow by using the current probability as well, since at for example 35%, 10 YES does not equal 10 NO
				log.Printf("Adding to bet on market %v: %v mana because of a limit order at %v, and we want to get to %v\n", manifoldMarket.Question, limitOrderAmount, limitOrderProbability, expectedMarketProbability)
			}
		}
	}

	var betRequest = ManifoldApi.PlaceBetRequest{
		ContractId: market.manifoldId,
		Outcome:    outcome,
		Amount:     amount,
		LimitProb:  expectedMarketProbability,
	}

	if betRequest.Amount >= 1 {
		log.Printf("Placing bet: %+v\n", betRequest)
	}

	return betRequest
}
