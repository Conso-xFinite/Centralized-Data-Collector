package binance_api

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/utils"
	"sync"
)

type WSChannelManager struct {
	mutex                 sync.Mutex
	pendingChannelArgList *utils.List[*binance_define.RedisChannelArg]
	subscribedChannels    *utils.Map[string, *binance_define.RedisChannelArg] // 记录已订阅频道列表, key: "<channel>:<ChainIndex>:<TokenContractAddress>", value: bool
	channelLen            int
}

func pendingChannelArgListEqualFunc(a, b *binance_define.RedisChannelArg) bool {
	return a.Channel == b.Channel &&
		a.ChainIndex == b.ChainIndex &&
		a.TokenContractAddress == b.TokenContractAddress
}

func NewWSChannelManager() *WSChannelManager {
	utils.NewList[*binance_define.RedisChannelArg]()
	return &WSChannelManager{
		pendingChannelArgList: utils.NewListWithEqualFunc[*binance_define.RedisChannelArg](pendingChannelArgListEqualFunc),
		subscribedChannels:    utils.NewMap[string, *binance_define.RedisChannelArg](),
		channelLen:            0,
	}
}

// // 是否可能接受更多的订阅
// func (cm *WSChannelManager) GetChannelLen() int {
// 	cm.mutex.Lock()
// 	len := cm.channelLen
// 	cm.mutex.Unlock()
// 	return len
// }

// func (cm *WSChannelManager) getKey(channel, chainIndex, tokenContractAddress string) string {
// 	return channel + ":" + chainIndex + ":" + tokenContractAddress
// }

// func (cm *WSChannelManager) _addChannelSubscribe(arg *okx_define.RedisOkxChannelArg) {
// 	// 检查是否已经存在未处理的订阅或取消请求
// 	index := cm.pendingChannelArgList.IndexOf(arg)
// 	if index >= 0 {
// 		pendingArg, _ := cm.pendingChannelArgList.Get(index)
// 		if pendingArg.IsSubscribe {
// 			// 已经存在未处理的订阅请求 -> do nothing
// 		} else {
// 			// 存在未处理的取消请求 -> remove it
// 			cm.channelLen += len(arg.Channel)
// 			cm.pendingChannelArgList.RemoveAt(index)
// 		}
// 	} else {
// 		// 判断是否已经订阅
// 		key := cm.getKey(arg.Channel, arg.ChainIndex, arg.TokenContractAddress)
// 		if _, hasSubscribed := cm.subscribedChannels.Get(key); hasSubscribed {
// 			// already subscribed -> do nothing
// 		} else {
// 			// not subscribed -> can subscribe
// 			cm.channelLen += len(arg.Channel)
// 			arg.IsSubscribe = true
// 			cm.pendingChannelArgList.Append(arg)
// 		}
// 	}
// }

// func (cm *WSChannelManager) _addChannelUnsubscribe(arg *okx_define.RedisOkxChannelArg) {
// 	// 检查是否已经存在未处理的订阅或取消请求
// 	index := cm.pendingChannelArgList.IndexOf(arg)
// 	if index >= 0 {
// 		pendingArg, _ := cm.pendingChannelArgList.Get(index)
// 		if !pendingArg.IsSubscribe {
// 			// 已经存在未处理的取消请求 -> do nothing
// 		} else {
// 			// 存在未处理的订阅请求 -> remove it
// 			cm.channelLen -= len(arg.Channel)
// 			cm.pendingChannelArgList.RemoveAt(index)
// 		}
// 	} else {
// 		// 判断是否已经订阅
// 		key := cm.getKey(arg.Channel, arg.ChainIndex, arg.TokenContractAddress)
// 		if _, hasSubscribed := cm.subscribedChannels.Get(key); hasSubscribed {
// 			// subscribed -> can unsubscribe
// 			cm.channelLen -= len(arg.Channel)
// 			arg.IsSubscribe = false
// 			cm.pendingChannelArgList.Append(arg)
// 		} else {
// 			// not subscribed -> do nothing
// 		}
// 	}
// }

// func (cm *WSChannelManager) _addChannelOption(arg *okx_define.RedisOkxChannelArg) {
// 	if arg.IsSubscribe {
// 		cm._addChannelSubscribe(arg)
// 	} else {
// 		cm._addChannelUnsubscribe(arg)
// 	}
// }

// func (cm *WSChannelManager) AddChannelSubscribe(arg *okx_define.RedisOkxChannelArg) {
// 	cm.mutex.Lock()
// 	cm._addChannelSubscribe(arg)
// 	cm.mutex.Unlock()
// }

// func (cm *WSChannelManager) AddChannelUnsubscribe(arg *okx_define.RedisOkxChannelArg) {
// 	cm.mutex.Lock()
// 	cm._addChannelUnsubscribe(arg)
// 	cm.mutex.Unlock()
// }

// func (cm *WSChannelManager) AddChannelOption(arg *okx_define.RedisOkxChannelArg) {
// 	cm.mutex.Lock()
// 	cm._addChannelOption(arg)
// 	cm.mutex.Unlock()
// }

// // 重连后, 将已订阅的频道全部推到pending队列中
// func (cm *WSChannelManager) RequeueAllSubscribedChannels() {
// 	cm.mutex.Lock()

// 	// 备份旧的pending列表
// 	oldPendingChannelArgs := cm.pendingChannelArgList.All()

// 	// 清空pending列表
// 	cm.pendingChannelArgList.Clear()

// 	// 将已订阅的频道全部推到pending队列中
// 	subs := cm.subscribedChannels.Values()
// 	for _, sub := range subs {
// 		cm.pendingChannelArgList.Append(sub)
// 	}

// 	// 清空已订阅列表
// 	cm.subscribedChannels.Clear()

// 	// 重新添加旧的pending列表
// 	for i := 0; i < len(oldPendingChannelArgs); i++ {
// 		cm._addChannelOption(oldPendingChannelArgs[i])
// 	}

// 	cm.mutex.Unlock()
// }

// // 提取并发送所有待处理的订阅和取消订阅请求参数
// func (cm *WSChannelManager) PopAllPendingChannelArgs() (subs, unsubs []*okx_define.WSSubscribeArg) {
// 	cm.mutex.Lock()
// 	pendingArgs := cm.pendingChannelArgList.All()
// 	cm.pendingChannelArgList.Clear()
// 	cm.mutex.Unlock()

// 	subs = make([]*okx_define.WSSubscribeArg, 0)
// 	unsubs = make([]*okx_define.WSSubscribeArg, 0)

// 	for _, arg := range pendingArgs {
// 		operationArg := &okx_define.WSSubscribeArg{
// 			Channel:              arg.Channel,
// 			ChainIndex:           arg.ChainIndex,
// 			TokenContractAddress: arg.TokenContractAddress,
// 		}
// 		key := cm.getKey(arg.Channel, arg.ChainIndex, arg.TokenContractAddress)
// 		if arg.IsSubscribe {
// 			subs = append(subs, operationArg)
// 			cm.subscribedChannels.Set(key, arg)
// 		} else {
// 			unsubs = append(unsubs, operationArg)
// 			cm.subscribedChannels.Delete(key)
// 		}
// 	}
// 	return subs, unsubs
// }
