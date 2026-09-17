package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler 负责调度 VU(虚拟用户)的启动/停止
// 支持线性、阶梯、波浪三种 Ramp-up 模式
type Scheduler struct {
	totalVUs int
	config   LoadConfig
	callback VUCallback

	// 控制信号
	startCh  chan struct{} // 启动新 VU
	stopCh   chan struct{} // 停止 VU
	pauseCh  chan struct{} // 暂停信号
	resumeCh chan struct{} // 恢复信号
	doneCh   chan struct{} // 所有 VU 结束

	// 状态
	activeVUs atomic.Int64
	startedAt time.Time

	mu     sync.Mutex
	paused bool

	// VU 控制:每个 VU 一个 chan
	vuStopChs []chan struct{}
}

// VUCallback 调度器回调 VU 执行
type VUCallback interface {
	OnVUStart(vuID int) error
	OnVUStop(vuID int)
}

// NewScheduler 创建 Scheduler
func NewScheduler(totalVUs int, cfg LoadConfig, callback VUCallback) *Scheduler {
	return &Scheduler{
		totalVUs: totalVUs,
		config:   cfg,
		callback: callback,
		startCh:  make(chan struct{}, totalVUs),
		stopCh:   make(chan struct{}),
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Run 启动并运行调度器,直到 context 取消或时间结束
// 返回当所有 VU 结束时
func (s *Scheduler) Run(ctx context.Context) error {
	s.startedAt = time.Now()

	// 计算 Ramp-up 计划
	plan := s.buildRampUpPlan(planInput{
		totalVUs: s.totalVUs,
		rampUp:   s.config.RampUp,
		duration: s.config.Duration,
	})

	// 启动 VU 控制通道
	s.mu.Lock()
	s.vuStopChs = make([]chan struct{}, 0, s.totalVUs)
	s.mu.Unlock()

	// Ramp-up 调度协程
	rampUpDone := make(chan struct{})
	go func() {
		defer close(rampUpDone)
		s.executeRampUp(ctx, plan)
	}()

	// 等待 context 取消或 duration 到期
	timer := time.NewTimer(s.config.Duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		// context 取消,停止所有 VU
		s.stopAllVUs()
	case <-timer.C:
		// duration 到期,停止所有 VU
		s.stopAllVUs()
	case <-s.stopCh:
		// 外部主动停止
		s.stopAllVUs()
	}

	// 等待 ramp-up 完成
	<-rampUpDone

	// 等待所有 VU 完成
	s.waitForAllVUs()

	close(s.doneCh)
	return nil
}

// buildRampUpPlan 计算 Ramp-up 计划
type planInput struct {
	totalVUs int
	rampUp   *RampUp
	duration time.Duration
}

// rampUpStep 单个 Ramp-up 步骤
type rampUpStep struct {
	at      time.Duration // 在压测开始后多久启动
	vuCount int           // 启动几个 VU
}

func (s *Scheduler) buildRampUpPlan(input planInput) []rampUpStep {
	if input.rampUp == nil || input.rampUp.Duration <= 0 || input.rampUp.Type == "none" {
		// 无 Ramp-up:立即启动所有 VU
		return []rampUpStep{{at: 0, vuCount: input.totalVUs}}
	}

	rampDur := input.rampUp.Duration

	switch input.rampUp.Type {
	case "linear":
		// 线性:均匀分配
		steps := 20 // 默认 20 步
		if input.rampUp.Steps > 0 {
			steps = input.rampUp.Steps
		}
		return linearPlan(input.totalVUs, rampDur, steps)

	case "step":
		// 阶梯:在 4 个时间点均匀启动
		steps := 4
		if input.rampUp.Steps > 0 {
			steps = input.rampUp.Steps
		}
		return linearPlan(input.totalVUs, rampDur, steps)

	case "wave":
		// 波浪:类似阶梯但更密集
		steps := 8
		if input.rampUp.Steps > 0 {
			steps = input.rampUp.Steps
		}
		return linearPlan(input.totalVUs, rampDur, steps)

	default:
		return []rampUpStep{{at: 0, vuCount: input.totalVUs}}
	}
}

// linearPlan 线性分配 VU 启动时间
func linearPlan(totalVUs int, duration time.Duration, steps int) []rampUpStep {
	if steps <= 0 {
		steps = 1
	}
	vusPerStep := totalVUs / steps
	remainder := totalVUs % steps
	stepDuration := duration / time.Duration(steps)

	plan := make([]rampUpStep, 0, steps)
	for i := 0; i < steps; i++ {
		count := vusPerStep
		if i == steps-1 {
			count += remainder // 最后一步承担余数
		}
		plan = append(plan, rampUpStep{
			at:      time.Duration(i) * stepDuration,
			vuCount: count,
		})
	}
	return plan
}

// executeRampUp 执行 Ramp-up 计划
func (s *Scheduler) executeRampUp(ctx context.Context, plan []rampUpStep) {
	for _, step := range plan {
		// 检查是否已停止
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		default:
		}

		// 等待到指定时间
		targetAt := s.startedAt.Add(step.at)
		waitDur := time.Until(targetAt)
		if waitDur > 0 {
			select {
			case <-time.After(waitDur):
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			}
		}

		// 启动 step.vuCount 个 VU
		for i := 0; i < step.vuCount; i++ {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			default:
				s.spawnVU()
			}
		}
	}
}

// spawnVU 启动一个新 VU
func (s *Scheduler) spawnVU() {
	s.activeVUs.Add(1)
	vuID := int(s.activeVUs.Load())

	stopCh := make(chan struct{})
	s.mu.Lock()
	s.vuStopChs = append(s.vuStopChs, stopCh)
	s.mu.Unlock()

	go func() {
		defer s.activeVUs.Add(-1)
		defer s.callback.OnVUStop(vuID)

		// 处理暂停逻辑
		if err := s.waitForResume(); err != nil {
			return
		}

		// 调用 VU 执行
		_ = s.callback.OnVUStart(vuID)
	}()
}

// waitForResume 等待(如果暂停)直到恢复
func (s *Scheduler) waitForResume() error {
	for {
		s.mu.Lock()
		paused := s.paused
		s.mu.Unlock()

		if !paused {
			return nil
		}

		// 等待 resume 信号
		select {
		case <-s.resumeCh:
			return nil
		case <-s.stopCh:
			return context.Canceled
		}
	}
}

// stopAllVUs 停止所有 VU
func (s *Scheduler) stopAllVUs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 发送停止信号
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}

	// 通知每个 VU 停止
	for _, ch := range s.vuStopChs {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
}

// waitForAllVUs 等待所有 VU 完成
func (s *Scheduler) waitForAllVUs() {
	for s.activeVUs.Load() > 0 {
		time.Sleep(50 * time.Millisecond)
	}
}

// Pause 暂停调度器
func (s *Scheduler) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = true
}

// Resume 恢复调度器
func (s *Scheduler) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.paused {
		s.paused = false
		// 通知所有等待 resume 的 VU
		select {
		case s.resumeCh <- struct{}{}:
		default:
		}
	}
}

// Stop 公开方法:外部调用以停止调度器
func (s *Scheduler) Stop() {
	s.stopAllVUs()
}

// Done 返回 done channel
func (s *Scheduler) Done() <-chan struct{} {
	return s.doneCh
}

// ActiveVUs 当前活跃 VU 数
func (s *Scheduler) ActiveVUs() int64 {
	return s.activeVUs.Load()
}

// IsPaused 是否暂停
func (s *Scheduler) IsPaused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paused
}