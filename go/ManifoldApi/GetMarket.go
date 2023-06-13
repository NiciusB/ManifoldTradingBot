package ManifoldApi

import (
	"encoding/json"
	"fmt"
)

type MarketPool struct {
	NO  float64 `json:"NO"`
	YES float64 `json:"YES"`
}
type Market struct {
	ID                    string     `json:"id"`
	CreatorID             string     `json:"creatorId"`
	CreatorUsername       string     `json:"creatorUsername"`
	CreatorName           string     `json:"creatorName"`
	CreatedTime           int64      `json:"createdTime"`
	CreatorAvatarURL      string     `json:"creatorAvatarUrl"`
	CloseTime             int64      `json:"closeTime"`
	Question              string     `json:"question"`
	URL                   string     `json:"url"`
	Pool                  MarketPool `json:"pool"`
	Probability           float64    `json:"probability"`
	Min                   float64    `json:"min"`
	Max                   float64    `json:"max"`
	P                     float64    `json:"p"`
	TotalLiquidity        float64    `json:"totalLiquidity"`
	OutcomeType           string     `json:"outcomeType"`
	Mechanism             string     `json:"mechanism"`
	Volume                float64    `json:"volume"`
	Volume24Hours         int        `json:"volume24Hours"`
	IsResolved            bool       `json:"isResolved"`
	Resolution            string     `json:"resolution"`
	ResolutionTime        int64      `json:"resolutionTime"`
	ResolutionProbability float64    `json:"resolutionProbability"`
	LastUpdatedTime       int64      `json:"lastUpdatedTime"`
	IsLogScale            bool       `json:"isLogScale"`
}

func GetMarket(marketId string) Market {
	sb := callManifoldApi("GET", fmt.Sprintf("v0/market/%s", marketId), nil)

	var response Market
	json.Unmarshal([]byte(sb), &response)
	return response
}
