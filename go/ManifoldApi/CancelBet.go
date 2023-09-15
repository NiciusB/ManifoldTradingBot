package ManifoldApi

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type cancelBetRequest struct {
	BetId string `json:"betId"`
}

func CancelBet(betId string) error {
	var req = cancelBetRequest{
		BetId: betId,
	}

	var body, err = json.Marshal(req)
	if err != nil {
		return err
	}

	sb := callManifoldApi("POST", fmt.Sprintf("v0/bet/cancel/%s", betId), bytes.NewBuffer(body))

	var response Bet
	json.Unmarshal([]byte(sb), &response)

	if response.ID == "" || !response.IsCancelled {
		return fmt.Errorf("failed to cancel bet: %+v", response)
	}

	return nil
}
