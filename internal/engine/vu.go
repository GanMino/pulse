package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/pulse/pulse/internal/engine/metrics"
)

// VU 是单个虚拟用户(协程)的执行循环
// 每个 VU 独立持有 localVars(extractors 提取的值)
type VU struct {
	vuID      int
	scenario  Scenario
	client    *http.Client
	resolver  *VariableResolver
	limiter   RateLimiter
	collector *metrics.Collector
	paused    func() bool
}

// NewVU 创建 VU
func NewVU(
	vuID int,
	scenario Scenario,
	client *http.Client,
	resolver *VariableResolver,
	limiter RateLimiter,
	collector *metrics.Collector,
	pausedCheck func() bool,
) *VU {
	return &VU{
		vuID:      vuID,
		scenario:  scenario,
		client:    client,
		resolver:  resolver,
		limiter:   limiter,
		collector: collector,
		paused:    pausedCheck,
	}
}

// Run 执行 VU 主循环,直到 context 取消
func (v *VU) Run(ctx context.Context) error {
	// 每个 VU 持有自己的 localVars(用于 extractors)
	localVars := make(map[string]string)

	for {
		// 暂停检查
		if v.paused != nil && v.paused() {
			select {
			case <-time.After(100 * time.Millisecond):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// 遍历场景中的所有请求
		if err := v.runScenario(ctx, localVars); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
		}

		// 检查全局 context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// runScenario 执行场景中的所有请求
func (v *VU) runScenario(ctx context.Context, localVars map[string]string) error {
	for _, req := range v.scenario.Requests {
		// 检查 context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 执行单个请求
		v.executeRequest(ctx, req, localVars)

		// Think Time(请求之间)
		tt := req.ThinkTime
		if tt == 0 {
			tt = v.scenario.Load.ThinkTime
		}
		if tt > 0 {
			select {
			case <-time.After(tt):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// executeRequest 执行单个 HTTP 请求
func (v *VU) executeRequest(ctx context.Context, req Request, localVars map[string]string) {
	// 限流(目标 RPS)
	if v.limiter != nil {
		if err := v.limiter.Wait(ctx); err != nil {
			return // context 取消
		}
	}

	// 1. 解析 URL 变量
	url, err := v.resolver.Substitute(req.URL, localVars)
	if err != nil {
		v.recordError(req.Name, 0, 0, err.Error(), true)
		return
	}

	// 2. 解析 Body 变量
	var body []byte
	if len(req.Body) > 0 {
		bodyJSON, mErr := marshalRequestBody(req.Body)
		if mErr != nil {
			v.recordError(req.Name, 0, 0, mErr.Error(), true)
			return
		}
		body, err = v.resolver.SubstituteBytes(bodyJSON, localVars)
		if err != nil {
			v.recordError(req.Name, 0, 0, err.Error(), true)
			return
		}
	}
	if req.BodyRaw != "" {
		body, err = v.resolver.SubstituteBytes([]byte(req.BodyRaw), localVars)
		if err != nil {
			v.recordError(req.Name, 0, 0, err.Error(), true)
			return
		}
	}

	// 3. 解析 Headers 变量
	headers := make(map[string]string, len(req.Headers))
	for k, val := range req.Headers {
		resolvedVal, sErr := v.resolver.Substitute(val, localVars)
		if sErr != nil {
			v.recordError(req.Name, 0, 0, sErr.Error(), true)
			return
		}
		headers[k] = resolvedVal
	}

	// 4. 构造 HTTP 请求
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		v.recordError(req.Name, 0, 0, err.Error(), true)
		return
	}
	for k, val := range headers {
		httpReq.Header.Set(k, val)
	}

	// 5. 发送请求
	start := time.Now()
	resp, err := v.client.Do(httpReq)
	latencyUs := time.Since(start).Microseconds()
	latencyMs := float64(latencyUs) / 1000.0

	if err != nil {
		v.recordError(req.Name, 0, latencyMs, err.Error(), true)
		return
	}
	defer resp.Body.Close()

	// 6. 读取响应
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		v.recordError(req.Name, resp.StatusCode, latencyMs, readErr.Error(), true)
		return
	}

	// 7. 执行断言
	var assertResult AssertResult
	if len(req.Assertions.Status) > 0 || req.Assertions.MaxLatencyMs > 0 || len(req.Assertions.BodyContains) > 0 {
		assertResult = AssertResponse(req.Assertions, resp.StatusCode, latencyMs, respBody)
	}

	// 8. 执行 extractors(仅对 2xx 响应)
	if len(req.Extractors) > 0 && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		v.extractVariables(req.Extractors, respBody, localVars)
	}

	// 9. 记录指标
	isError := !assertResult.Passed || resp.StatusCode >= 400
	v.recordRequest(req.Name, resp.StatusCode, latencyUs, int64(len(body)), int64(len(respBody)), isError)
}

// extractVariables 从响应中提取变量
func (v *VU) extractVariables(extractors map[string]string, body []byte, localVars map[string]string) {
	for varName, jsonPath := range extractors {
		val, err := ExtractJSONPath(string(body), jsonPath)
		if err != nil {
			continue
		}
		if strVal, ok := ExtractFirstString(val); ok {
			localVars[varName] = strVal
		}
	}
}

// recordRequest 记录成功的请求
func (v *VU) recordRequest(name string, statusCode int, latencyUs int64, bytesIn int64, bytesOut int64, isError bool) {
	v.collector.Record(name, statusCode, latencyUs, bytesIn, bytesOut, isError)
}

// recordError 记录错误
func (v *VU) recordError(name string, statusCode int, latencyMs float64, errMsg string, isError bool) {
	v.collector.Record(name, statusCode, int64(latencyMs*1000), 0, int64(len(errMsg)), isError)
}

// marshalRequestBody 序列化 Body 为 JSON
func marshalRequestBody(body interface{}) ([]byte, error) {
	return json.Marshal(body)
}