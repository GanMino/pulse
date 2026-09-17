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

// TestAssertResponse_Basic 测试断言逻辑
func TestAssertResponse_Basic(t *testing.T) {
	tests := []struct {
		name       string
		assertions Assertions
		status     int
		latency    float64
		body       []byte
		wantPassed bool
		wantReason string
	}{
		{
			name: "StatusMatch",
			assertions: Assertions{Status: []int{200}},
			status: 200,
			wantPassed: true,
		},
		{
			name: "StatusMismatch",
			assertions: Assertions{Status: []int{200}},
			status: 500,
			wantPassed: false,
			wantReason: "status",
		},
		{
			name: "LatencyUnderLimit",
			assertions: Assertions{MaxLatencyMs: 200},
			status: 200,
			latency: 100,
			wantPassed: true,
		},
		{
			name: "LatencyOverLimit",
			assertions: Assertions{MaxLatencyMs: 100},
			status: 200,
			latency: 200,
			wantPassed: false,
			wantReason: "latency",
		},
		{
			name: "BodyContains_Pass",
			assertions: Assertions{BodyContains: []string{"success"}},
			status: 200,
			body: []byte(`{"result":"success"}`),
			wantPassed: true,
		},
		{
			name: "BodyContains_Fail",
			assertions: Assertions{BodyContains: []string{"success"}},
			status: 200,
			body: []byte(`{"result":"error"}`),
			wantPassed: false,
			wantReason: "body",
		},
		{
			name: "NoAssertions",
			assertions: Assertions{},
			status: 500,
			wantPassed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AssertResponse(tt.assertions, tt.status, tt.latency, tt.body)
			if result.Passed != tt.wantPassed {
				t.Errorf("passed = %v, want %v (reason: %s)", result.Passed, tt.wantPassed, result.Reason)
			}
		})
	}
}

// TestScheduler_RampUpPlan 测试 Ramp-up 计划计算
func TestScheduler_RampUpPlan(t *testing.T) {
	tests := []struct {
		name       string
		rampUp     *RampUp
		totalVUs   int
		wantSteps  int
		wantTotal  int
	}{
		{
			name: "Linear10Steps",
			rampUp: &RampUp{Type: "linear", Duration: 30 * time.Second, Steps: 10},
			totalVUs: 100,
			wantSteps: 10,
			wantTotal: 100,
		},
		{
			name: "Step5Steps",
			rampUp: &RampUp{Type: "step", Duration: 10 * time.Second, Steps: 5},
			totalVUs: 50,
			wantSteps: 5,
			wantTotal: 50,
		},
		{
			name: "Wave",
			rampUp: &RampUp{Type: "wave", Duration: 30 * time.Second},
			totalVUs: 200,
			wantSteps: 8, // 默认 8 步
			wantTotal: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScheduler(tt.totalVUs, LoadConfig{}, &mockCallback{})
			plan := s.buildRampUpPlan(planInput{
				totalVUs: tt.totalVUs,
				rampUp:   tt.rampUp,
				duration: 1 * time.Minute,
			})

			if len(plan) != tt.wantSteps {
				t.Errorf("steps = %d, want %d", len(plan), tt.wantSteps)
			}
			total := 0
			for _, step := range plan {
				total += step.vuCount
			}
			if total != tt.wantTotal {
				t.Errorf("total VUs = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

type mockCallback struct {
	startedCount atomic.Int64
	stoppedCount atomic.Int64
}

func (m *mockCallback) OnVUStart(vuID int) error {
	m.startedCount.Add(1)
	return nil
}

func (m *mockCallback) OnVUStop(vuID int) {
	m.stoppedCount.Add(1)
}

// BenchmarkLocalEngine_SimpleLoad 性能基准测试
func BenchmarkLocalEngine_SimpleLoad(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine := NewLocal(Config{
			MaxVUs:           50,
			KeepAlive:        true,
			EnableHTTP2:      true,
			SnapshotInterval: 100 * time.Millisecond,
		})

		run := &Run{
			ID: int64(i),
			Scenario: Scenario{
				Name: "Bench",
				Requests: []Request{
					{Name: "Test", Method: "GET", URL: server.URL},
				},
				Load: LoadConfig{
					VUs:      10,
					Duration: 200 * time.Millisecond,
					RampUp: &RampUp{
						Type: "linear",
						Duration: 50 * time.Millisecond,
					},
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = engine.Start(ctx, run)
		_, _ = engine.Wait()
		cancel()
	}
}