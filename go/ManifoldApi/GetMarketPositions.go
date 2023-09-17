package ManifoldApi

import (
	"encoding/json"
	"fmt"
)

type MarketPosition []struct {
	From struct {
		Day struct {
			Value         float64 `json:"value"`
			Profit        int     `json:"profit"`
			Invested      float64 `json:"invested"`
			PrevValue     float64 `json:"prevValue"`
			ProfitPercent int     `json:"profitPercent"`
		} `json:"day"`
		Week struct {
			Value         float64 `json:"value"`
			Profit        int     `json:"profit"`
			Invested      float64 `json:"invested"`
			PrevValue     float64 `json:"prevValue"`
			ProfitPercent int     `json:"profitPercent"`
		} `json:"week"`
		Month struct {
			Value         float64 `json:"value"`
			Profit        float64 `json:"profit"`
			Invested      float64 `json:"invested"`
			PrevValue     float64 `json:"prevValue"`
			ProfitPercent float64 `json:"profitPercent"`
		} `json:"month"`
	} `json:"from"`
	Loan        float64 `json:"loan"`
	Payout      float64 `json:"payout"`
	Profit      float64 `json:"profit"`
	UserID      string  `json:"userId"`
	Invested    float64 `json:"invested"`
	UserName    string  `json:"userName"`
	HasShares   bool    `json:"hasShares"`
	ContractID  string  `json:"contractId"`
	HasNoShares bool    `json:"hasNoShares"`
	LastBetTime int64   `json:"lastBetTime"`
	TotalShares struct {
		NO  int     `json:"NO"`
		YES float64 `json:"YES"`
	} `json:"totalShares"`
	HasYesShares     bool    `json:"hasYesShares"`
	UserUsername     string  `json:"userUsername"`
	ProfitPercent    float64 `json:"profitPercent"`
	UserAvatarURL    string  `json:"userAvatarUrl"`
	MaxSharesOutcome string  `json:"maxSharesOutcome"`
}

func GetMarketPositions(marketId string) []MarketPosition {
	sb := callManifoldApi("GET", fmt.Sprintf("v0/market/%s/positions", marketId), nil)

	var response []MarketPosition
	json.Unmarshal([]byte(sb), &response)
	return response
}
