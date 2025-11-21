package binance_api

import (
	"Centralized-Data-Collector/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var globalLastConnect time.Time = time.Now().Add(-1 * time.Minute)
var globalConnectMu sync.Mutex // 用于保证全局连接频率不超过 3 次/秒

var ErrWSNotConnected = fmt.Errorf("WS not connected")
var ErrWSNotLogined = fmt.Errorf("WS not logined")

type WConnector struct {
	conn            *websocket.Conn
	connected       bool
	logined         bool
	myMu            sync.Mutex // 保护 conn/connected/logined/lastSendMsgTime 以及 发送消息的频率不超过 8 次/秒
	lastSendMsgTime time.Time
}

func NewWConnector() *WConnector {
	return &WConnector{
		conn:            nil,
		connected:       false,
		logined:         false,
		lastSendMsgTime: time.Now().Add(-1 * time.Minute),
	}
}

func (c *WConnector) LockDialWebSocket(urlStr string, requestHeader http.Header) (*websocket.Conn, error) {
	c.myMu.Lock()
	globalConnectMu.Lock()
	if sleep := 334*time.Millisecond - time.Since(globalLastConnect); sleep > 0 {
		time.Sleep(sleep)
	}
	conn, _, err := websocket.DefaultDialer.Dial(urlStr, requestHeader)
	globalLastConnect = time.Now()
	if err != nil {
		logger.Error("Failed to connect WS: %v", err)
		conn = nil
	} else if conn == nil {
		logger.Error("Failed to connect WS: connection is nil")
		err = fmt.Errorf("connection is nil")
	} else {
		c.connected = true
		c.conn = conn
		c.lastSendMsgTime = time.Now().Add(-1 * time.Minute)
	}
	globalConnectMu.Unlock()
	c.myMu.Unlock()
	return conn, err
}

func (c *WConnector) isConnected(conn *websocket.Conn) bool {
	return c.connected && conn == c.conn
}

func (c *WConnector) LockIsConnected(conn *websocket.Conn) bool {
	c.myMu.Lock()
	isConnected := c.isConnected(conn)
	c.myMu.Unlock()
	return isConnected
}

func (c *WConnector) isLogined(conn *websocket.Conn) bool {
	return c.isConnected(conn) && c.logined
}

func (c *WConnector) LockIsLogined(conn *websocket.Conn) bool {
	c.myMu.Lock()
	isLogined := c.isLogined(conn)
	c.myMu.Unlock()
	return isLogined
}

func (c *WConnector) close(conn *websocket.Conn) error {
	var err error

	if c.isConnected(conn) {
		err = conn.Close()
		c.connected = false
		c.logined = false
	} else {
		err = nil
	}
	return err
}

func (c *WConnector) LockClose(conn *websocket.Conn) error {
	logger.Debug("LockClose start")
	c.myMu.Lock()
	err := c.close(conn)
	c.myMu.Unlock()
	logger.Debug("LockClose end")
	return err
}

func (c *WConnector) writeMessage(conn *websocket.Conn, data []byte) error {
	var err error
	if c.isConnected(conn) {
		err = conn.WriteMessage(websocket.TextMessage, data)
		c.lastSendMsgTime = time.Now()
	} else {
		err = ErrWSNotConnected
	}
	return err
}

func (c *WConnector) writeJSON(conn *websocket.Conn, v interface{}) error {
	jsonData, _ := json.Marshal(v)
	return c.writeMessage(conn, jsonData)
}

func (c *WConnector) LockWritePingMessage(conn *websocket.Conn) error {
	c.myMu.Lock()
	if sleep := 125*time.Millisecond - time.Since(c.lastSendMsgTime); sleep > 0 {
		time.Sleep(sleep)
	}
	err := c.writeMessage(conn, []byte("ping"))
	c.myMu.Unlock()
	return err
}

func (c *WConnector) readMessage(conn *websocket.Conn) (p []byte, err error) {
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		logger.Warn("WebSocket read error: %v", err)
		return nil, err
	} else if messageType != websocket.TextMessage {
		logger.Error("WebSocket read invalid message type: %d", messageType)
		return nil, fmt.Errorf("invalid message type: %d", messageType)
	}
	return message, err
}

func (c *WConnector) LockReadPushMessage(conn *websocket.Conn) (p []byte, err error) {
	if !c.LockIsLogined(conn) {
		return nil, ErrWSNotLogined
	}
	return c.readMessage(conn)
}

func (c *WConnector) LockLogin(conn *websocket.Conn) (err error) {
	c.myMu.Lock()
	defer c.myMu.Unlock()
	if !c.isConnected(conn) {
		return ErrWSNotConnected
	}
	if sleep := 125*time.Millisecond - time.Since(c.lastSendMsgTime); sleep > 0 {
		time.Sleep(sleep)
	}
	c.logined = true

	return nil
}
