package binance_api

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"encoding/json"
	"fmt"
	"net/url"
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

// Client represents the OKX API client
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

// ✅ 这里是你手动维护的订阅代币对列表（小写、USDT 结尾）
var subscribeSymbols = []string{
	"btcusdt", "ethusdt", "bnbusdt", "adausdt", "solusdt", "xrpusdt",

	// ...你可以继续补到100个
}

// 构造 combined stream URL
func buildCombinedURL(symbols []string) string {
	parts := make([]string, 0, len(symbols))
	for _, s := range symbols {
		parts = append(parts, fmt.Sprintf("%s@aggTrade", s))
	}
	streams := strings.Join(parts, "/")
	u := url.URL{
		Scheme: "wss",
		Host:   "stream.binance.com:9443",
		Path:   "/stream",
		RawQuery: url.Values{
			"streams": {streams},
		}.Encode(),
	}
	return u.String()
}

func (c *Client) Login() error {

	if c.connector.LockIsLogined(c.conn) {
		logger.Error("WebSocket is already logged in") // 已连接, 不需要重复连接
		return nil
	}

	var streamUrl string
	if c.index == 0 {
		streamUrl = "wss://stream.binance.com:9443/stream?streams=btcusdt%40aggTrade"
	} else if c.index == 1 {
		streamUrl = "wss://stream.binance.com:9443/stream?streams=btcusdt%40kline_1m"
	} else if c.index == 2 {
		streamUrl = "wss://stream.binance.com:9443/stream?streams=btcusdt%40avgPrice"
	}

	conn, resq, err := websocket.DefaultDialer.Dial(streamUrl, nil)
	if err != nil {
		logger.Error("WebSocket dial failed: %v, http response: %+v", err)
		logger.Error("WebSocket dial failed: %v, http response: %+v", resq)
	}
	if err != nil {
		// logError("[%s] dial error: %v", name, err)
		logger.Error("WebSocket dial error: %v", err)
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

	// set pong handler to update lastPongTimestamp
	conn.SetPongHandler(func(appData string) error {
		c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
		return nil
	})

	c.conn = conn

	c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
	go c.listenMsg(conn)
	logger.Info("WebSocket connected and logged in successfully, index: %d", c.index)
	return nil
}

func (c *Client) Start() {
	// wg.Add(1)
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
		// if c.index > 0 {
		// 	time.Sleep(time.Second * time.Duration(c.index)) // staggered start
		// }
		var conn *websocket.Conn = nil
		for {
			if c.forceStop.Load() {
				break
			} else if !c.connector.LockIsLogined(conn) {
				logger.Debug("Login: %v", !c.connector.LockIsLogined(conn))
				if err := c.Login(); err != nil {
					logger.Error("Login error: %v", err)
				} else {
					// c.channelManager.RequeueAllSubscribedChannels()
					conn = c.conn
				}
			} else {
				// if pastTime := utils.GetCurrentTimestampSec() - c.lastPongTimestamp.Load(); pastTime >= 30 {
				// 	logger.Error("No pong received in the last 30 seconds, reconnecting...")
				// 	c.connector.LockClose(c.conn)
				// } else if pastTime >= 10 {
				// 	if err := c.connector.LockWritePingMessage(conn); err != nil {
				// 		log.Printf("Ping error: %v", err)
				// 		c.connector.LockClose(c.conn)
				// 	}
				// }
				//  else if
				// err := c.sendPendingMessages(conn); err != nil {
				// 	log.Printf("sendPendingMessages error: %v", err)
				// 	c.connector.LockClose(c.conn)
				// }
				// time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (c *Client) onPongResp() {
	c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())
}

// func (c *Client) onSubscribeResp(resp *binance_define.WSSubscribeResp) {
// 	if resp.Event == "subscribe" {
// 		if resp.IsSuccess() {
// 			logger.Info("Subscribed to channel successfully: %v", resp.Arg)
// 		} else {
// 			// TODO: 这里要不要做一些错误处理?
// 			logger.Error("Failed to subscribe to channel: %s", resp.Msg)
// 		}
// 	} else { // "unsubscribe"
// 		if resp.IsSuccess() {
// 			logger.Info("Unsubscribed from channel successfully: %v", resp.Arg)
// 		} else {
// 			// TODO: 这里要不要做一些错误处理?
// 			logger.Error("Failed to unsubscribe from channel: %s", resp.Msg)
// 		}
// 	}
// }

// func (c *Client) onPushedMsg(pushedMsg *binance_define.WSSubscribeResp, isFirst bool) {
// 	newSinglePushMsg := func(arg *binance_define.WSSubscribeArg, data interface{}) *binance_define.WSSinglePushMsg {
// 		return &binance_define.WSSinglePushMsg{
// 			IsFirst: isFirst,
// 			Arg:     arg,
// 			Data:    data,
// 		}
// 	}

// 	// 将推送的数据加入到处理队列中
// 	switch pushedMsg.Arg.Channel {
// 	case "price":
// 		var priceData []*binance_define.WSPriceData
// 		json.Unmarshal(pushedMsg.Data, &priceData)
// 		for i := 0; i < len(priceData); i++ {
// 			c.pool.AddPushDataToQueue(newSinglePushMsg(&pushedMsg.Arg, priceData[i]))
// 		}

// 	case "dex-token-candle1s":
// 		var candleData [][binance_define.WSCandleDataFieldTotalCount]string
// 		json.Unmarshal(pushedMsg.Data, &candleData)
// 		for i := 0; i < len(candleData); i++ {
// 			c.pool.AddPushDataToQueue(newSinglePushMsg(&pushedMsg.Arg, candleData[i]))
// 		}
// 	case "trades":
// 		var tradeData []*binance_define.WSTradeData
// 		json.Unmarshal(pushedMsg.Data, &tradeData)
// 		for i := 0; i < len(tradeData); i++ {
// 			c.pool.AddPushDataToQueue(newSinglePushMsg(&pushedMsg.Arg, tradeData[i]))
// 		}
// 	}
// }

// 处理消息和 pong
func (c *Client) listenMsg(conn *websocket.Conn) {

	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	// 记录首次收到某个 channel 的标记（与之前行为兼容）
	// channelDataReceivedMap := make(map[string]bool)
	// hasReceivedData := func(arg *binance_define.WSSubscribeArg) bool {
	// 	key := fmt.Sprintf("%s:%s:%s", arg.Channel, arg.ChainIndex, arg.TokenContractAddress)
	// 	_, exists := channelDataReceivedMap[key]
	// 	if !exists {
	// 		channelDataReceivedMap[key] = true
	// 	}
	// 	return exists
	// }

	for {
		if c.forceStop.Load() {
			logger.Info("force stop requested, breaking listen loop")
			break
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Warn("WebSocket read error: %v", err)
			break
		}

		// 更新最后一次活跃时间（收到任意消息视作活跃）
		c.lastPongTimestamp.Store(utils.GetCurrentTimestampSec())

		// 先尝试解析为通用 map 以便快速判断
		var generic map[string]interface{}
		if err := json.Unmarshal(message, &generic); err != nil {
			logger.Error("Failed to unmarshal message: %v, raw: %s", err, message)
			continue
		}

		// 辅助函数：从一个 interface{}（可能是 map[string]interface{} 或 json.RawMessage）
		// 安全提取出包含事件的 map[string]interface{}
		var dataMap map[string]interface{}
		if d, ok := generic["data"]; ok && d != nil {
			// combined stream: top-level has "data" object
			if dm, ok := d.(map[string]interface{}); ok {
				dataMap = dm
			} else {
				// 有时候 data 可能是 json.RawMessage ——再反序列化一次
				var raw json.RawMessage
				b, _ := json.Marshal(d)
				if json.Unmarshal(b, &raw) == nil {
					var dm2 map[string]interface{}
					if json.Unmarshal(b, &dm2) == nil {
						dataMap = dm2
					}
				}
			}
		} else {
			// single stream: top-level 就包含事件字段
			dataMap = generic
		}

		// 如果是订阅确认（例如 {"result": null, "id": 1}），记录日志并继续
		if _, ok := generic["result"]; ok {
			logger.Info("Subscription response: %s", string(message))
			continue
		}

		evt, _ := dataMap["e"].(string) // comma-ok safe; if absent evt == ""
		// if evt != "aggTrade" && evt != "avgPrice" && evt != "kline" {
		// 	// logger.Debug("Non-aggTrade message: %v", generic)
		// 	continue
		// }

		// 把原始 data（优先使用 data 字段原始字节）反序列化到结构体，避免浮点/类型问题
		var payload []byte
		if raw, ok := generic["data"]; ok {
			// 重新编码 generic["data"] 为 JSON bytes，然后反序列化
			b, err := json.Marshal(raw)
			if err != nil {
				logger.Error("Failed to marshal generic[data]: %v", err)
				continue
			}
			payload = b
		} else {
			// 如果没有 data 字段，就直接用原始 message（single stream）
			payload = message
		}

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
			c.pool.AddPushDataToQueue(msg)
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
			c.pool.AddPushDataToQueue(msg)

		}

		// || evt == "avgPrice" || evt == "kline"

		// 成功：trade 是结构体，数值字段已正确解析（注意 JSON 中数字若为浮点会被 decode 到 int64）
		// logger.Info("aggTrade parsed: %+v", trade)

		// tradeModelJson, _ := json.MarshalIndent(trade, "", "  ")
		// logger.Debug("Parsed trade model: %s", tradeModelJson)

		// 将 trade 放进池子（如果有 pool）
		// if c.pool != nil {
		// 	// 标记是否为第一次收到该通道的数据
		// 	// isFirst := !hasReceivedData(arg)

		// 	// AddPushDataToQueue 方法在你的 WSPool 中应该存在并接受 interface{} 或特定类型

		// } else {
		// 	// 如果没有 pool，直接记录日志
		// 	logger.Info("aggTrade %s @ %s qty=%s id=%d ts=%d", trade.Price, trade.Symbol, trade.Quantity, trade.AggID, trade.Timestamp)
		// }
		continue
	}

	// 其他消息类型：直接记录（debug）
	// logger.Debug("Received other message: %s", string(message))
	// }

	// 读循环退出后，确保连接被关闭
	_ = conn.Close()

	// // key: channel:tokenContractAddress, value: bool
	// channelDataReceivedMap := make(map[string]bool)

	// hasReceivedData := func(arg *binance_define.WSSubscribeArg) bool {
	// 	key := fmt.Sprintf("%s:%s:%s", arg.Channel, arg.ChainIndex, arg.TokenContractAddress)
	// 	_, exists := channelDataReceivedMap[key]
	// 	if !exists {
	// 		channelDataReceivedMap[key] = true
	// 	}
	// 	return exists
	// }

	// validChannels := []string{"price", "dex-token-candle1s", "trades"}

	// for {
	// 	message, err := c.connector.LockReadPushMessage(conn)
	// 	if err != nil {
	// 		logger.Warn("WebSocket read error: %v", err)
	// 		break
	// 	}

	// 	// log.Printf("Received text message: %s", message)
	// 	if string(message) == "pong" {
	// 		c.onPongResp()
	// 	} else {
	// 		// log.Printf("Received message: msgType = %d, message = %s", msgType, message)
	// 		// 解析消息
	// 		var subscribeResp binance_define.WSSubscribeResp
	// 		if err := json.Unmarshal(message, &subscribeResp); err != nil {
	// 			logger.Error("Failed to unmarshal pushed message = %s, err = %v", message, err)
	// 			// TODO: 不确定是否要处理, 暂时返回吧
	// 			break
	// 		} else if !subscribeResp.IsSuccess() {
	// 			if utils.SliceContains([]string{binance_define.ErrCode_InvalidTimestamp,
	// 				binance_define.ErrCode_TimestampRequestExpired,
	// 				binance_define.ErrCode_PleaseLogin,
	// 				binance_define.ErrCode_RequestsTooFrequent,
	// 				binance_define.ErrCode_NoMultipleLogins,
	// 				binance_define.ErrCode_LoginInternalError,
	// 			}, subscribeResp.Code) {
	// 				// 自动重新登录
	// 				logger.Warn("Re-login due to error: code = %s, msg = %s", subscribeResp.Code, subscribeResp.Msg)
	// 				break
	// 			} else if utils.SliceContains([]string{binance_define.ErrCode_InvalidRequest,
	// 				binance_define.ErrCode_InvalidArgs,
	// 				binance_define.ErrCode_WrongURLOrNotExist,
	// 			}, subscribeResp.Code) {
	// 				// 订阅参数错误, 直接丢弃
	// 				logger.Warn("Invalid subscription arguments, discarding: code = %s, msg = %s", subscribeResp.Code, subscribeResp.Msg)
	// 			} else if utils.SliceContains([]string{binance_define.ErrCode_InvalidApiKey,
	// 				binance_define.ErrCode_InvalidSign,
	// 				binance_define.ErrCode_WrongPassphase,
	// 				binance_define.ErrCode_OnlyWhitelistAllowed,
	// 				binance_define.ErrCode_ApiKeyNotExist,
	// 			}, subscribeResp.Code) {
	// 				// 无法自动恢复的错误, 需要人工介入
	// 				logger.Error("Critical error received, manual intervention required: code = %s, msg = %s", subscribeResp.Code, subscribeResp.Msg)
	// 				break
	// 			} else {
	// 				// 其他错误, 记录日志, 并重新登录
	// 				logger.Warn("Other error occurred, re-login may be required: code = %s, msg = %s", subscribeResp.Code, subscribeResp.Msg)
	// 				break
	// 			}
	// 		} else if subscribeResp.Event == "subscribe" || subscribeResp.Event == "unsubscribe" {
	// 			// 处理订阅响应
	// 			c.onSubscribeResp(&subscribeResp)
	// 		} else if subscribeResp.Event == "notice" {
	// 			// 处理系统通知
	// 			if subscribeResp.Code == binance_define.ErrCode_ConnectionWillCloseForServiceUpgrade {
	// 				logger.Warn("Connection will be closed for service upgrade, reconnecting...")
	// 				break
	// 			} else {
	// 				// 其他通知, 待收集并分类处理
	// 				logger.Error("Received notice: code = %s, msg = %s", subscribeResp.Code, subscribeResp.Msg)
	// 				break
	// 			}
	// 		} else if subscribeResp.Event == "" {
	// 			if utils.SliceContains(validChannels, subscribeResp.Arg.Channel) && len(subscribeResp.Data) > 0 {
	// 				// 处理推送的业务数据
	// 				c.onPushedMsg(&subscribeResp, !hasReceivedData(&subscribeResp.Arg))
	// 			} else {
	// 				logger.Error("Received message: message = %s", message)
	// 				break
	// 			}
	// 		} else {
	// 			subscribeRespJson, _ := json.MarshalIndent(subscribeResp, "", "  ")
	// 			// TODO: 不确定是否要处理, 暂时记录日志吧
	// 			logger.Error("Received message:  message = %s", subscribeRespJson)
	// 			break
	// 		}
	// 	}
	// }
	// c.connector.LockClose(conn)
}

// // TODO: 优化这个函数的逻辑, 需要加一些条件判断, 避免不必要的计算
// func (c *Client) sendPendingMessages(conn *websocket.Conn) error {
// 	// 提取所有待处理的订阅和取消订阅请求参数
// 	// subs, unsubs := c.channelManager.PopAllPendingChannelArgs()

// 	// 发送取消订阅请求
// 	if len(unsubs) > 0 {
// 		if err := c.connector.LockWriteUnsubscribeMessage(conn, unsubs); err != nil {
// 			return err
// 		}
// 		logger.Info("[client-%d] Unsubscribed %d channels", c.index, len(unsubs))
// 	}

// 	// 发送订阅请求
// 	if len(subs) > 0 {
// 		if err := c.connector.LockWriteSubscribeMessage(conn, subs); err != nil {
// 			return err
// 		}
// 		logger.Info("[client-%d] Subscribed %d channels", c.index, len(subs))
// 	}

// 	return nil
// }

func (cm *Client) AddChannelSubscribe(arg *binance_define.RedisChannelArg) {
	// cm.channelManager.AddChannelSubscribe(arg)

}

// func (cm *Client) AddChannelUnsubscribe(arg *binance_define.RedisChannelArg) {
// 	cm.channelManager.AddChannelUnsubscribe(arg)
// }

// func (cm *Client) AddChannelOption(arg *binance_define.RedisChannelArg) {
// 	cm.channelManager.AddChannelOption(arg)
// }

// func (cm *Client) CanAddSubscription(channel string) bool {
// 	return cm.channelManager.GetChannelLen() < 64000+len(channel)
// }

// 给上层调用的关闭接口
func (c *Client) Close() error {
	c.forceStop.Store(true)
	return c.connector.LockClose(c.conn)
}
