package binance_api

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket服务器每秒最多接受5个消息。消息包括:
// PING帧
// PONG帧
// JSON格式的消息, 比如订阅, 断开订阅.
// 如果用户发送的消息超过限制，连接会被断开连接。反复被断开连接的IP有可能被服务器屏蔽。
// 单个连接最多可以订阅1024个Streams。
// 每IP地址、每5分钟最多可以发送300次连接请求。

type Client struct {
	conn                   *websocket.Conn
	index                  int
	pool                   *WSPool
	connector              *WConnector
	lastPongTimestamp      atomic.Int64 // 秒级时间戳
	forceStop              atomic.Bool
	channelManager         *WSChannelManager
	hasFinishedInitConnect bool
}

// NewClient creates a new API client
func NewClient(pool *WSPool, index int) *Client {
	client := &Client{
		conn:                   nil,
		index:                  index,
		pool:                   pool,
		connector:              NewWConnector(),
		channelManager:         NewWSChannelManager(),
		hasFinishedInitConnect: false,
	}
	client.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
	client.forceStop.Store(false)
	return client
}

func (c *Client) Login() error {
	logger.Info("Login")
	if c.connector.LockIsLogined(c.conn) {
		logger.Error("WebSocket is already logged in") // 已连接, 不需要重复连接
		return nil
	}

	baseBinanceUrl := "wss://stream.binance.com:9443/stream?streams="

	conn, err := c.connector.LockDialWebSocket(baseBinanceUrl, nil)
	if err != nil {
		logger.Error("WebSocket dial error: %v", err)
		return err
	}
	if !c.hasFinishedInitConnect {
		c.hasFinishedInitConnect = true
		c.pool.FinishClientInitialized(c.index)
	}

	err = c.connector.LockLogin(conn)
	if err != nil {
		logger.Error("Error logging in: %v", err)
		c.connector.LockClose(conn)
		return err
	}
	c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
	c.conn = conn
	conn.SetPingHandler(func(appData string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
		if err != nil {
			logger.Error("Error logging in: %v", err)
			c.connector.LockClose(conn)
			return err
		}
		c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
		logger.Debug("Sent pong success :  %s", appData)
		return nil
	})
	go c.listenMsg(conn)
	return nil
}

func (c *Client) Start() {
	go func() {
		for {
			if c.index == 0 {
				break
			}
			// 等待前一个连接初始化完成
			if c.pool.IsClientInitialized(c.index - 1) {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}

		var conn *websocket.Conn = nil

		for {
			if c.forceStop.Load() {
				break
			} else if !c.connector.LockIsLogined(conn) {
				if err := c.Login(); err != nil {
					logger.Error("Login error: %v", err)
				} else {
					//重连的时候需要将已订阅列表中数据转移到pending列表中
					logger.Info("Client-%d connected and logged in", c.index)
					conn = c.conn
					// c.channelManager.reAddSubscibeList()
				}
			} else {
				if pastTime := utils.GetCurrentTimestampSec() - c.lastPongTimestamp.Load(); pastTime >= 40 {
					logger.Error("No ping received in the last 60 seconds, reconnecting...")
					c.connector.LockClose(c.conn)
					c.channelManager.reAddSubscibeList()
					c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
				} else {
					// 动态订阅逻辑在这里实现
					if err := c.sendPendingMessages(conn); err != nil {
						log.Printf("sendPendingMessages error: %v", err)
						c.connector.LockClose(c.conn)
					}
					time.Sleep(1 * time.Second)
				}

			}
		}
	}()
}

func isContentsSubscribeMessage(messageMap map[string]interface{}) (string, error) {

	dataMap, ok := messageMap["data"].(map[string]interface{})
	if !ok {
		fmt.Println("data field missing or not an object")
		return "", &json.UnmarshalTypeError{}
	}

	eVal, ok := dataMap["e"].(string)
	if !ok {
		logger.Debug("messageMap no string type")
		return eVal, &json.UnmarshalTypeError{}
	}

	// 判断 e 字段是否包含特定字符串
	if strings.Contains(eVal, "aggTrade") || strings.Contains(eVal, "kline") || strings.Contains(eVal, "24hrMiniTicker") {
		return eVal, nil
	}
	return "nil", nil
}

// 处理消息和 pong
func (c *Client) listenMsg(conn *websocket.Conn) {
	//是否需要判断 连接状态后续确认

	channelDataReceivedMap := make(map[string]bool)

	hasReceivedData := func(stream string, msg *binance_define.WSSinglePushMsg) bool {
		_, exists := channelDataReceivedMap[stream]
		if !exists {
			channelDataReceivedMap[stream] = true
			msg.IsFirst = true
		}
		c.pool.AddPushDataToQueue(msg)
		return exists
	}

	for {
		message, err := c.connector.LockReadPushMessage(conn)
		if err != nil {
			logger.Warn("WebSocket read error: %v", err)
			break
		}
		// logger.Debug("Received message: %s", string(message))
		// {"result":null,"id":1763534632712956000}
		// {"stream":"ethusdt@miniTicker","data":{"e":"24hrMiniTicker","E":1763534633053,"s":"ETHUSDT","c":"3040.65000000",
		// "o":"2976.95000000","h":"3169.95000000","l":"2973.53000000","v":"620337.05010000","q":"1909901787.54333500"}}

		var messageJson map[string]interface{}
		parseErr := json.Unmarshal([]byte(string(message)), &messageJson)
		if parseErr != nil {
			logger.Error("message format failed")
		}
		//订阅和取消订阅消息
		if messageJson["result"] == nil && messageJson["id"] != nil {
			logger.Debug("订阅和取消订阅消息 %s", message)
			continue
		}

		if messageJson["stream"] != nil && messageJson["data"] != nil {
			stream, ok := messageJson["stream"].(string)
			if ok {
				// 不是 string 类型
				isExist := c.channelManager.IsInSubscirbe(stream)
				if isExist {
					evt, err := isContentsSubscribeMessage(messageJson)
					if err != nil {
						logger.Error("Failed to marshal messageJson: %v", messageJson)
						// logger.Error("Failed to marshal messageJson: %v", err)
						continue
					}
					dataField, ok := messageJson["data"]
					if !ok {
						fmt.Println("data field not found")
						continue
					}

					payload, err := json.Marshal(dataField)
					if err != nil {
						fmt.Println("marshal data failed:", err)
						continue
					}
					// logger.Debug("evt %s", evt)
					if evt == "aggTrade" {
						var trade binance_define.BinanceAggTrade
						if err := json.Unmarshal(payload, &trade); err != nil {
							logger.Error("Failed to unmarshal aggTrade payload: %v, payload: %s", err, string(payload))
							continue
						}
						msg := &binance_define.WSSinglePushMsg{
							IsFirst:   false,
							EventType: evt,
							Data:      trade,
						}
						hasReceivedData(stream, msg)

					} else if evt == "kline" {
						var kline binance_define.KlineMessage
						if err := json.Unmarshal(payload, &kline); err != nil {
							logger.Error("Failed to unmarshal aggTrade payload: %v, payload: %s", err, string(payload))
							continue
						}

						msg := &binance_define.WSSinglePushMsg{
							IsFirst:   false,
							EventType: evt,
							Data:      kline,
						}
						hasReceivedData(stream, msg)

					} else if evt == "24hrMiniTicker" {
						var miniTicker binance_define.Binance24hrMiniTicker
						if err := json.Unmarshal(payload, &miniTicker); err != nil {
							logger.Error("Failed to unmarshal aggTrade payload: %v, payload: %s", err, string(payload))
							continue
						}
						msg := &binance_define.WSSinglePushMsg{
							IsFirst:   false,
							EventType: evt,
							Data:      miniTicker,
						}
						hasReceivedData(stream, msg)
					} else {
						logger.Debug("Received other event type: %s, message: %s", evt, string(message))
					}
				}
			}

		}
	}
}

// func (c *Client) onPullHistoryMsg(event string, pushedMsg *binance_define.WSSinglePushMsg) {
// newSinglePushMsg := func(arg *binance_define.WSSubscribeArg, data interface{}) *binance_define.WSSinglePushMsg {
// 	return &binance_define.WSSinglePushMsg{
// 		IsFirst: isFirst,
// 		Arg:     arg,
// 		Data:    data,
// 	}
// }

// 将推送的数据加入到处理队列中
// switch pushedMsg.Arg.Channel {
// case "price":
// 	var priceData []*okx_define.WSPriceData
// 	json.Unmarshal(pushedMsg.Data, &priceData)
// 	for i := 0; i < len(priceData); i++ {
// 		c.pool.AddPushDataToQueue(newSinglePushMsg(&pushedMsg.Arg, priceData[i]))
// 	}

// case "dex-token-candle1s":
// 	var candleData [][okx_define.WSCandleDataFieldTotalCount]string
// 	json.Unmarshal(pushedMsg.Data, &candleData)
// 	for i := 0; i < len(candleData); i++ {
// 		c.pool.AddPushDataToQueue(newSinglePushMsg(&pushedMsg.Arg, candleData[i]))
// 	}
// case "trades":
// 	var tradeData []*okx_define.WSTradeData
// 	json.Unmarshal(pushedMsg.Data, &tradeData)
// 	for i := 0; i < len(tradeData); i++ {
// 		c.pool.AddPushDataToQueue(newSinglePushMsg(&pushedMsg.Arg, tradeData[i]))
// 	}
// }
// }

// 处理总订阅数所形成的json 超过4096kb的情况
func getAllowRangeSubscirbeMessage(values []string) ([]byte, int) {
	msg := binance_define.SubscribeMsg{
		Method: "SUBSCRIBE",
		Params: values,
		ID:     int(utils.GetCurrentTimestampNano()),
	}

	data, _ := json.Marshal(msg)
	// 如果序列化后长度 <= 4000，满足条件，直接返回
	if len(data) <= 4000 {
		return data, len(values) - 1
	}
	return getAllowRangeSubscirbeMessage(values[:len(values)-1])

}

// TODO 后续看是否需要
func (c *Client) sendPendingMessages(conn *websocket.Conn) error {

	clientSubscribelist := c.channelManager.subscribePendingChannelArgList.All()

	if len(clientSubscribelist) > 0 {
		values := []string{} // 创建空切片
		for _, subscribe := range clientSubscribelist {
			values = append(values, subscribe.TokenPair+"@"+subscribe.Channel)
		}
		logger.Info("Client-%d 需要订阅的数据 %s", c.index, values)
		for {
			data, index := getAllowRangeSubscirbeMessage(values)
			logger.Debug("index %d", index)
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Subscribe error: %v", err)
			}
			for _, v := range values[:index+1] {
				c.channelManager.subscribedChannels.Set(v, true)
			}
			logger.Debug("index %d", len(values))
			c.channelManager.subscribePendingChannelArgList.RemoveRange(0, index+1)
			if index == (len(values) - 1) {
				break
			}
			values = values[index:]
		}
	}

	clientUnsubscribelist := c.channelManager.unsubscribePendingChannelArgList.All()
	if len(clientUnsubscribelist) > 0 {
		values := []string{} // 创建空切片
		for _, subscribe := range clientUnsubscribelist {
			values = append(values, subscribe.TokenPair+"@"+subscribe.Channel)
		}
		logger.Info("Client-%d 需要订阅取消订阅的数据 %s", c.index, values)
		subMsg := binance_define.SubscribeMsg{
			Method: "UNSUBSCRIBE",
			Params: values,
			ID:     int(utils.GetCurrentTimestampNano()),
		}
		data, _ := json.Marshal(subMsg)

		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("UNSubscribe error: %v", err)
		}
		for _, v := range values {
			c.channelManager.subscribedChannels.Delete(v)
		}
		c.channelManager.unsubscribePendingChannelArgList.Clear()

	}
	return nil
}

func (cm *Client) AddChannelSubscribe(arg *binance_define.RedisChannelArg) {
	cm.channelManager.AddChannelSubscribe(arg)

}

func (cm *Client) BatchAddChannelSubscribe(args *utils.List[*binance_define.RedisChannelArg]) {
	cm.channelManager.BatchAddChannelSubscribe(args)

}

func (cm *Client) AddChannelUnsubscribe(arg *binance_define.RedisChannelArg) {
	cm.channelManager.AddChannelUnsubscribe(arg)
}

// 给上层调用的关闭接口
func (c *Client) Close() error {
	c.forceStop.Store(true)
	return c.connector.LockClose(c.conn)
}
