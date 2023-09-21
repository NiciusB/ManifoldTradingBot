package ManifoldApi

import (
	"bytes"
	"encoding/json"
	"errors"
)

type PlaceBetRequest struct {
	ContractId string  `json:"contractId"`
	Outcome    string  `json:"outcome"`
	Amount     int64   `json:"amount"`
	LimitProb  float64 `json:"limitProb,omitempty"`
	ExpiresAt  string  `json:"expiresAt,omitempty"`
}

type placeBetResponse struct {
	Message     string  `json:"message"`
	OrderAmount int     `json:"orderAmount"`
	Amount      float64 `json:"amount"`
	Shares      float64 `json:"shares"`
	IsFilled    bool    `json:"isFilled"`
	IsCancelled bool    `json:"isCancelled"`
	Fills       []struct {
		MatchedBetID string  `json:"matchedBetId"`
		Amount       float64 `json:"amount"`
		Shares       float64 `json:"shares"`
		Timestamp    int64   `json:"timestamp"`
	} `json:"fills"`
	ContractID  string  `json:"contractId"`
	Outcome     string  `json:"outcome"`
	ProbBefore  float64 `json:"probBefore"`
	ProbAfter   float64 `json:"probAfter"`
	LoanAmount  int     `json:"loanAmount"`
	CreatedTime int64   `json:"createdTime"`
	Fees        struct {
		CreatorFee   int `json:"creatorFee"`
		PlatformFee  int `json:"platformFee"`
		LiquidityFee int `json:"liquidityFee"`
	} `json:"fees"`
	IsAnte       bool   `json:"isAnte"`
	IsRedemption bool   `json:"isRedemption"`
	IsChallenge  bool   `json:"isChallenge"`
	Visibility   string `json:"visibility"`
	BetID        string `json:"betId"`
}

func PlaceBet(bet PlaceBetRequest) (*placeBetResponse, error) {
	var body, err = json.Marshal(bet)
	if err != nil {
		return nil, err
	}

	sb := callManifoldApiWithFullUrl("POST", "https://api-nggbo3neva-uc.a.run.app/placebet", bytes.NewBuffer(body))

	var response placeBetResponse
	json.Unmarshal([]byte(sb), &response)

	if response.Message != "" {
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

func PlaceInstantlyCancelledLimitOrder(betRequest PlaceBetRequest) (*placeBetResponse, error) {
	var placedBet, err = PlaceBet(betRequest)
	if err != nil {
		return nil, err
	}

	if !placedBet.IsFilled {
		// Cancel instantly if it didn't fully fill
		err := CancelBet(placedBet.BetID)
		if err != nil {
			return nil, err
		}
	}

	return placedBet, nil
}
