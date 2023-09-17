package ManifoldApi

import (
	"encoding/json"
	"fmt"
)

func GetMe() User {
	sb := callManifoldApi("GET", fmt.Sprintf("v0/me"), nil)

	var response User
	json.Unmarshal([]byte(sb), &response)
	return response
}
