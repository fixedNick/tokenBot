package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"main/config"
	"main/core"
	"main/utils"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

var (
	GetMarketItem_Counter = 0
	GetInventory_Counter  = 0
	BuyItem_Counter       = 0
	SellItem_Counter      = 0
	GetBalance_Counter    = 0
)

type Items []*core.Item
type InventoryItems []*core.InventoryItem

var CURRENT_MARKET_TYPE = 0
var MAX_MARKET_TYPE = 5

func GetMarketItem() *core.Item {

	minDelay := 7213
	maxDelay := 9012

	randTick := (rand.Intn(maxDelay-minDelay+1) + minDelay)
	time.Sleep(time.Millisecond * time.Duration(randTick))

	var wormType string
	switch CURRENT_MARKET_TYPE {
	case 0:
		wormType = "common"
	case 1:
		wormType = "uncommon"
	case 2:
		wormType = "rare"
	case 3:
		wormType = "epic"
	case 4:
		wormType = "legendary"
	default:
		wormType = "rare"
	}

	slog.Info(fmt.Sprintf("[%.2fms] Getting market items from [%s]...", float64(randTick)/1000.0, strings.ToUpper(wormType)))

	url := fmt.Sprintf("https://alb.seeddao.org/api/v1/market/v2?market_type=worm&worm_type=%s&sort_by_price=ASC&sort_by_updated_at=&page=1", wormType)
	method := "GET"

	req := NewRequest(url, method, nil)

	resp, err := http.DefaultClient.Do(req)
	GetMarketItem_Counter++
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var bodyReader io.Reader

	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(resp.Body)
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
	} else {
		bodyReader = resp.Body
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		slog.Warn(fmt.Sprintf("GetMarketItem WARNING: Status Code: %d. Body: %s", resp.StatusCode, string(body)))

		if resp.StatusCode == 401 {
			panic("Authorization token has exprired... Program finished")
		}

		if resp.StatusCode == 400 && strings.Contains(string(body), "too many requests") {

			slog.Warn(fmt.Sprintf("Request counts:\n\rGetMarketItem:\t%d\n\rGetInventory:\t%d\n\rSellItem:\t%d\n\rBuyItem:\t%d\n\rGetBalance:\t%d", GetMarketItem_Counter, GetInventory_Counter, SellItem_Counter, BuyItem_Counter, GetBalance_Counter))
			slog.Warn("Too many requests. Waiting 5 minutes...")
			time.Sleep(time.Minute * 5)
			return nil
		}

		panic("Unexpected status code")
	}

	rawJson := make(map[string]interface{})
	if err = json.Unmarshal(body, &rawJson); err != nil {
		fmt.Println("body: ", string(body))
		fmt.Println("Unmarshal error:")
		panic(err)
	}

	dataJson, ok := rawJson["data"]
	if !ok {
		panic("Response json doesn't contain 'data' field")
	}
	inDataJson, _ := json.Marshal(dataJson)
	var rawData map[string]interface{}
	json.Unmarshal(inDataJson, &rawData)
	itemsJson, _ := json.Marshal(rawData["items"])

	var items Items
	if err = json.Unmarshal(itemsJson, &items); err != nil {
		panic(err)
	}

	CURRENT_MARKET_TYPE = (CURRENT_MARKET_TYPE + 1) % MAX_MARKET_TYPE

	if len(items) == 0 {
		slog.Warn("MARKET WITH NO ITEMS!!!! POSSIBLY BANNED!!!!")
		panic("Antiban")
	}

	var bestPriceItem *core.Item
	for _, item := range items {
		if bestPriceItem == nil {
			bestPriceItem = item
			continue
		}

		if utils.ToFloat(bestPriceItem.PriceGross) > utils.ToFloat(item.PriceGross) {
			bestPriceItem = item
		}
	}

	return bestPriceItem
}

func BuyItem(marketId string) bool {
	url := "https://alb.seeddao.org/api/v1/market-item/buy"
	method := "POST"

	req := NewRequest(url, method, bytes.NewBuffer([]byte(fmt.Sprintf("{\"market_id\":\"%s\"}", marketId))))

	resp, err := http.DefaultClient.Do(req)
	BuyItem_Counter++
	if err != nil {
		slog.Error(fmt.Sprintf("Error buying item: %s", err.Error()))
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {

		var bodyReader io.Reader

		if resp.Header.Get("Content-Encoding") == "br" {
			bodyReader = brotli.NewReader(resp.Body)
		} else if resp.Header.Get("Content-Encoding") == "gzip" {
			bodyReader, err = gzip.NewReader(resp.Body)
			if err != nil {
				panic(err)
			}
		} else {
			bodyReader = resp.Body
		}

		body, err := io.ReadAll(bodyReader)
		if err != nil {
			panic(err)
		}

		slog.Warn(fmt.Sprintf("[WARNING] Buying item failed. Status code: %d. Body: %s", resp.StatusCode, string(body)))
	}

	return resp.StatusCode == 200
}

func GetInventory(page int) InventoryItems {
	url := fmt.Sprintf("https://alb.seeddao.org/api/v1/worms/me?page=%d", page)
	method := "GET"

	req := NewRequest(url, method, nil)

	resp, err := http.DefaultClient.Do(req)
	GetInventory_Counter++
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Status code is not 200, it is %d", resp.StatusCode))
	}

	var bodyReader io.Reader

	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(resp.Body)
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
	} else {
		bodyReader = resp.Body
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	var resultItems []*core.InventoryItem

	itemsCount, ok := result["data"].(map[string]interface{})["total"].(float64)
	if !ok {
		panic("Items count not in raw json")
	}

	data, ok := result["data"].(map[string]interface{})["items"]
	if !ok {
		panic("Items not in raw json")
	}
	bytesData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(bytesData, &resultItems); err != nil {
		panic(err)
	}

	// Delay before next operations...
	minDelay := 1227
	maxDelay := 2091

	randTick := (rand.Intn(maxDelay-minDelay+1) + minDelay)
	time.Sleep(time.Millisecond * time.Duration(randTick))
	slog.Info(fmt.Sprintf("[%.2fs] Fetching inventory items. Total items: %d. Current Page: %d", (float64(randTick) / 1000.0), int(itemsCount), page))

	pageSize := 20
	totalPages := math.Ceil(float64(itemsCount) / float64(pageSize))
	if int(totalPages) > page {
		resultItems = append(resultItems, GetInventory(page+1)...)
	}

	return InventoryItems(resultItems)
}

func SellItem(item *core.InventoryItem, price float64) bool {
	url := "https://alb.seeddao.org/api/v1/market-item/add"
	method := "POST"
	req := NewRequest(url, method, bytes.NewBuffer([]byte(fmt.Sprintf("{\"worm_id\": \"%s\", \"price\": %d}", item.Id, utils.ToBigInt(price)))))

	resp, err := http.DefaultClient.Do(req)
	SellItem_Counter++
	if err != nil {
		slog.Error(fmt.Sprintf("Error selling item: %s", err.Error()))
		return false
	}

	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func GetBalance() float64 {
	url := "https://alb.seeddao.org/api/v1/profile/balance"
	method := "GET"

	req := NewRequest(url, method, nil)

	resp, err := http.DefaultClient.Do(req)
	GetBalance_Counter++
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Status code is not 200, it is %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	if balance, ok := result["data"].(float64); ok {
		slog.Info("Current Balance: " + fmt.Sprintf("%f", utils.ToFloat(int64(balance))))
		return utils.ToFloat(int64(balance))
	}
	panic(fmt.Sprintf("Cannot Unmarshal balance from response. Received response:\n\r%s", body))
}

func NewRequest(url string, method string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://cf.seeddao.org")
	req.Header.Set("Referer", "https://cf.seeddao.org/")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="128", "Not;A=Brand";v="24", "Opera GX";v="114"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 OPR/114.0.0.0")

	// Set custom header
	req.Header.Set("Telegram-Data", CONFIG.TelegramData)

	return req
}

func buyItemsUntilLimit(balance float64, minBalance float64) float64 {
	slog.Info("Start bying items.")
	for {

		if balance < minBalance {
			slog.Info("Buying items finished. Current Balance: " + fmt.Sprintf("%f", balance))
			return balance
		}
		item := GetMarketItem()
		if item == nil {
			randTick := (rand.Intn(1574-923+1) + 923)
			time.Sleep(time.Millisecond * time.Duration(randTick))
			continue
		}

		var price float64 = 0
		switch item.Type() {
		case core.ItemTypeCommon:
			price = CONFIG.Buy.Common
		case core.ItemTypeUncommon:
			price = CONFIG.Buy.Uncommon
		case core.ItemTypeRare:
			price = CONFIG.Buy.Rare
		case core.ItemTypeEpic:
			price = CONFIG.Buy.Epic
		case core.ItemTypeLegendary:
			price = CONFIG.Buy.Legendary
		}

		itemPrice := utils.ToFloat(item.PriceGross)

		if itemPrice <= price && itemPrice <= balance {

			status := BuyItem(item.Id)

			randTick := (rand.Intn(1574-923+1) + 923)
			time.Sleep(time.Millisecond * time.Duration(randTick))

			if status {
				balance -= itemPrice
				slog.Info(fmt.Sprintf("+++ %s [%.2f Gems] | Balance: %.2f Gems", strings.ToUpper(item.WormType), itemPrice, balance))
				continue
			}
		}
	}
}

func waitPricebleBalance(minBalance float64) float64 {
	for {
		balance := GetBalance()
		time.Sleep(time.Millisecond * 1733)
		if balance > minBalance {
			return balance
		}

		slog.Info("Waiting for priceble balance. Current Balance: " + fmt.Sprintf("%f", balance))
	}
}

func sellAllItems() {
	slog.Info("Start selling items...")
	items := GetInventory(1)

	slog.Info(fmt.Sprintf("Items count: %d", len(items)))

	for _, item := range items {
		if item.OnMarket {
			continue
		}

		var price float64 = 0
		switch item.Type() {
		case core.ItemTypeCommon:
			price = CONFIG.Sell.Common
		case core.ItemTypeUncommon:
			price = CONFIG.Sell.Uncommon
		case core.ItemTypeRare:
			price = CONFIG.Sell.Rare
		case core.ItemTypeEpic:
			price = CONFIG.Sell.Epic
		case core.ItemTypeLegendary:
			price = CONFIG.Sell.Legendary
		}

		status := SellItem(item, price)
		time.Sleep(time.Millisecond * 1452)
		if status {
			slog.Info(fmt.Sprintf("--- %s [%.2f Gems] | id: %s", strings.ToUpper(item.WormType), price, item.Id))
			continue
		}
		slog.Info("Failed to sell item [id] " + item.Id + " [type] " + item.WormType)
	}
}

func run() {

	if len(os.Args) > 1 {
		if os.Args[1] == "sell" {
			sellAllItems()
		}
	}
	balance := GetBalance()
	minBalance := CONFIG.Buy.Epic

	for {

		if balance < minBalance {
			sellAllItems()
			balance = waitPricebleBalance(minBalance)
			continue
		}

		balance = buyItemsUntilLimit(balance, minBalance)

	}
}

var CONFIG *config.Config

func main() {
	CONFIG = config.NewConfig("config.yaml")
	run()

}
