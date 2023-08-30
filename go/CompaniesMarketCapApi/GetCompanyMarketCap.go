package CompaniesMarketCapApi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type api8MarketCapResponse struct {
	ID            string `json:"id"`
	Moves         int    `json:"moves"`
	Rank          string `json:"rank"`
	Ticker        string `json:"ticker"`
	TickerClean   string `json:"ticker_clean"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	Type          string `json:"type"`
	LogoURL       string `json:"logo_url"`
	LogoURLDark   string `json:"logo_url_dark"`
	ImgPath       string `json:"imgPath"`
	Marketcap     int    `json:"marketcap"`
	MarketcapText string `json:"marketcapText"`
	Price         string `json:"price"`
	PriceText     string `json:"priceText"`
	OneD          string `json:"1d"`
	OneDText      string `json:"1dText"`
	Sparkline     string `json:"sparkline"`
	SparkColor    string `json:"sparkColor"`
	Favorite      bool   `json:"favorite"`
}

func GetCompanyMarketCap(stockSymbol string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://8marketcap.com/api.php?action=search&query=%s", stockSymbol))
	if err != nil {
		log.Fatalln(err)
	}
	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//Convert the body to type string
	sb := string(body)

	var apiResponse []api8MarketCapResponse
	json.Unmarshal([]byte(sb), &apiResponse)

	for _, item := range apiResponse {
		if item.Ticker == stockSymbol {
			return item.Marketcap, nil
		}
	}

	return 0, fmt.Errorf("unable to find market cap for company: %s", stockSymbol)
}
