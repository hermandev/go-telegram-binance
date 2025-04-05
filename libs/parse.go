package libs

import (
	"go-telegram-binance/models"
	"regexp"
	"strconv"
	"strings"
)

// Fungsi untuk parsing angka dengan persen
func parsePercentage(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// Fungsi untuk parsing nilai angka umum
func parseNumber(s string) float64 {
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return val
}

// Fungsi untuk parsing teks ke JSON
func ParseTextToJSON(text string) (models.MarketData, error) {
	lines := strings.Split(text, "\n")
	var data models.MarketData
	reNumber := regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)
	reAsset := regexp.MustCompile(`\d+\.\s([A-Z0-9]+):\s?([-+]?[0-9]*\.?[0-9]+%)`)

	mode := ""

	for i := range lines {
		line := strings.TrimSpace(lines[i])

		if i == 0 {
			data.Date = line
		} else if strings.Contains(line, "Market Cap:") {
			data.MarketCap = strings.TrimPrefix(line, "Market Cap: ")
		} else if strings.Contains(line, "24h Volume:") {
			data.Volume24h = strings.TrimPrefix(line, "24h Volume: ")
		} else if strings.Contains(line, "BTC Dominance:") {
			data.BTCDominance = strings.TrimPrefix(line, "BTC Dominance: ")
		} else if strings.Contains(line, "ETH Dominance:") {
			data.ETHDominance = strings.TrimPrefix(line, "ETH Dominance: ")
		} else if strings.Contains(line, "BK® Health Standard:") {
			data.BKHealthStandard = parseNumber(reNumber.FindString(line))
		} else if strings.Contains(line, "BK® Sentiment:") {
			data.BKSentiment = parsePercentage(reNumber.FindString(line))
		} else if strings.Contains(line, "BINANCE ANALYSIS") {
			mode = "binance"
		} else if strings.Contains(line, "BINANCE FUTURES") {
			mode = "binance_futures"
		} else if strings.Contains(line, "Top Gainers") {
			for j := i + 1; j < len(lines); j++ {
				match := reAsset.FindStringSubmatch(lines[j])
				if match == nil {
					break
				}
				asset := models.AssetChange{
					Name:   match[1],
					Change: parsePercentage(match[2]),
				}
				switch mode {
				case "binance":
					data.BinanceAnalysis.TopGainers = append(data.BinanceAnalysis.TopGainers, asset)
				case "binance_futures":
					data.BinanceFutures.TopGainers = append(data.BinanceFutures.TopGainers, asset)
				}
			}
		} else if strings.Contains(line, "Top Losers") {
			for j := i + 1; j < len(lines); j++ {
				match := reAsset.FindStringSubmatch(lines[j])
				if match == nil {
					break
				}
				asset := models.AssetChange{
					Name:   match[1],
					Change: parsePercentage(match[2]),
				}
				switch mode {
				case "binance":
					data.BinanceAnalysis.TopLosers = append(data.BinanceAnalysis.TopLosers, asset)
				case "binance_futures":
					data.BinanceFutures.TopLosers = append(data.BinanceFutures.TopLosers, asset)
				}
			}
		} else if strings.Contains(line, "LARGEST VOLUME") {
			for j := i + 1; j < len(lines); j++ {
				parts := strings.Split(lines[j], " ($")
				if len(parts) != 2 {
					break
				}
				data.LargestVolume = append(data.LargestVolume, models.VolumeData{
					Name:  strings.TrimSpace(parts[0]),
					Value: "$" + strings.TrimSpace(parts[1]),
				})
			}
		} else if strings.Contains(line, "DAILY OUTLOOK") {
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) == "" {
					break
				}
				data.DailyOutlook += lines[j] + " "
			}
			data.DailyOutlook = strings.TrimSpace(data.DailyOutlook)
		}
	}

	return data, nil
}
