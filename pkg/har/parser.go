// Package har 提供 HAR (HTTP Archive) 文件解析
// 用于从浏览器导出的 HAR 文件生成压测场景
package har

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pulse/pulse/internal/model"
)

// HAR 表示 HAR 1.2 文件结构
// 参考规范: http://www.softwareishard.com/blog/har-12-spec/
type HAR struct {
	Log Log `json:"log"`
}

// Log 是 HAR 的日志条目
type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Entries []Entry `json:"entries"`
}

// Creator 是 HAR 创建工具信息
type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Entry 是单个 HTTP 请求/响应条目
type Entry struct {
	StartedDateTime time.Time `json:"startedDateTime"`
	Time            float64   `json:"time"`
	Request         Request   `json:"request"`
	Response        Response  `json:"response"`
	Cache           Cache     `json:"cache"`
	ServerIPAddress string    `json:"serverIPAddress"`
	Connection      string    `json:"connection"`
}

// Request 是 HAR 中的请求
type Request struct {
	Method      string   `json:"method"`
	URL         string   `json:"url"`
	HTTPVersion string   `json:"httpVersion"`
	Headers     []Header `json:"headers"`
	QueryString []NameValue `json:"queryString"`
	PostData    *PostData `json:"postData,omitempty"`
	HeadersSize int      `json:"headersSize"`
	BodySize    int      `json:"bodySize"`
}

// Response 是 HAR 中的响应
type Response struct {
	Status      int      `json:"status"`
	StatusText  string   `json:"statusText"`
	HTTPVersion string   `json:"httpVersion"`
	Headers     []Header `json:"headers"`
	Content     Content  `json:"content"`
	RedirectURL string   `json:"redirectURL"`
	HeadersSize int      `json:"headersSize"`
	BodySize    int      `json:"bodySize"`
}

// Header 是 HAR 中的 header
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NameValue 是 URL query string 或 form data
type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PostData 是请求体
type PostData struct {
	MimeType string       `json:"mimeType"`
	Text     string       `json:"text"`
	Params   []NameValue  `json:"params,omitempty"`
}

// Content 是响应体
type Content struct {
	Size        int    `json:"size"`
	Compression int    `json:"compression"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text"`
	Encoding    string `json:"encoding"` // "base64" 表示二进制
}

// Cache 是缓存状态
type Cache struct {
	BeforeRequest *CacheState `json:"beforeRequest,omitempty"`
	AfterRequest  *CacheState `json:"afterRequest,omitempty"`
}

// CacheState 是缓存状态
type CacheState struct {
	Expires        string `json:"expires"`
	LastAccess     string `json:"lastAccess"`
	ETag           string `json:"eTag"`
	HitCount       int    `json:"hitCount"`
	ExpirationTime string `json:"expirationTime"`
}

// Parse 解析 HAR 字节数组
func Parse(data []byte) (*HAR, error) {
	var har HAR
	if err := json.Unmarshal(data, &har); err != nil {
		return nil, fmt.Errorf("invalid HAR JSON: %w", err)
	}
	if har.Log.Version == "" {
		return nil, fmt.Errorf("missing HAR log.version")
	}
	return &har, nil
}

// ParseFile 解析 HAR 文件
func ParseFile(path string) (*HAR, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// 浏览器特定 header 列表(默认排除)
var browserHeaders = map[string]bool{
	":authority":          true,
	":method":             true,
	":path":               true,
	":scheme":             true,
	"host":                true, // 由 URL 推断
	"cookie":              true,
	"user-agent":          true,
	"sec-ch-ua":           true,
	"sec-ch-ua-mobile":    true,
	"sec-ch-ua-platform":  true,
	"sec-fetch-dest":      true,
	"sec-fetch-mode":      true,
	"sec-fetch-site":      true,
	"sec-fetch-user":      true,
	"upgrade-insecure-requests": true,
	"accept-encoding":     true,
	"accept-language":     true,
	"dnt":                 true,
	"connection":          true,
	"referer":             true,
	"origin":              true,
}

// ConvertOptions HAR 转换选项
type ConvertOptions struct {
	// IncludeBrowserHeaders 是否保留浏览器特定 header
	IncludeBrowserHeaders bool

	// Deduplicate 是否去重(同 method+URL+body 合并)
	Deduplicate bool

	// DefaultVUs 默认 VUs
	DefaultVUs int

	// DefaultDuration 默认时长
	DefaultDuration string

	// BaseURL 替换 URL 的 base 部分(如 "https://api.example.com")
	// 留空表示保留原始 URL
	BaseURL string
}

// DefaultOptions 返回默认选项
func DefaultOptions() ConvertOptions {
	return ConvertOptions{
		IncludeBrowserHeaders: false,
		Deduplicate:           true,
		DefaultVUs:            10,
		DefaultDuration:       "1m",
	}
}

// ToScenario 将 HAR 转换为 ScenarioConfig
func (h *HAR) ToScenario(opts ConvertOptions) *model.ScenarioConfig {
	if opts.DefaultVUs <= 0 {
		opts.DefaultVUs = 10
	}
	if opts.DefaultDuration == "" {
		opts.DefaultDuration = "1m"
	}

	scenario := &model.ScenarioConfig{
		Load: model.LoadConfig{
			VUs:      opts.DefaultVUs,
			Duration: opts.DefaultDuration,
		},
		Variables: make(map[string]model.VariableValue),
	}

	seen := make(map[string]bool)
	requestIndex := 0

	for _, entry := range h.Log.Entries {
		req := entry.Request

		// URL 处理:提取 base
		finalURL := req.URL
		if opts.BaseURL != "" && strings.HasPrefix(finalURL, opts.BaseURL) {
			finalURL = "/" + strings.TrimPrefix(finalURL, opts.BaseURL)
			finalURL = strings.TrimPrefix(finalURL, "//")
		}

		// 去重
		if opts.Deduplicate {
			key := fmt.Sprintf("%s|%s|%s", req.Method, finalURL, req.PostData.Text)
			if seen[key] {
				continue
			}
			seen[key] = true
		}

		// Headers 转换
		headers := make(map[string]string)
		for _, h := range req.Headers {
			if !opts.IncludeBrowserHeaders && browserHeaders[strings.ToLower(h.Name)] {
				continue
			}
			headers[h.Name] = h.Value
		}

		// Body 转换
		bodyObj := parseBody(req.PostData)
		bodyRaw := ""
		if req.PostData != nil && bodyObj == nil {
			bodyRaw = req.PostData.Text
		}

		// 提取请求名(从 URL 推断)
		name := generateRequestName(finalURL, req.Method, requestIndex)

		scenario.Requests = append(scenario.Requests, model.RequestConfig{
			Name:       name,
			Method:     req.Method,
			URL:        finalURL,
			Headers:    headers,
			Body:       bodyObj,
			BodyRaw:    bodyRaw,
			Extractors: autoExtractors(entry.Response, req.Method),
		})

		requestIndex++
	}

	return scenario
}

// parseBody 解析请求体
func parseBody(pd *PostData) map[string]interface{} {
	if pd == nil || pd.Text == "" {
		return nil
	}

	// 尝试解析为 JSON
	mimeType := strings.ToLower(pd.MimeType)
	if strings.Contains(mimeType, "json") {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(pd.Text), &result); err == nil {
			return result
		}
	}

	// 尝试 form-urlencoded
	if strings.Contains(mimeType, "x-www-form-urlencoded") || strings.Contains(mimeType, "form") {
		result := make(map[string]interface{})
		values, err := url.ParseQuery(pd.Text)
		if err == nil {
			for k, v := range values {
				if len(v) == 1 {
					result[k] = v[0]
				} else {
					result[k] = v
				}
			}
			return result
		}
	}

	return nil
}

// autoExtractors 自动生成响应提取器
// MVP:对 JSON 响应尝试提取常见的 token 字段
func autoExtractors(resp Response, method string) map[string]string {
	extractors := make(map[string]string)

	if !strings.Contains(strings.ToLower(resp.Content.MimeType), "json") {
		return extractors
	}

	text := resp.Content.Text
	if text == "" {
		return extractors
	}

	// 尝试解析 JSON
	var data interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return extractors
	}

	// 自动检测常见的 token/session 字段
	patterns := []struct {
		field   string
		jsonPath string
	}{
		{"token", "$.token"},
		{"access_token", "$.access_token"},
		{"accessToken", "$.accessToken"},
		{"data.token", "$.data.token"},
		{"data.accessToken", "$.data.accessToken"},
		{"sessionId", "$.sessionId"},
	}

	for _, p := range patterns {
		if strings.Contains(text, fmt.Sprintf("%q", p.field)) ||
			strings.Contains(text, fmt.Sprintf("\"%s\"", p.field)) {
			// 简单检测,实际 JSONPath 验证在运行时
			extractors[strings.ReplaceAll(p.jsonPath, "$.", "")] = p.jsonPath
			break // 只取第一个匹配的
		}
	}

	return extractors
}

// generateRequestName 生成请求名
func generateRequestName(urlStr, method string, index int) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Sprintf("%s_%d", method, index)
	}

	path := u.Path
	// 简化路径(如 /api/users/123 → /api/users/:id)
	path = simplifyPath(path)
	// 移除前导 /
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		return method
	}

	// 限制长度
	if len(path) > 30 {
		path = path[:30] + "..."
	}

	return fmt.Sprintf("%s %s", method, path)
}

var paramRegex = regexp.MustCompile(`/[0-9]+`)

// simplifyPath 简化路径中的数字 ID
func simplifyPath(path string) string {
	return paramRegex.ReplaceAllString(path, "/:id")
}

// readFile 读取文件
func readFile(path string) ([]byte, error) {
	return readFileOS(path)
}

// 防止 base64 未使用告警
var _ = base64.StdEncoding