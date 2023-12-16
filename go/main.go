package main

import (
	"ManifoldTradingBot/ManifoldApi"
	modulestock "ManifoldTradingBot/ModuleStock"
	modulevelocity "ManifoldTradingBot/ModuleVelocity"
	"ManifoldTradingBot/utils"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	utils.ConnectSupabaseWebsocket()
	utils.ConnectRedisClient()

	log.Println("Bot started up correctly!")

	var enableStockModule = os.Getenv("ENABLE_STOCK_MODULE") != "false"
	if enableStockModule {
		// go modulestock.Run()
	}

	var enableVelocityModule = os.Getenv("ENABLE_VELOCITY_MODULE") != "false"
	if enableVelocityModule {
		go modulevelocity.Run()
	}

	// Infinite loop for keeping module goroutines alive, since they never end anyway. We use it to monitor api queue length too
	for {
		time.Sleep(time.Second * 5)

		var apiQueueLength = ManifoldApi.GetQueueLength()
		if apiQueueLength >= 100 {
			log.Printf("API queue is long: %v\n", apiQueueLength)
		}
	}
}
