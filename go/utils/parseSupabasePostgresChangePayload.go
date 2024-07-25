package utils

import (
	"github.com/mitchellh/mapstructure"
)

type ManifoldWebsocketBet struct {
	ID   string `json:"id"`
	Fees struct {
		CreatorFee   float64 `json:"creatorFee"`
		PlatformFee  int     `json:"platformFee"`
		LiquidityFee int     `json:"liquidityFee"`
	} `json:"fees"`
	Fills []struct {
		Fees struct {
			CreatorFee   float64 `json:"creatorFee"`
			PlatformFee  int     `json:"platformFee"`
			LiquidityFee int     `json:"liquidityFee"`
		} `json:"fees"`
		Amount       float64     `json:"amount"`
		Shares       float64     `json:"shares"`
		Timestamp    int64       `json:"timestamp"`
		MatchedBetID interface{} `json:"matchedBetId"`
	} `json:"fills"`
	IsAPI        bool    `json:"isApi"`
	Amount       float64 `json:"amount"`
	Shares       float64 `json:"shares"`
	UserID       string  `json:"userId"`
	Outcome      string  `json:"outcome"`
	AnswerID     string  `json:"answerId"`
	IsFilled     bool    `json:"isFilled"`
	ExpiresAt    int64   `json:"expiresAt,omitempty"`
	LimitProb    float64 `json:"limitProb,omitempty"`
	ProbAfter    float64 `json:"probAfter"`
	BetGroupID   string  `json:"betGroupId"`
	ContractID   string  `json:"contractId"`
	LoanAmount   int     `json:"loanAmount"`
	ProbBefore   float64 `json:"probBefore"`
	Visibility   string  `json:"visibility"`
	CreatedTime  int64   `json:"createdTime"`
	IsCancelled  bool    `json:"isCancelled"`
	OrderAmount  int     `json:"orderAmount"`
	IsRedemption bool    `json:"isRedemption"`
	BetID        string  `json:"betId"`
}
type ManifolWebsocketNewBetEventData struct {
	Bets []ManifoldWebsocketBet `json:"bets"`
}

func ParseManifoldNewBetEventPayload(dataRaw interface{}) (*ManifolWebsocketNewBetEventData, error) {
	var result ManifolWebsocketNewBetEventData
	var err = mapstructure.Decode(dataRaw, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
