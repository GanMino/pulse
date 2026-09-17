package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLocalEngine_BasicLoadTest 端到端压测测试
func TestLocalEngine_BasicLoadTest(t *testing.T) {
	// 1. 启动 mock HTTP 服务器
	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0,"data":{"token":"abc123","id":1}}`))
	}))
	defer server.Close()

	// 2. 创建引擎
	engine := NewLocal(Config{
		MaxVUs:           100,
		KeepAlive:        true,
		EnableHTTP2:      true,
		DialTimeout:      5 * time.Second,
		ResponseTimeout:  10 * time.Second,
		SnapshotInterval: 100 * time.Millisecond,
	})

	// 3. 创建测试场景
	run := &Run{
		ID: 1,
		Scenario: Scenario{
			Name: "Login Flow",
			Requests: []Request{
				{
					Name:   "Login",
					Method: "POST",
					URL:    server.URL + "/login",
					Body:   map[string]interface{}{"user": "alice", "pass": "123"},
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Extractors: map[string]string{
						"token": "$.data.token",
					},
					Assertions: Assertions{
						Status: []int{200},
					},
				},
				{
					Name:   "GetUser",
					Method: "GET",
					URL:    server.URL + "/user/1",
					Headers: map[string]string{
						"Authorization": "Bearer {{token}}",
					},
				},
			},
			Load: LoadConfig{
				VUs:      5,
				Duration: 500 * time.Millisecond,
				RampUp: &RampUp{
					Type:     "linear",
					Duration: 100 * time.Millisecond,
				},
			},
		},
	}

	// 4. 启动压测
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	// 5. 等待完成
	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	// 6. 验证结果
	if summary.TotalRequests == 0 {
		t.Error("expected some requests")
	}
	if summary.ErrorRate > 0 {
		t.Errorf("unexpected errors: %v", summary.ErrorRate)
	}
	if summary.AvgRPS <= 0 {
		t.Error("expected positive RPS")
	}

	t.Logf("Total: %d, Errors: %d, RPS: %.1f, P95: %.2fms",
		summary.TotalRequests, summary.TotalErrors, summary.AvgRPS, summary.P95Ms)
	t.Logf("Server received: %d requests", requestCount.Load())

	if requestCount.Load() == 0 {
		t.Error("server should have received requests")
	}
}

// TestLocalEngine_MultiRequestSequential 验证多请求顺序执行 + extractors
func TestLocalEngine_MultiRequestSequential(t *testing.T) {
	var loginCalls, getUserCalls atomic.Int64
	var lastAuthHeader atomic.Value

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			loginCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"code":0,"data":{"token":"abc123"}}`))
		case "/user/1":
			getUserCalls.Add(1)
			lastAuthHeader.Store(r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           10,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := &Run{
		ID: 2,
		Scenario: Scenario{
			Name: "Login then GetUser",
			Requests: []Request{
				{
					Name:   "Login",
					Method: "POST",
					URL:    server.URL + "/login",
					Body:   map[string]interface{}{"user": "alice"},
					Extractors: map[string]string{
						"token": "$.data.token",
					},
				},
				{
					Name:   "GetUser",
					Method: "GET",
					URL:    server.URL + "/user/1",
					Headers: map[string]string{
						"Authorization": "Bearer {{token}}",
					},
				},
			},
			Load: LoadConfig{
				VUs:      2,
				Duration: 300 * time.Millisecond,
				RampUp: &RampUp{
					Type:     "linear",
					Duration: 50 * time.Millisecond,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	if loginCalls.Load() == 0 {
		t.Error("login should be called")
	}
	if getUserCalls.Load() == 0 {
		t.Error("getUser should be called")
	}

	// 验证 Authorization header 是从 extractors 中提取的
	auth := lastAuthHeader.Load()
	if auth != nil && auth != "Bearer abc123" {
		t.Errorf("expected Authorization=Bearer abc123, got %v", auth)
	}

	t.Logf("Login: %d, GetUser: %d", loginCalls.Load(), getUserCalls.Load())
	t.Logf("Summary: %d reqs, %.1f RPS", summary.TotalRequests, summary.AvgRPS)
}

// TestLocalEngine_PauseResume 验证暂停/恢复
func TestLocalEngine_PauseResume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           5,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := &Run{
		ID: 3,
		Scenario: Scenario{
			Name: "Pause Test",
			Requests: []Request{
				{Name: "Test", Method: "GET", URL: server.URL},
			},
			Load: LoadConfig{
				VUs:      3,
				Duration: 1 * time.Second,
				RampUp: &RampUp{Type: "linear", Duration: 100 * time.Millisecond},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	// 等待启动
	time.Sleep(300 * time.Millisecond)

	// 暂停
	if err := engine.Pause(); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if engine.Status() != StatusPaused {
		t.Errorf("expected Paused, got %s", engine.Status())
	}

	pausedStatus := engine.Status()
	time.Sleep(200 * time.Millisecond)

	// 恢复
	if err := engine.Resume(); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if engine.Status() != StatusRunning {
		t.Errorf("expected Running, got %s", engine.Status())
	}

	// 等待完成
	_, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	_ = pausedStatus
}

// TestLocalEngine_StopBeforeCompletion 测试提前停止
func TestLocalEngine_StopBeforeCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // 模拟慢请求
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           10,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := &Run{
		ID: 4,
		Scenario: Scenario{
			Name: "Stop Test",
			Requests: []Request{
				{Name: "Test", Method: "GET", URL: server.URL},
			},
			Load: LoadConfig{
				VUs:      5,
				Duration: 10 * time.Second,
				RampUp: &RampUp{Type: "linear", Duration: 100 * time.Millisecond},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	// 启动后立刻停止
	time.Sleep(200 * time.Millisecond)
	if err := engine.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}

	// 等待完成(应该快速结束)
	done := make(chan struct{})
	go func() {
		_, _ = engine.Wait()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not stop within timeout")
	}

	if status := engine.Status(); status != StatusAborted {
		t.Errorf("expected Aborted, got %s", status)
	}
}

// TestLocalEngine_WithRateLimit 测试 RPS 限流
func TestLocalEngine_WithRateLimit(t *testing.T) {
	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           20,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := &Run{
		ID: 5,
		Scenario: Scenario{
			Name: "Rate Limit Test",
			Requests: []Request{
				{Name: "Test", Method: "GET", URL: server.URL},
			},
			Load: LoadConfig{
				VUs:       20,                // 20 个 VU
				Duration:  500 * time.Millisecond,
				TargetRPS: 100,               // 但限流到 100 RPS
				RampUp: &RampUp{
					Type:     "linear",
					Duration: 100 * time.Millisecond,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	// 100 RPS × 0.5s ≈ 50 请求(允许一些误差)
	expected := int64(40)
	if summary.TotalRequests < expected {
		t.Errorf("expected at least %d requests (100 RPS * 0.5s), got %d", expected, summary.TotalRequests)
	}

	t.Logf("Total: %d requests (expected ~50 for 100 RPS * 0.5s)", summary.TotalRequests)
}

// 防止未使用的 import 警告
var _ = sync.Mutex{}