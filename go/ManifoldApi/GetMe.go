package ManifoldApi

import (
	"encoding/json"
)

func GetMe() User {
	sb := callManifoldApi("GET", "v0/me", nil)

	var response User
	json.Unmarshal([]byte(sb), &response)
	return response
}
