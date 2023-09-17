package ManifoldApi

import (
	"encoding/json"
	"fmt"
)

type UserProfitCached struct {
	AllTime float64 `json:"allTime"`
	Monthly float64 `json:"monthly"`
	Daily   float64 `json:"daily"`
	Weekly  float64 `json:"weekly"`
}
type User struct {
	ID                   string           `json:"id"`
	CreatedTime          int64            `json:"createdTime"`
	Name                 string           `json:"name"`
	Username             string           `json:"username"`
	Balance              float64          `json:"balance"`
	TotalDeposits        float64          `json:"totalDeposits"`
	ProfitCached         UserProfitCached `json:"profitCached"`
	IsBot                bool             `json:"isBot"`
	IsAdmin              bool             `json:"isAdmin"`
	IsTrustworthy        bool             `json:"isTrustworthy"`
	CurrentBettingStreak int              `json:"currentBettingStreak"`
	LastBetTime          int64            `json:"lastBetTime"`
}

func GetUser(userId string) User {
	sb := callManifoldApi("GET", fmt.Sprintf("v0/user/by-id/%s", userId), nil)

	var response User
	json.Unmarshal([]byte(sb), &response)
	return response
}
