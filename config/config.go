package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TeleApiID        int
	TeleApiHash      string
	TeleChannelName  string
	BinanceApiKey    string
	BinanceSecretKey string
}

var AppConfig *Config

func LoadConfig() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, fallback to OS env")
	}

	apiId, err := strconv.Atoi(os.Getenv("TELE_API_ID"))
	if err != nil {
		log.Println("Error parse TELE_API_ID")
	}

	AppConfig = &Config{
		TeleApiID:        apiId,
		TeleApiHash:      os.Getenv("TELE_API_HASH"),
		TeleChannelName:  os.Getenv("TELE_CHANNEL_NAME"),
		BinanceApiKey:    os.Getenv("BINANCE_API_KEY"),
		BinanceSecretKey: os.Getenv("BINANCE_SECRET_KEY"),
	}
}
