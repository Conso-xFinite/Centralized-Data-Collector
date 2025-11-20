package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Binance API endpoints
const (
	baseURL       = "https://api.binance.com"
	klinesPath    = "/api/v3/uiKlines"
	aggTradesPath = "/api/v3/aggTrades"
)

// KlineResponse 单条Kline数据
type KlineResponse struct {
	OpenTime                 int64
	Open, High, Low, Close   string
	Volume                   string
	CloseTime                int64
	QuoteAssetVolume         string
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
	Ignore                   string
}

// AggTradeResponse 单条aggTrade数据
type AggTradeResponse struct {
	AggTradeID   int64  `json:"a"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	FirstID      int64  `json:"f"`
	LastID       int64  `json:"l"`
	Timestamp    int64  `json:"T"`
	IsBuyerMaker bool   `json:"m"`
	Ignore       bool   `json:"M"`
}

// FetchKlines 拉取Kline 1m数据
func FetchKlines(symbol string, interval string, limit int, startTime, endTime int64) ([]KlineResponse, error) {
	u, _ := url.Parse(baseURL + klinesPath)
	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("interval", interval)
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if startTime > 0 {
		q.Set("startTime", fmt.Sprintf("%d", startTime))
	}
	if endTime > 0 {
		q.Set("endTime", fmt.Sprintf("%d", endTime))
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	var raw [][]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make([]KlineResponse, 0, len(raw))
	for _, item := range raw {
		result = append(result, KlineResponse{
			OpenTime:                 int64(item[0].(float64)),
			Open:                     item[1].(string),
			High:                     item[2].(string),
			Low:                      item[3].(string),
			Close:                    item[4].(string),
			Volume:                   item[5].(string),
			CloseTime:                int64(item[6].(float64)),
			QuoteAssetVolume:         item[7].(string),
			NumberOfTrades:           int64(item[8].(float64)),
			TakerBuyBaseAssetVolume:  item[9].(string),
			TakerBuyQuoteAssetVolume: item[10].(string),
			Ignore:                   item[11].(string),
		})
	}
	return result, nil
}

// FetchAggTrades 拉取aggTrade数据
func FetchAggTrades(symbol string, limit int, startTime, endTime int64) ([]AggTradeResponse, error) {
	u, _ := url.Parse(baseURL + aggTradesPath)
	q := u.Query()
	q.Set("symbol", symbol)
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if startTime > 0 {
		q.Set("startTime", fmt.Sprintf("%d", startTime))
	}
	if endTime > 0 {
		q.Set("endTime", fmt.Sprintf("%d", endTime))
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var trades []AggTradeResponse
	if err := json.Unmarshal(data, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}
