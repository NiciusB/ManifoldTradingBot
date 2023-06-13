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

func GetBets(contractId string, before string) []Bet {
	sb := callManifoldApi("GET", fmt.Sprintf("v0/bets?contractId=%s&before=%s", contractId, before), nil)

	var response []Bet
	json.Unmarshal([]byte(sb), &response)
	return response
}
