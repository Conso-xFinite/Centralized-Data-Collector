package collector

import (
	"context"
	"fmt"
	"sync"
)

// BlockingSignalProcessor 支持阻塞式入队的信号处理器
type BlockingSignalProcessor struct {
	queue   chan struct{} // 信号队列（无 payload）
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	handler func(ctx context.Context) error
}

// NewBlockingSignalProcessor 创建 processor，buffer 表示队列容量
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
			for v := range p.queue {
				_ = p.handler(p.ctx)
				_ = v
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

// Signal 阻塞式入队：如果队列满，会等待空间
func (p *BlockingSignalProcessor) Signal() {
	select {
	case <-p.ctx.Done():
		return
	case p.queue <- struct{}{}: // 阻塞直到有空间
	}
}

// Stop 优雅停止
func (p *BlockingSignalProcessor) Stop() {
	p.cancel()
	close(p.queue)
	p.wg.Wait()
}

///// 示例用法 /////

// func exampleHandler(ctx context.Context) error {
// 	fmt.Println("start handling at", time.Now().Format("15:04:05.000"))
// 	time.Sleep(500 * time.Millisecond)
// 	fmt.Println("finish handling at", time.Now().Format("15:04:05.000"))
// 	return nil
// }
