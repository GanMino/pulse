// Package event 提供事件总线抽象(预留,具体实现使用 Wails Runtime Events)
package event

// Bus 是事件总线的抽象接口
// 当前 MVP 直接使用 Wails Runtime EventsEmit/EventsOn
// 保留此接口为未来替换实现(如 NATS、Redis Pub/Sub 等)
type Bus struct {
	// 后续可扩展:订阅者列表、事件过滤、持久化等
	closed bool
}

// NewBus 创建新的事件总线
func NewBus() *Bus {
	return &Bus{
		closed: false,
	}
}

// IsClosed 检查总线是否已关闭
func (b *Bus) IsClosed() bool {
	return b.closed
}

// Close 关闭事件总线
func (b *Bus) Close() {
	b.closed = true
}

// 预定义事件名称常量
const (
	// 应用生命周期
	EventAppReady     = "app:ready"
	EventAppShutdown  = "app:shutdown"

	// 测试运行
	EventTestStarted   = "test:started"
	EventTestPaused    = "test:paused"
	EventTestResumed   = "test:resumed"
	EventTestStopped   = "test:stopped"
	EventTestCompleted = "test:completed"
	EventTestFailed    = "test:failed"

	// 实时数据
	EventMetricUpdate = "metric:update"
	EventMetricError  = "metric:error"

	// 场景管理
	EventScenarioSaved   = "scenario:saved"
	EventScenarioDeleted = "scenario:deleted"

	// Agent 管理(预留)
	EventAgentRegistered = "agent:registered"
	EventAgentOffline    = "agent:offline"
)