package ManifoldApi

import (
	"fmt"
)

func CancelBet(betId string) {
	callManifoldApi("POST", fmt.Sprintf("v0/bet/cancel/%s", betId), nil)
}
