package ManifoldApi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func callManifoldApi(method string, path string, reqBody io.Reader) string {
	req, err := http.NewRequest(method, fmt.Sprintf("https://manifold.markets/api/%s", path), reqBody)
	if err != nil {
		log.Fatalln(err)
	}

	var manifoldApiKey = os.Getenv("MANIFOLD_API_KEY")

	req.Header.Add("User-Agent", "ManifoldTradingBot/1.0.0 for @NiciusBot")
	req.Header.Add("Authorization", "Key "+manifoldApiKey)
	if method == "POST" && reqBody != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)

	return sb
}
