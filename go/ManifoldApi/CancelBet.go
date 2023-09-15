package ManifoldApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

type cancelBetRequest struct {
	BetId string `json:"betId"`
}

func CancelBet(betId string) {
	var req = cancelBetRequest{
		BetId: betId,
	}

	var body, error = json.Marshal(req)
	if error != nil {
		log.Fatalln(error)
	}

	sb := callManifoldApi("POST", fmt.Sprintf("v0/bet/cancel/%s", betId), bytes.NewBuffer(body))

	var response Bet
	json.Unmarshal([]byte(sb), &response)

	if response.ID == "" || !response.IsCancelled {
		log.Fatalf("failed to cancel bet: %+v", response)
	}
}
