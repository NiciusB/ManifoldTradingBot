package main

import (
	modulestock "ManifoldTradingBot/ModuleStock"
	modulevelocity "ManifoldTradingBot/ModuleVelocity"
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

	log.Println("Bot started up correctly!")

	var enableStockModule = os.Getenv("ENABLE_STOCK_MODULE") != "false"
	if enableStockModule {
		go modulestock.Run()
	}

	var enableVelocityModule = os.Getenv("ENABLE_VELOCITY_MODULE") != "false"
	if enableVelocityModule {
		go modulevelocity.Run()
	}

	// Infinite loop for keeping module goroutines alive, since they never end anyway
	for {
		time.Sleep(time.Second)
	}
}
