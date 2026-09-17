package engine

import (
	"context"
	"sync"
	"sync/atomic"
)

// Pool 是一个动态可调整的 Goroutine 池
// 用于控制最大并发 VUs
type Pool struct {
	minSize int
	maxSize int

	mu       sync.RWMutex
	workers  []*worker
	taskCh   chan func()
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc

	activeCount atomic.Int64
	idleCount   atomic.Int64

	closed atomic.Bool
}

type worker struct {
	id      int
	taskCh  chan func()
	stopped chan struct{}
}

// NewPool 创建新的 Goroutine Pool
func NewPool(minSize, maxSize int) *Pool {
	if minSize < 0 {
		minSize = 0
	}
	if maxSize < minSize {
		maxSize = minSize
	}
	if maxSize == 0 {
		maxSize = 100 // 默认最大 100 并发
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		minSize: minSize,
		maxSize: maxSize,
		taskCh:  make(chan func(), maxSize*2),
		ctx:     ctx,
		cancel:  cancel,
	}

	// 启动最小数量的 worker
	p.wg.Add(minSize)
	for i := 0; i < minSize; i++ {
		p.spawnWorker(i)
	}

	return p
}

// spawnWorker 启动一个 worker goroutine
func (p *Pool) spawnWorker(id int) {
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.ctx.Done():
				return
			case task, ok := <-p.taskCh:
				if !ok {
					return
				}
				p.activeCount.Add(1)
				p.idleCount.Add(-1)
				// 任务执行过程中不允许阻塞太久
				func() {
					defer func() {
						// 防止 panic 影响整个 pool
						_ = recover()
						p.activeCount.Add(-1)
						p.idleCount.Add(1)
					}()
					task()
				}()
			}
		}
	}()
}

// Submit 提交一个任务到 Pool
// 如果 Pool 已关闭,返回 false
func (p *Pool) Submit(task func()) bool {
	if p.closed.Load() {
		return false
	}
	select {
	case <-p.ctx.Done():
		return false
	case p.taskCh <- task:
		return true
	default:
		// taskCh 满了,尝试扩展 worker
		if p.tryExpand() {
			select {
			case p.taskCh <- task:
				return true
			default:
				return false
			}
		}
		return false
	}
}

// tryExpand 尝试增加一个 worker
func (p *Pool) tryExpand() bool {
	p.mu.Lock()
	current := len(p.workers) + p.minSize
	p.mu.Unlock()
	if current >= p.maxSize {
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	current = len(p.workers) + p.minSize
	if current >= p.maxSize {
		return false
	}

	p.wg.Add(1)
	p.spawnWorker(len(p.workers) + p.minSize)
	p.idleCount.Add(1)
	return true
}

// Size 返回当前 worker 数
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.workers) + p.minSize
}

// Active 返回当前正在执行任务的 worker 数
func (p *Pool) Active() int64 {
	return p.activeCount.Load()
}

// Idle 返回空闲 worker 数
func (p *Pool) Idle() int64 {
	return p.idleCount.Load()
}

// Close 关闭 Pool
// 等待所有正在执行的任务完成
func (p *Pool) Close() {
	if p.closed.Swap(true) {
		return // 已经关闭
	}
	p.cancel()
	close(p.taskCh)
	p.wg.Wait()
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	for p.taskCh != nil {
		select {
		case <-p.ctx.Done():
			return
		default:
			if p.Active() == 0 && len(p.taskCh) == 0 {
				return
			}
		}
	}
}