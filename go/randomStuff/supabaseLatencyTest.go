package main

import (
	"ManifoldTradingBot/utils"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	utils.ConnectSupabaseWebsocket()

	err = utils.SendSupabaseWebsocketMessage(`{
		"event": "phx_join",
		"topic": "realtime:*",
		"payload": {
			"config": {
				"broadcast": {
					"self": false
				},
				"presence": {
					"key": ""
				},
				"postgres_changes": [
					{
					"table": "contract_bets",
					"event": "INSERT"
					}
				]
			}
		},
		"ref": null
		}`)
	if err != nil {
		log.Println("subscribing to supabase error:", err)
	}

	log.Println("Connected")

	utils.AddSupabaseWebsocketEventListener(func(event utils.SupabaseEvent) {
		if event.Event == "postgres_changes" {
			var payload, err = utils.ParseSupabasePostgresChangePayload(event.Payload)
			if err != nil {
				log.Printf("Error while decoding postgres_changes: %+\n", err)
			} else {
				var bet = payload.Data.Record.Data

				if bet.IsRedemption {
					// Ignore redemptions, as they are always the opposite of another bet
					return
				}

				for _, fill := range bet.Fills {
					if fill.Timestamp != bet.CreatedTime {
						// Ignore fills of limit orders, other than the initial one
						return
					}
				}

				var timeDiff = time.Since(time.UnixMilli(bet.CreatedTime))
				println(timeDiff.Milliseconds())
			}
		}
	})

	for {
		time.Sleep(time.Hour)
	}
}
