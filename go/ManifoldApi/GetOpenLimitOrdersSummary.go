package ManifoldApi

type getOpenLimitOrdersSummaryResponse = map[float64]float64

func GetOpenLimitOrdersSummary(marketId string) getOpenLimitOrdersSummaryResponse {
	var openOrders = GetOpenLimitOrders(marketId)
	var response = make(getOpenLimitOrdersSummaryResponse)

	for _, bet := range openOrders {
		response[bet.LimitProb] += bet.OrderAmount - bet.Amount
	}

	return response
}
