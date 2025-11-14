package okx

import (
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// 定义 API 凭证
const (
	API_KEY    = "a6184391-2a7d-4098-87df-68cc5f783dcd" // Replace with your API key
	SECRET_KEY = "FAE81C0879424978074EAD111BFEB2C4"     // Replace with your secret key
	PASSPHRASE = "Xiaowentian35!"                       // Replace with your passphrase

	// API_KEY    = "3b730d2c-791a-4477-8200-eb2f521526f9" // Replace with your API key
	// SECRET_KEY = "2836C7A0B8B92717CBA0A2413DABA34E"     // Replace with your secret key
	// PASSPHRASE = "Lightdex0923@"                        // Replace with your passphrase
	HOST_NAME = "https://web3.okx.com"
)

func PreHash(timestamp, method, requestPath string, params map[string]interface{}) string {
	// 根据字符串和参数创建预签名
	queryString := ""
	if method == "GET" && len(params) > 0 {
		values := make([]string, 0, len(params))
		for k, v := range params {
			values = append(values, k+"="+url.QueryEscape(fmt.Sprintf("%v", v)))
		}
		queryString = "?" + strings.Join(values, "&")
	}
	if method == "POST" && len(params) > 0 {
		queryBytes, _ := json.Marshal(params)
		queryString = string(queryBytes)
	}
	return timestamp + method + requestPath + queryString
}

func Sign(message, secretKey string) string {
	// 使用 HMAC-SHA256 对预签名字符串进行签名
	sign := utils.HMACSHA256([]byte(secretKey), []byte(message))
	return base64.StdEncoding.EncodeToString(sign)
}

func CreateSignature(method, requestPath string, params map[string]interface{}) (signature, timestamp string) {
	// 获取 ISO 8601 格式时间戳
	timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	// 生成签名
	message := PreHash(timestamp, method, requestPath, params)
	signature = Sign(message, SECRET_KEY)
	return signature, timestamp
}

func CreateSignature2(method, requestPath string, params map[string]interface{}) (signature, timestamp string) {
	// 获取 ISO 8601 格式时间戳
	// timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	// 获取秒级时间戳
	timestamp = fmt.Sprintf("%d", time.Now().Unix())

	// 生成签名
	message := PreHash(timestamp, method, requestPath, params)
	signature = Sign(message, SECRET_KEY)
	return signature, timestamp
}

func SendGetRequest(ctx context.Context, requestPath string, params map[string]interface{}) (httpCode int, body []byte, err error) {
	// 生成签名
	signature, timestamp := CreateSignature("GET", requestPath, params)

	// 生成请求头
	headers := map[string]string{
		"OK-ACCESS-KEY":        API_KEY,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  timestamp,
		"OK-ACCESS-PASSPHRASE": PASSPHRASE,
	}

	queryString := ""
	if len(params) > 0 {
		values := make([]string, 0, len(params))
		for k, v := range params {
			values = append(values, k+"="+url.QueryEscape(fmt.Sprintf("%v", v)))
		}
		queryString = "?" + strings.Join(values, "&")
	}

	url := HOST_NAME + requestPath + queryString

	logger.Info("GET URL: %s", url)

	return utils.HttpRequest(ctx, "GET", url, nil, headers)
}

func SendPostRequest(ctx context.Context, requestPath string, params map[string]interface{}) (httpCode int, resp map[string]interface{}, err error) {
	// 生成签名
	signature, timestamp := CreateSignature("POST", requestPath, params)

	// 生成请求头
	headers := map[string]string{
		"OK-ACCESS-KEY":        API_KEY,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  timestamp,
		"OK-ACCESS-PASSPHRASE": PASSPHRASE,
	}

	url := HOST_NAME + requestPath
	return utils.HttpPostWithJsonMap(ctx, url, params, headers)
}

// // GET 请求示例
// const getRequestPath = '/api/v6/dex/aggregator/quote';
// const getParams = {
//   'chainId': 42161,
//   'amount': 1000000000000,
//   'toTokenAddress': '0xff970a61a04b1ca14834a43f5de4533ebddb5cc8',
//   'fromTokenAddress': '0x82aF49447D8a07e3bd95BD0d56f35241523fBab1'
// };
// sendGetRequest(getRequestPath, getParams);

// // POST 请求示例
// const postRequestPath = '/api/v5/mktplace/nft/ordinals/listings';
// const postParams = {
//   'slug': 'sats'
// };
// sendPostRequest(postRequestPath, postParams);
