package binance_api

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"strings"
	"sync"
)

type WSChannelManager struct {
	mutex                            sync.Mutex
	subscribePendingChannelArgList   *utils.List[*binance_define.RedisChannelArg] //等待订阅列表
	unsubscribePendingChannelArgList *utils.List[*binance_define.RedisChannelArg] //等待取消订阅列表
	subscribedChannels               *utils.Map[string, bool]                     // 记录已订阅频道列表, key: "<channel>@<TokenPair", value: bool
}

func pendingChannelArgListEqualFunc(a, b *binance_define.RedisChannelArg) bool {
	return a.TokenPair == b.TokenPair &&
		a.Channel == b.Channel
}

func NewWSChannelManager() *WSChannelManager {
	return &WSChannelManager{
		subscribePendingChannelArgList:   utils.NewListWithEqualFunc[*binance_define.RedisChannelArg](pendingChannelArgListEqualFunc),
		unsubscribePendingChannelArgList: utils.NewListWithEqualFunc[*binance_define.RedisChannelArg](pendingChannelArgListEqualFunc),
		subscribedChannels:               utils.NewMap[string, bool](),
	}
}

func (cm *WSChannelManager) _addChannelSubscribe(arg *binance_define.RedisChannelArg) {
	// 检查是否已经存在未处理的订阅
	index := cm.subscribePendingChannelArgList.IndexOf(arg)
	//外部已经判断是是否已完成订阅，如果pending队列不存在就需要加入
	if index < 0 {
		cm.subscribePendingChannelArgList.Append(arg)
	} else {
		logger.Debug("已存在pending 队列中")
	}
}

func (cm *WSChannelManager) _reAddSubscibeList() {
	logger.Debug("断线重连后的原已订阅数组长度 %d", len(cm.subscribedChannels.Keys()))
	for _, key := range cm.subscribedChannels.Keys() {
		parts := strings.Split(key, "@")
		arg := binance_define.RedisChannelArg{
			TypeSubscribe: "subscribe",
			Channel:       parts[1],
			TokenPair:     parts[0],
		}
		cm.subscribePendingChannelArgList.Append(&arg)
	}
	cm.subscribedChannels.Clear()
}

func (cm *WSChannelManager) _addChannelUnsubscribe(arg *binance_define.RedisChannelArg) {
	// 检查是否已经存在未处理的订阅
	cm.unsubscribePendingChannelArgList.Append(arg)

}

func (cm *WSChannelManager) _isInSubscribe(stream string) bool {
	// 检查是否已经存在未处理的订阅
	_, ok := cm.subscribedChannels.Get(stream)
	return ok
}

func (cm *WSChannelManager) _BatchaddPendingChannelSubscribe(args *utils.List[*binance_define.RedisChannelArg]) {
	cm.subscribePendingChannelArgList.AppendBatch(args.All())
}

func (cm *WSChannelManager) AddChannelSubscribe(arg *binance_define.RedisChannelArg) {
	cm.mutex.Lock()
	cm._addChannelSubscribe(arg)
	cm.mutex.Unlock()
}

func (cm *WSChannelManager) reAddSubscibeList() {
	cm.mutex.Lock()
	cm._reAddSubscibeList()
	cm.mutex.Unlock()
}

func (cm *WSChannelManager) BatchAddChannelSubscribe(args *utils.List[*binance_define.RedisChannelArg]) {
	cm.mutex.Lock()
	cm._BatchaddPendingChannelSubscribe(args)
	cm.mutex.Unlock()
}

func (cm *WSChannelManager) AddChannelUnsubscribe(arg *binance_define.RedisChannelArg) {
	cm.mutex.Lock()
	cm._addChannelUnsubscribe(arg)
	cm.mutex.Unlock()
}

func (cm *WSChannelManager) IsInSubscirbe(stream string) bool {
	return cm._isInSubscribe(stream)
}
