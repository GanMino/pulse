package engine

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// NewHTTPClient 创建针对压测场景优化的 HTTP 客户端
// 关键优化:
//  1. 启用连接池(避免每次新建连接)
//  2. 启用 Keep-Alive(连接复用)
//  3. 启用 HTTP/2(多路复用)
//  4. 优化超时配置(避免慢响应拖垮整体)
func NewHTTPClient(cfg Config) *http.Client {
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ResponseTimeout == 0 {
		cfg.ResponseTimeout = 30 * time.Second
	}
	if cfg.TLSHandshakeTimeout == 0 {
		cfg.TLSHandshakeTimeout = 5 * time.Second
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: 30 * time.Second, // Keep-Alive 探测间隔
		}).DialContext,

		// 连接池配置
		MaxIdleConns:        1000,               // 全局最大空闲连接
		MaxIdleConnsPerHost: 200,                // 每个 host 的最大空闲连接
		MaxConnsPerHost:     0,                  // 不限制每个 host 的最大连接数(默认 0 = 无限制)
		IdleConnTimeout:     90 * time.Second,   // 空闲连接超时

		// TLS 配置
		TLSHandshakeTimeout: cfg.TLSHandshakeTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // MVP 不跳过证书验证
			MinVersion:         tls.VersionTLS12,
		},

		// 响应头超时
		ResponseHeaderTimeout: 10 * time.Second,

		// 期望继续超时
		ExpectContinueTimeout: 1 * time.Second,

		// HTTP/2
		ForceAttemptHTTP2: cfg.EnableHTTP2,

		// 禁用压缩(MVP:避免压缩开销影响测量)
		DisableCompression: false,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.ResponseTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 不跟随重定向(简化压测场景)
			return http.ErrUseLastResponse
		},
	}
}