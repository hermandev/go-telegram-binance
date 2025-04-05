package main

import (
	"go-telegram-binance/config"
	"go-telegram-binance/libs"
	"log"
)

func main() {
	config.LoadConfig()

	config.InitBinanceClient(config.AppConfig.BinanceApiKey, config.AppConfig.BinanceSecretKey, false, true)
	telegram := libs.InitTelegramClient(config.AppConfig.TeleApiID, config.AppConfig.TeleApiHash)

	// Jalankan Telegram Client
	if err := telegram.Start(); err != nil {
		log.Fatal(err)
	}
}
