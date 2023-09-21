package utils

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

type SupabaseBet struct {
	Amount      int    `json:"amount"`
	AnswerID    string `json:"answerID"`
	ContractID  string `json:"contractId"`
	CreatedTime int64  `json:"createdTime"`
	Fees        struct {
		CreatorFee   int `json:"creatorFee"`
		LiquidityFee int `json:"liquidityFee"`
		PlatformFee  int `json:"platformFee"`
	} `json:"fees"`
	Fills []struct {
		Amount       int     `json:"amount"`
		MatchedBetID string  `json:"matchedBetId"`
		Shares       float64 `json:"shares"`
		Timestamp    int64   `json:"timestamp"`
	} `json:"fills"`
	ID            string  `json:"id"`
	IsAnte        bool    `json:"isAnte"`
	IsAPI         bool    `json:"isApi"`
	IsCancelled   bool    `json:"isCancelled"`
	IsChallenge   bool    `json:"isChallenge"`
	IsFilled      bool    `json:"isFilled"`
	IsRedemption  bool    `json:"isRedemption"`
	LoanAmount    int     `json:"loanAmount"`
	OrderAmount   int     `json:"orderAmount"`
	Outcome       string  `json:"outcome"`
	ProbAfter     float64 `json:"probAfter"`
	ProbBefore    float64 `json:"probBefore"`
	Shares        float64 `json:"shares"`
	UserAvatarURL string  `json:"userAvatarUrl"`
	UserID        string  `json:"userId"`
	UserName      string  `json:"userName"`
	UserUsername  string  `json:"userUsername"`
	Visibility    string  `json:"visibility"`
}
type postgresChangesPayload struct {
	Data struct {
		Columns []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"columns"`
		CommitTimestamp time.Time   `json:"commit_timestamp"`
		Errors          interface{} `json:"errors"`
		Record          struct {
			Data SupabaseBet `json:"data"`
		} `json:"record"`
		Schema string `json:"schema"`
		Table  string `json:"table"`
		Type   string `json:"type"`
	}
	Ids []int
}

func ParseSupabasePostgresChangePayload(payloadRaw interface{}) (*postgresChangesPayload, error) {
	var result postgresChangesPayload
	var err = mapstructure.Decode(payloadRaw, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
