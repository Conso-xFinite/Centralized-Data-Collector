package binance_api

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"encoding/json"
	"log"
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
		logger.Info("Client-%d 需要订阅的数据 %s", c.index, len(clientSubscribelist))
		logger.Info("Client-%d 已订阅的数据 %s", c.index, len(c.channelManager.subscribedChannels.Keys()))
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
		logger.Info("Client-%d 需要订阅的数据 %s", c.index, len(clientUnsubscribelist))
		logger.Info("Client-%d 已订阅的数据 %s", c.index, len(c.channelManager.subscribedChannels.Keys()))
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
