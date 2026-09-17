package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestCollector_BasicRecord(t *testing.T) {
	c := NewCollector(10)

	c.Record("login", 200, 50000, 1024, 256, false)
	c.Record("login", 200, 60000, 1024, 256, false)
	c.Record("login", 500, 80000, 1024, 256, true)

	if c.TotalRequests() != 3 {
		t.Errorf("total = %d, want 3", c.TotalRequests())
	}
	if c.TotalErrors() != 1 {
		t.Errorf("errors = %d, want 1", c.TotalErrors())
	}

	expectedRate := 1.0 / 3.0
	if rate := c.ErrorRate(); rate < expectedRate-0.01 || rate > expectedRate+0.01 {
		t.Errorf("error rate = %f, want ~%f", rate, expectedRate)
	}
}

func TestCollector_StatusCodes(t *testing.T) {
	c := NewCollector(10)

	c.Record("r1", 200, 10000, 0, 0, false)
	c.Record("r1", 200, 10000, 0, 0, false)
	c.Record("r1", 404, 10000, 0, 0, true)
	c.Record("r1", 500, 10000, 0, 0, true)
	c.Record("r1", 500, 10000, 0, 0, true)

	codes := c.StatusCodes()
	if codes["200"] != 2 {
		t.Errorf("200 count = %d, want 2", codes["200"])
	}
	if codes["404"] != 1 {
		t.Errorf("404 count = %d, want 1", codes["404"])
	}
	if codes["500"] != 2 {
		t.Errorf("500 count = %d, want 2", codes["500"])
	}
}

func TestCollector_PerRequest(t *testing.T) {
	c := NewCollector(10)

	c.Record("login", 200, 50000, 0, 0, false)
	c.Record("login", 200, 60000, 0, 0, false)
	c.Record("logout", 200, 30000, 0, 0, false)

	perReq := c.PerRequestSnapshot()
	if len(perReq) != 2 {
		t.Errorf("got %d requests, want 2", len(perReq))
	}

	login := perReq["login"]
	if login.Count != 2 {
		t.Errorf("login count = %d, want 2", login.Count)
	}
	if login.Latency.P50 == 0 {
		t.Error("expected non-zero latency for login")
	}

	logout := perReq["logout"]
	if logout.Count != 1 {
		t.Errorf("logout count = %d, want 1", logout.Count)
	}
}

func TestCollector_ConcurrentRecord(t *testing.T) {
	c := NewCollector(10)

	var wg sync.WaitGroup
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				c.Record("test", 200, 50000, 0, 0, false)
			}
		}()
	}
	wg.Wait()

	if c.TotalRequests() != 10000 {
		t.Errorf("total = %d, want 10000", c.TotalRequests())
	}
}

func TestSlidingWindow_RPS(t *testing.T) {
	window := NewSlidingWindow(1 * time.Second)

	// 1 秒内发送 100 个请求
	now := time.Now()
	for i := 0; i < 100; i++ {
		window.Add(now)
	}

	rps := window.RPS()
	if rps < 80 || rps > 120 {
		t.Errorf("rps = %f, want ~100", rps)
	}
}

func TestSlidingWindow_ExpiresOldData(t *testing.T) {
	window := NewSlidingWindow(500 * time.Millisecond)

	// 注入一个很老的请求
	oldTime := time.Now().Add(-1 * time.Hour)
	window.Add(oldTime)

	// 注入一个新请求
	window.Add(time.Now())

	// 老的不应被计算
	if window.Count() > 10 {
		t.Errorf("expected old data to be excluded, got count=%d", window.Count())
	}
}