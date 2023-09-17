package ManifoldApi

import (
	"encoding/json"
	"fmt"
)

type Bet struct {
	ID   string `json:"id"`
	Fees struct {
		CreatorFee   int `json:"creatorFee"`
		PlatformFee  int `json:"platformFee"`
		LiquidityFee int `json:"liquidityFee"`
	} `json:"fees"`
	Amount        float64 `json:"amount"`
	IsAnte        bool    `json:"isAnte"`
	Shares        float64 `json:"shares"`
	UserID        string  `json:"userId"`
	Outcome       string  `json:"outcome"`
	AnswerId      string  `json:"answerId"`
	ProbAfter     float64 `json:"probAfter"`
	ContractID    string  `json:"contractId"`
	LoanAmount    float64 `json:"loanAmount"`
	ProbBefore    float64 `json:"probBefore"`
	Visibility    string  `json:"visibility"`
	CreatedTime   int64   `json:"createdTime"`
	IsChallenge   bool    `json:"isChallenge"`
	IsRedemption  bool    `json:"isRedemption"`
	IsAPI         bool    `json:"isApi,omitempty"`
	IsFilled      *bool   `json:"isFilled,omitempty"`
	UserName      string  `json:"userName,omitempty"`
	IsCancelled   bool    `json:"isCancelled,omitempty"`
	OrderAmount   float64 `json:"orderAmount,omitempty"`
	UserUsername  string  `json:"userUsername,omitempty"`
	UserAvatarURL string  `json:"userAvatarUrl,omitempty"`
	LimitProb     float64 `json:"limitProb,omitempty"`
	Fills         []struct {
		Amount       float64 `json:"amount"`
		Shares       float64 `json:"shares"`
		Timestamp    int64   `json:"timestamp"`
		MatchedBetID string  `json:"matchedBetId"`
	} `json:"fills,omitempty"`
}

var getBetsAfterTimestampIterations = 0

func GetBetsAfterTimestamp(MinCreatedTime int64) []Bet {
	var limit = 3
	getBetsAfterTimestampIterations++

	var lastBetId = ""
	var response []Bet

	for {
		sb := callManifoldApi("GET", fmt.Sprintf("v0/bets?limit=%v&before=%s&cacheBust=%v", limit, lastBetId, getBetsAfterTimestampIterations%999999999999999), nil)
		var bets []Bet
		json.Unmarshal([]byte(sb), &bets)

		var filteredBets []Bet
		for _, bet := range bets {
			if bet.CreatedTime >= MinCreatedTime {
				filteredBets = append(filteredBets, bet)
			}
		}

		response = append(response, filteredBets...)

		if len(filteredBets) < limit {
			break
		}

		lastBetId = bets[len(bets)-1].ID

		if limit < 50 {
			limit = 50
		} else if limit < 500 {
			limit = 500
		} else {
			limit = 1000
		}
	}

	return response
}

func GetAllBetsForMarket(marketId string) []Bet {
	var lastBetId = ""
	var response []Bet
	var limit = 1000

	for {
		sb := callManifoldApi("GET", fmt.Sprintf("v0/bets?contractId=%s&before=%s&limit=%v", marketId, lastBetId, limit), nil)
		var bets []Bet
		json.Unmarshal([]byte(sb), &bets)

		response = append(response, bets...)

		if len(bets) < limit {
			break
		}

		lastBetId = bets[len(bets)-1].ID
	}

	return response
}

type getOpenLimitOrdersSummaryResponse = map[float64]float64

func GetOpenLimitOrdersSummary(marketId string) getOpenLimitOrdersSummaryResponse {
	var allBets = GetAllBetsForMarket(marketId)

	var response = make(getOpenLimitOrdersSummaryResponse)

	for _, bet := range allBets {
		if !bet.IsCancelled && bet.IsFilled != nil && !*bet.IsFilled {
			response[bet.LimitProb] += bet.OrderAmount - bet.Amount
		}
	}

	return response
}
