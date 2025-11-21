package binance_api

/* OKXWSPool 管理多个 OKX WebSocket 客户端实例
 * 每个客户端实例对应一个连接，可以根据需要扩展以支持更多功能
 * 连接限制， 每秒3个请求(基于 API KEY)
 * 连接数限制， 每个 API KEY 最多 30 个连接
 * 每个连接的 订阅/取消订阅/登录 请求 也有限制， 每125毫秒1个请求
 * 如果出现网络问题，okx系统会自动关闭连接。
 * 如果订阅未建立 或者超过30秒未推送数据， 连接将自动中断。
 * 保持连接稳定：
 * 1. 实现心跳机制，定期发送心跳包，确保连接活跃， 心跳间隔不超过30秒。发送字符串 "ping"，期望收到 "pong" 响应。
 * 2. 实现重连机制，当连接断开时，自动尝试重连。
 */

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"sync"
	"sync/atomic"
)

// OKXWSPool 管理多个 OKX WebSocket 客户端实例
type WSPool struct {
	clients []*Client
	// 设定一个数组, 用于记录每个连接是否已完成初始化连接
	clientsInitialized   []atomic.Bool // 用于保证 client 按顺序初始化
	connCount            atomic.Int32
	mu                   sync.Mutex
	singlePushedMsgQueue *utils.SafeQueue[*binance_define.WSSinglePushMsg]
}

// NewOKXWSPool 创建连接池
func NewWSPool(maxConn int) *WSPool {
	pool := &WSPool{
		clients:              make([]*Client, maxConn),
		singlePushedMsgQueue: utils.NewSafeQueue[*binance_define.WSSinglePushMsg](),
		clientsInitialized:   make([]atomic.Bool, maxConn),
	}

	for i := 0; i < maxConn; i++ {
		pool.clientsInitialized[i].Store(false)
		client := NewClient(pool, i)
		client.Start()
		pool.clients[i] = client
		pool.connCount.Add(1)
	}
	return pool
}

func (p *WSPool) GetClientsLength() int {
	len := len(p.clients)
	return len
}

// 判断某个连接已否已经完成初始化连接
func (p *WSPool) IsClientInitialized(index int) bool {
	return p.clientsInitialized[index].Load()
}

func (p *WSPool) FinishClientInitialized(index int) {
	p.clientsInitialized[index].Store(true)
}

// 获取一个可以接收更多订阅的连接索引
func (p *WSPool) GetCanAddSubscriptionClientIndex(channel string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := 0; i < int(p.connCount.Load()); i++ {
		// if p.clients[i].CanAddSubscription(channel) {
		// 	return i
		// }
	}
	return -1
}

func (p *WSPool) GetKey(tokenPair, channel string) string {
	return tokenPair + "@" + channel
}

func (p *WSPool) GetClientByIndex(index int32) *Client {
	return p.clients[index]
}

func (p *WSPool) AddChannelSubscribe(arg *binance_define.RedisChannelArg) bool {

	key := p.GetKey(arg.TokenPair, arg.Channel)
	index := utils.HMACSHA256FromStringToUint(key, &p.connCount)

	_, ok := p.clients[index].channelManager.subscribedChannels.Get(key)
	if ok {
		return false
	}

	p.clients[index].AddChannelSubscribe(arg)
	logger.Info("[client-%d] subscription, %s: %s",
		index, arg.Channel, arg.TokenPair)
	return true
}

func (p *WSPool) BatchAddChannelSubscribe(index int32, args *utils.List[*binance_define.RedisChannelArg]) bool {
	p.clients[index].BatchAddChannelSubscribe(args)
	return true
}

func (p *WSPool) GetChanneJoinIndex(arg *binance_define.RedisChannelArg) int32 {
	key := p.GetKey(arg.TokenPair, arg.Channel)
	index := utils.HMACSHA256FromStringToUint(key, &p.connCount)
	_, ok := p.clients[index].channelManager.subscribedChannels.Get(key)
	if ok {
		return -1
	}

	return int32(index)
}

func (p *WSPool) AddChannelUnsubscribe(arg *binance_define.RedisChannelArg) bool {
	key := p.GetKey(arg.TokenPair, arg.Channel)

	index := utils.HMACSHA256FromStringToUint(key, &p.connCount)
	// _, ok := p.clients[index].channelManager.subscribedChannels.Get(key)
	// if !ok {
	// 	return false
	// }
	p.clients[index].AddChannelUnsubscribe(arg)
	logger.Info("[client-%d] unsubscription, %s:",
		index, key)
	return true
}

func (p *WSPool) AddPushDataToQueue(subscribeResp *binance_define.WSSinglePushMsg) {
	// dataJson, _ := json.MarshalIndent(subscribeResp.Data, "", "  ")
	// logger.Debug("received [%s]: %s", subscribeResp.Arg.Channel, dataJson)
	p.singlePushedMsgQueue.Push(subscribeResp)
}

func (p *WSPool) FetchData() []*binance_define.WSSinglePushMsg {
	return p.singlePushedMsgQueue.PopBatch(1000)
}

// Close 关闭所有连接
func (p *WSPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.clients {
		c.Close()
	}
	p.clients = nil
}
