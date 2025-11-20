package collector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ThrottleSignalProcessor 支持短时间内重复信号丢弃
type BlockingSignalProcessor struct {
	queue   chan struct{} // 信号队列（无 payload）
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	handler func(ctx context.Context) error
}

// NewThrottleSignalProcessor 创建 processor，buffer 表示队列容量
func NewBlockingSignalProcessor(buffer int, handler func(ctx context.Context) error) *BlockingSignalProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	p := &BlockingSignalProcessor{
		queue:   make(chan struct{}, buffer),
		ctx:     ctx,
		cancel:  cancel,
		handler: handler,
	}
	p.wg.Add(1)
	go p.run()
	return p
}

// run 背景协程：顺序消费队列
func (p *BlockingSignalProcessor) run() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			// 优雅退出：处理完队列中剩余的信号
			for range p.queue {
				if err := p.handler(p.ctx); err != nil {
					fmt.Println("handler error:", err)
				}
			}
			return
		case _, ok := <-p.queue:
			if !ok {
				return
			}
			// 同步处理，保证顺序
			if err := p.handler(p.ctx); err != nil {
				fmt.Println("handler error:", err)
			}
		}
	}
}

// Signal 入队，如果队列非空且在 1 秒内有信号，直接丢弃
func (p *BlockingSignalProcessor) Signal() {
	select {
	case <-p.ctx.Done():
		return
	default:
	}

	// 尝试入队，非阻塞
	select {
	case p.queue <- struct{}{}:
		// 入队成功
	case <-time.After(time.Second):
		// 超过 1 秒仍然不能入队，直接丢弃
		fmt.Println("signal dropped due to throttle")
	}
}

// Stop 优雅停止
func (p *BlockingSignalProcessor) Stop() {
	p.cancel()
	close(p.queue)
	p.wg.Wait()
}
