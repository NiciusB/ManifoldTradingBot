package ManifoldApi

type getOpenLimitOrdersResponse = Bet

func GetOpenLimitOrders(marketId string) []getOpenLimitOrdersResponse {
	var lastBetId = ""
	var response []getOpenLimitOrdersResponse

	for {
		var bets = GetBets(marketId, lastBetId)
		if len(bets) == 0 {
			break
		}

		lastBetId = bets[len(bets)-1].ID

		for _, bet := range bets {
			if !bet.IsCancelled && bet.IsFilled != nil && !*bet.IsFilled {
				response = append(response, bet)
			}
		}
	}

	return response
}
