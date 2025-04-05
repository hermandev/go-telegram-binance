package config

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

type BinanceClient struct {
	client *futures.Client
}

var (
	instance *BinanceClient
	once     sync.Once
)

func InitBinanceClient(apiKey, secretKey string, debug, dev bool) *BinanceClient {
	once.Do(func() {
		instance = &BinanceClient{
			client: binance.NewFuturesClient(apiKey, secretKey),
		}
	})

	if dev {
		instance.client.BaseURL = "https://testnet.binancefuture.com"
	}
	instance.client.Debug = debug
	return instance
}

func GetInstance() *BinanceClient {
	if instance == nil {
		panic("BinanceClient belum diinisialisasi! Panggil InitBinanceClient terlebih dahulu.")
	}
	return instance
}

func (b *BinanceClient) CreateOrder(symbol string, side futures.SideType) error {
	b.client.BaseURL = "https://testnet.binancefuture.com"
	log.Printf("Proses Order %s = %s", symbol, side)

	// Dapatkan harga pasar saat ini
	price, err := b.client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Printf("Error fetching price for %s: %v", symbol, err)
		return err
	}

	// Asumsikan hanya ada satu harga yang dikembalikan
	currentPrice, err := strconv.ParseFloat(price[0].Price, 64)
	if err != nil {
		log.Printf("Error Get currentPrice  %v", err)
		return err
	}

	// Hitung jumlah koin berdasarkan jumlah USDT
	qty := 20 / currentPrice

	fmt.Sprintln("Qty = ", math.Round(qty))
	// return nil

	// Periksa dan set mode margin menjadi terisolasi jika diperlukan
	marginType, err := b.client.NewGetPositionRiskService().
		Symbol(symbol).
		Do(context.Background())
	if err != nil {
		log.Printf("Error fetching margin type: %v", err)
		return err
	}

	if marginType[0].MarginType != "isolated" {
		err = b.client.NewChangeMarginTypeService().
			Symbol(symbol).
			MarginType(futures.MarginTypeIsolated).
			Do(context.Background())
		if err != nil {
			log.Printf("Error setting margin type: %v", err)
			return err
		}
	}

	// Set leverage menjadi 10x
	_, err = b.client.NewChangeLeverageService().
		Symbol(symbol).
		Leverage(10).
		Do(context.Background())
	if err != nil {
		log.Printf("Error setting leverage: %v", err)
		return err
	}

	// Buat order market beli dengan BTC
	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(futures.OrderTypeMarket).
		PositionSide(futures.PositionSideTypeBoth).
		Quantity(fmt.Sprint(math.Round(qty))). // Menggunakan jumlah BTC yang diformat
		Do(context.Background())
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return err
	} else {
		log.Println("Order Success :", order)
	}
	return nil
}
