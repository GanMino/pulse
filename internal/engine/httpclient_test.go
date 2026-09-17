package engine

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	cfg := Config{
		KeepAlive:           true,
		EnableHTTP2:         true,
		DialTimeout:         5 * time.Second,
		ResponseTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := NewHTTPClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != cfg.ResponseTimeout {
		t.Errorf("timeout = %v, want %v", client.Timeout, cfg.ResponseTimeout)
	}
	if client.Transport == nil {
		t.Fatal("expected non-nil transport")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if !transport.ForceAttemptHTTP2 {
		t.Error("expected HTTP/2 to be enabled")
	}
	if transport.MaxIdleConnsPerHost != 200 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 200", transport.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout == 0 {
		t.Error("expected non-zero IdleConnTimeout")
	}
}

func TestNewHTTPClient_DefaultValues(t *testing.T) {
	// 不提供任何参数,应该使用默认值
	client := NewHTTPClient(Config{})
	if client.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", client.Timeout)
	}
}