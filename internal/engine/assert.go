package engine

import (
	"bytes"
	"fmt"
)

// AssertResult 断言结果
type AssertResult struct {
	Passed bool
	Reason string // 不通过时的原因
}

// AssertResponse 对响应执行断言
func AssertResponse(a Assertions, statusCode int, latencyMs float64, body []byte) AssertResult {
	// 检查状态码
	if len(a.Status) > 0 {
		matched := false
		for _, s := range a.Status {
			if s == statusCode {
				matched = true
				break
			}
		}
		if !matched {
			return AssertResult{
				Passed: false,
				Reason: fmt.Sprintf("status %d not in expected %v", statusCode, a.Status),
			}
		}
	}

	// 检查延迟
	if a.MaxLatencyMs > 0 && latencyMs > float64(a.MaxLatencyMs) {
		return AssertResult{
			Passed: false,
			Reason: fmt.Sprintf("latency %.2fms exceeds limit %dms", latencyMs, a.MaxLatencyMs),
		}
	}

	// 检查 body 包含
	for _, s := range a.BodyContains {
		if s == "" {
			continue
		}
		if !bytes.Contains(body, []byte(s)) {
			return AssertResult{
				Passed: false,
				Reason: fmt.Sprintf("body does not contain %q", s),
			}
		}
	}

	return AssertResult{Passed: true}
}