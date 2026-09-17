package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================
// 端到端压测基准测试套件
// ============================================

// TestScenario_HelloWorld 基础场景:hello world,无任何处理
func TestScenario_HelloWorld(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           1000,
		KeepAlive:        true,
		EnableHTTP2:      false,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := makeSimpleGETRun(1, server.URL+"/", 100, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	printSummary(t, "Hello World", summary)
	assertPerformance(t, "Hello World", summary, 5000, 50)
}

// TestScenario_JSON 模拟 JSON 响应场景
func TestScenario_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"alice","email":"alice@example.com"}`))
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           500,
		KeepAlive:        true,
		EnableHTTP2:      false,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := makeSimpleGETRun(2, server.URL+"/api/user", 50, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	printSummary(t, "JSON API", summary)
	assertPerformance(t, "JSON API", summary, 3000, 100)
}

// TestScenario_POSTLogin 模拟登录(POST + 简单处理)
func TestScenario_POSTLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0,"data":{"token":"abc123"}}`))
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           200,
		KeepAlive:        true,
		EnableHTTP2:      false,
		SnapshotInterval: 100 * time.Millisecond,
	})

	run := &Run{
		ID: 3,
		Scenario: Scenario{
			Name: "POST Login",
			Requests: []Request{
				{
					Name:   "Login",
					Method: "POST",
					URL:    server.URL + "/login",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body:   map[string]interface{}{"user": "alice", "pass": "123456"},
					Assertions: Assertions{Status: []int{200}},
				},
			},
			Load: LoadConfig{
				VUs:      50,
				Duration: 2 * time.Second,
				RampUp:   &RampUp{Type: "linear", Duration: 100 * time.Millisecond},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	printSummary(t, "POST Login", summary)
	assertPerformance(t, "POST Login", summary, 2000, 200)
}

// ============================================
// 并发能力基准测试
// ============================================

// TestScenario_ConcurrentStress 高并发测试(目标:验证 1k+ VUs 稳定)
func TestScenario_ConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewLocal(Config{
		MaxVUs:           5000,
		KeepAlive:        true,
		EnableHTTP2:      false,
		SnapshotInterval: 500 * time.Millisecond,
	})

	run := makeSimpleGETRun(4, server.URL+"/", 1000, 3*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := engine.Start(ctx, run); err != nil {
		t.Fatalf("start: %v", err)
	}

	summary, err := engine.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	printSummary(t, "1000 VUs Stress", summary)
	t.Logf("Server received: %d requests", requestCount.Load())

	// 1000 VUs × 3s:应该达到相当高的 RPS
	assertPerformance(t, "1000 VUs", summary, 20000, 500)

	// 打印 Go runtime 统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Memory: Alloc=%d MB, Sys=%d MB, GC=%d",
		m.Alloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

// ============================================
// 基准函数(可独立跑)
// ============================================

// BenchmarkEngine_LocalhostRPS localhost RPS 基准
func BenchmarkEngine_LocalhostRPS(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	for i := 0; i < b.N; i++ {
		engine := NewLocal(Config{
			MaxVUs:           200,
			KeepAlive:        true,
			EnableHTTP2:      false,
			SnapshotInterval: time.Second,
		})

		run := makeSimpleGETRun(int64(i), server.URL, 100, 200*time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

		if err := engine.Start(ctx, run); err != nil {
			b.Fatal(err)
		}

		_, err := engine.Wait()
		if err != nil {
			b.Fatal(err)
		}
		cancel()
	}
}

// BenchmarkEngine_HighConcurrency 1k 并发基准
func BenchmarkEngine_HighConcurrency(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	for i := 0; i < b.N; i++ {
		engine := NewLocal(Config{
			MaxVUs:           2000,
			KeepAlive:        true,
			EnableHTTP2:      false,
			SnapshotInterval: time.Second,
		})

		run := makeSimpleGETRun(int64(i), server.URL, 1000, 500*time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		if err := engine.Start(ctx, run); err != nil {
			b.Fatal(err)
		}

		_, err := engine.Wait()
		if err != nil {
			b.Fatal(err)
		}
		cancel()
	}
}

// BenchmarkEngine_RampUp Ramp-up 启动开销
func BenchmarkEngine_RampUp(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	for i := 0; i < b.N; i++ {
		engine := NewLocal(Config{
			MaxVUs:           500,
			KeepAlive:        true,
			EnableHTTP2:      false,
			SnapshotInterval: time.Second,
		})

		run := makeSimpleGETRun(int64(i), server.URL, 100, 100*time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)

		if err := engine.Start(ctx, run); err != nil {
			b.Fatal(err)
		}

		_, err := engine.Wait()
		if err != nil {
			b.Fatal(err)
		}
		cancel()
	}
}

// ============================================
// 工具函数
// ============================================

// makeSimpleGETRun 创建一个简单的 GET 场景
func makeSimpleGETRun(id int64, url string, vus int, duration time.Duration) *Run {
	return &Run{
		ID: id,
		Scenario: Scenario{
			Name: "Simple GET",
			Requests: []Request{
				{
					Name:    "Test",
					Method:  "GET",
					URL:     url,
					Headers: map[string]string{"User-Agent": "Pulse-Bench/1.0"},
				},
			},
			Load: LoadConfig{
				VUs:      vus,
				Duration: duration,
				RampUp: &RampUp{
					Type:     "linear",
					Duration: 50 * time.Millisecond,
				},
			},
		},
	}
}

// printSummary 打印性能摘要
func printSummary(t *testing.T, name string, summary *RunSummary) {
	t.Logf("═══════════════════════════════════════")
	t.Logf("  Scenario: %s", name)
	t.Logf("═══════════════════════════════════════")
	t.Logf("  Total Requests:  %d", summary.TotalRequests)
	t.Logf("  Total Errors:    %d", summary.TotalErrors)
	t.Logf("  Error Rate:      %.2f%%", summary.ErrorRate*100)
	t.Logf("  Duration:        %d ms", summary.DurationMs)
	t.Logf("  Avg RPS:         %.0f", summary.AvgRPS)
	t.Logf("───────────────────────────────────────")
	t.Logf("  Latency (ms):")
	t.Logf("    Min:    %.2f", summary.LatencyMs.Min)
	t.Logf("    Avg:    %.2f", summary.LatencyMs.Avg)
	t.Logf("    P50:    %.2f", summary.LatencyMs.P50)
	t.Logf("    P90:    %.2f", summary.LatencyMs.P90)
	t.Logf("    P95:    %.2f", summary.LatencyMs.P95)
	t.Logf("    P99:    %.2f", summary.LatencyMs.P99)
	t.Logf("    Max:    %.2f", summary.LatencyMs.Max)
	t.Logf("───────────────────────────────────────")
	t.Logf("  Status Codes:")
	for code, count := range summary.StatusCodes {
		t.Logf("    %s: %d", code, count)
	}
	t.Logf("═══════════════════════════════════════")

	// 输出 JSON 格式(便于脚本处理)
	if data, err := json.MarshalIndent(summary, "", "  "); err == nil {
		t.Logf("JSON: %s", string(data))
	}
}

// assertPerformance 断言性能基线
func assertPerformance(t *testing.T, name string, summary *RunSummary, minRPS, maxP95 float64) {
	if summary.AvgRPS < minRPS {
		t.Errorf("[%s] Avg RPS = %.0f, expected >= %.0f", name, summary.AvgRPS, minRPS)
	}
	if summary.P95Ms > maxP95 {
		t.Errorf("[%s] P95 = %.2fms, expected <= %.0fms", name, summary.P95Ms, maxP95)
	}
	if summary.ErrorRate > 0.01 {
		t.Errorf("[%s] Error rate = %.2f%%, expected <= 1%%", name, summary.ErrorRate*100)
	}
}

// 防止未使用 import
var (
	_ = fmt.Sprintf
	_ = sync.Mutex{}
)