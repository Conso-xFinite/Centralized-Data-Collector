package binance_api

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
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
		logger.Debug("Received ping: %s", appData)
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
					logger.Info("Client-%d connected and logged in", c.index)
					conn = c.conn
				}
			} else {
				if pastTime := utils.GetCurrentTimestampSec() - c.lastPongTimestamp.Load(); pastTime >= 60 {
					logger.Error("No ping received in the last 60 seconds, reconnecting...")
					c.connector.LockClose(c.conn)
				}
				// 动态订阅逻辑在这里实现
				if err := c.sendPendingMessages(conn); err != nil {
					log.Printf("sendPendingMessages error: %v", err)
					c.connector.LockClose(c.conn)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
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
	//是否需要判断 连接状态后续确认
	for {
		message, err := c.connector.LockReadPushMessage(conn)
		if err != nil {
			logger.Warn("WebSocket read error: %v", err)
			break
		}
		// logger.Debug("Received message: %s", string(message))
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
			dataMap = generic
		}

		if _, ok := generic["result"]; ok {
			logger.Info("Subscription response: %s", string(message))
			continue
		}

		evt, _ := dataMap["e"].(string)

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
			c.pool.AddPushDataToQueue(msg)
		} else {
			logger.Debug("Received other event type: %s, message: %s", evt, string(message))
		}
		continue
	}

}

// TODO: 优化这个函数的逻辑, 需要加一些条件判断, 避免不必要的计算
func (c *Client) sendPendingMessages(conn *websocket.Conn) error {
	clientSubscribelist := c.channelManager.subscribePendingChannelArgList.All()
	if len(clientSubscribelist) > 0 {
		logger.Debug("client sendPendingMessages d%", clientSubscribelist)
		values := []string{} // 创建空切片
		for _, subscribe := range clientSubscribelist {
			values = append(values, subscribe.TokenPair+"@"+subscribe.Channel)
		}
		logger.Debug("subscribe values s%", values)
		subMsg := binance_define.SubscribeMsg{
			Method: "SUBSCRIBE",
			Params: values,
			ID:     1,
		}
		data, _ := json.Marshal(subMsg)

		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Subscribe error: %v", err)
		}
		for _, v := range values {
			c.channelManager.subscribedChannels.Set(v, true)
		}
		c.channelManager.subscribePendingChannelArgList.Clear()
	}

	// 动态订阅 BTCUSDT 的 aggTrade
	// subMsg := binance_define.SubscribeMsg{
	// 	Method: "SUBSCRIBE",
	// 	Params: []string{
	// 		"btcusdt@aggTrade", "btcusdt@kline_1m", "btcusdt@miniTicker",
	// 		"ethusdt@aggTrade", "ethusdt@kline_1m", "ethusdt@miniTicker",
	// 		"bnbusdt@aggTrade", "bnbusdt@kline_1m", "bnbusdt@miniTicker",
	// 		"usdtusdt@aggTrade", "usdtusdt@kline_1m", "usdtusdt@miniTicker",
	// 		"xrpusdt@aggTrade", "xrpusdt@kline_1m", "xrpusdt@miniTicker",
	// 		"adausdt@aggTrade", "adausdt@kline_1m", "adausdt@miniTicker",
	// 		"dogeusdt@aggTrade", "dogeusdt@kline_1m", "dogeusdt@miniTicker",
	// 		"maticusdt@aggTrade", "maticusdt@kline_1m", "maticusdt@miniTicker",
	// 		"solusdt@aggTrade", "solusdt@kline_1m", "solusdt@miniTicker",
	// 		"dotusdt@aggTrade", "dotusdt@kline_1m", "dotusdt@miniTicker",
	// 		"avaxusdt@aggTrade", "avaxusdt@kline_1m", "avaxusdt@miniTicker",
	// 		"shibusdt@aggTrade", "shibusdt@kline_1m", "shibusdt@miniTicker",
	// 		"tronusdt@aggTrade", "tronusdt@kline_1m", "tronusdt@miniTicker",
	// 		"wbtcusdt@aggTrade", "wbtcusdt@kline_1m", "wbtcusdt@miniTicker",
	// 		"linkusdt@aggTrade", "linkusdt@kline_1m", "linkusdt@miniTicker",
	// 		"nearusdt@aggTrade", "nearusdt@kline_1m", "nearusdt@miniTicker",
	// 		"ltcusdt@aggTrade", "ltcusdt@kline_1m", "ltcusdt@miniTicker",
	// 		"atomusdt@aggTrade", "atomusdt@kline_1m", "atomusdt@miniTicker",
	// 		"xlmusdt@aggTrade", "xlmusdt@kline_1m", "xlmusdt@miniTicker",
	// 		"fttusdt@aggTrade", "fttusdt@kline_1m", "fttusdt@miniTicker",
	// 		"filecoinusdt@aggTrade", "filecoinusdt@kline_1m", "filecoinusdt@miniTicker",
	// 		"apeusdt@aggTrade", "apeusdt@kline_1m", "apeusdt@miniTicker",
	// 		"cakeusdt@aggTrade", "cakeusdt@kline_1m", "cakeusdt@miniTicker",
	// 		"etcusdt@aggTrade", "etcusdt@kline_1m", "etcusdt@miniTicker",
	// 		"egldusdt@aggTrade", "egldusdt@kline_1m", "egldusdt@miniTicker",
	// 		"thetausdt@aggTrade", "thetausdt@kline_1m", "thetausdt@miniTicker",
	// 		"vechainusdt@aggTrade", "vechainusdt@kline_1m", "vechainusdt@miniTicker",
	// 		"xtzusdt@aggTrade", "xtzusdt@kline_1m", "xtzusdt@miniTicker",
	// 		"uniusdt@aggTrade", "uniusdt@kline_1m", "uniusdt@miniTicker",
	// 		"algorandusdt@aggTrade", "algorandusdt@kline_1m", "algorandusdt@miniTicker",
	// 		"ftmusdt@aggTrade", "ftmusdt@kline_1m", "ftmusdt@miniTicker",
	// 		"reefusdt@aggTrade", "reefusdt@kline_1m", "reefusdt@miniTicker",
	// 		"galausdt@aggTrade", "galausdt@kline_1m", "galausdt@miniTicker",
	// 		"eosusdt@aggTrade", "eosusdt@kline_1m", "eosusdt@miniTicker",
	// 		"axsusdt@aggTrade", "axsusdt@kline_1m", "axsusdt@miniTicker",
	// 		"hntusdt@aggTrade", "hntusdt@kline_1m", "hntusdt@miniTicker",
	// 		"neousdt@aggTrade", "neousdt@kline_1m", "neousdt@miniTicker",
	// 		"chzusdt@aggTrade", "chzusdt@kline_1m", "chzusdt@miniTicker",
	// 		"arusdt@aggTrade", "arusdt@kline_1m", "arusdt@miniTicker",
	// 		"runeusdt@aggTrade", "runeusdt@kline_1m", "runeusdt@miniTicker",
	// 		"zecusdt@aggTrade", "zecusdt@kline_1m", "zecusdt@miniTicker",
	// 		"bttusdt@aggTrade", "bttusdt@kline_1m", "bttusdt@miniTicker",
	// 		"gmxusdt@aggTrade", "gmxusdt@kline_1m", "gmxusdt@miniTicker",
	// 		"crvusdt@aggTrade", "crvusdt@kline_1m", "crvusdt@miniTicker",
	// 		"klayusdt@aggTrade", "klayusdt@kline_1m", "klayusdt@miniTicker",
	// 		"paxgusdt@aggTrade", "paxgusdt@kline_1m", "paxgusdt@miniTicker",
	// 		"icpusdt@aggTrade", "icpusdt@kline_1m", "icpusdt@miniTicker",
	// 		"manausdt@aggTrade", "manausdt@kline_1m", "manausdt@miniTicker",
	// 		"enjusdt@aggTrade", "enjusdt@kline_1m", "enjusdt@miniTicker",
	// 		"sandusdt@aggTrade", "sandusdt@kline_1m", "sandusdt@miniTicker",
	// 		"zilusdt@aggTrade", "zilusdt@kline_1m", "zilusdt@miniTicker",
	// 		"dashusdt@aggTrade", "dashusdt@kline_1m", "dashusdt@miniTicker",
	// 		"omgusdt@aggTrade", "omgusdt@kline_1m", "omgusdt@miniTicker",
	// 		"wavesusdt@aggTrade", "wavesusdt@kline_1m", "wavesusdt@miniTicker",
	// 		"ankrusdt@aggTrade", "ankrusdt@kline_1m", "ankrusdt@miniTicker",
	// 		"qntusdt@aggTrade", "qntusdt@kline_1m", "qntusdt@miniTicker",
	// 		"minausdt@aggTrade", "minausdt@kline_1m", "minausdt@miniTicker",
	// 		"aaveusdt@aggTrade", "aaveusdt@kline_1m", "aaveusdt@miniTicker",
	// 		"zrxusdt@aggTrade", "zrxusdt@kline_1m", "zrxusdt@miniTicker",
	// 		"bakcusdt@aggTrade", "bakcusdt@kline_1m", "bakcusdt@miniTicker",
	// 		"rendusdt@aggTrade", "rendusdt@kline_1m", "rendusdt@miniTicker",
	// 		"leousdt@aggTrade", "leousdt@kline_1m", "leousdt@miniTicker",
	// 		"tomousdt@aggTrade", "tomousdt@kline_1m", "tomousdt@miniTicker",
	// 		"stxusdt@aggTrade", "stxusdt@kline_1m", "stxusdt@miniTicker",
	// 		"iousdt@aggTrade", "iousdt@kline_1m", "iousdt@miniTicker",
	// 		"tgto@aggTrade", "tgto@kline_1m", "tgto@miniTicker",

	// 		// "arweaveusdt@aggTrade", "arweaveusdt@kline_1m", "arweaveusdt@miniTicker",
	// 		// "lrcusdt@aggTrade", "lrcusdt@kline_1m", "lrcusdt@miniTicker",
	// 	},
	// ID: 1,

	// hasSubscribe = true
	// }

	// 提取所有待处理的订阅和取消订阅请求参数
	// subs, unsubs := c.channelManager.PopAllPendingChannelArgs()

	// // 发送取消订阅请求
	// if len(unsubs) > 0 {
	// 	if err := c.connector.LockWriteUnsubscribeMessage(conn, unsubs); err != nil {
	// 		return err
	// 	}
	// 	logger.Info("[client-%d] Unsubscribed %d channels", c.index, len(unsubs))
	// }

	// // 发送订阅请求
	// if len(subs) > 0 {
	// 	if err := c.connector.LockWriteSubscribeMessage(conn, subs); err != nil {
	// 		return err
	// 	}
	// 	logger.Info("[client-%d] Subscribed %d channels", c.index, len(subs))
	// }

	return nil
}

func (cm *Client) AddChannelSubscribe(arg *binance_define.RedisChannelArg) {
	cm.channelManager.AddChannelSubscribe(arg)

}

func (cm *Client) BatchAddChannelSubscribe(args *utils.List[*binance_define.RedisChannelArg]) {
	cm.channelManager.BatchAddChannelSubscribe(args)

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
