package models

// Struktur JSON
type MarketData struct {
	Date             string  `json:"date"`
	MarketCap        string  `json:"market_cap"`
	Volume24h        string  `json:"volume_24h"`
	BTCDominance     string  `json:"btc_dominance"`
	ETHDominance     string  `json:"eth_dominance"`
	BKHealthStandard float64 `json:"bk_health_standard"`
	BKSentiment      float64 `json:"bk_sentiment"`
	BinanceAnalysis  struct {
		TopGainers []AssetChange `json:"top_gainers"`
		TopLosers  []AssetChange `json:"top_losers"`
	} `json:"binance_analysis"`
	BinanceFutures struct {
		TopGainers []AssetChange `json:"top_gainers"`
		TopLosers  []AssetChange `json:"top_losers"`
	} `json:"binance_futures"`
	LargestVolume []VolumeData `json:"largest_volume"`
	DailyOutlook  string       `json:"daily_outlook"`
}

// Struktur untuk aset perubahan harga (Top Gainers / Losers)
type AssetChange struct {
	Name   string  `json:"name"`
	Change float64 `json:"change"`
}

// Struktur untuk volume terbesar
type VolumeData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
