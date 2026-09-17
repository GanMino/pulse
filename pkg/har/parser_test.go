package har

import (
	"strings"
	"testing"
)

// 示例 HAR 数据
const sampleHAR = `{
  "log": {
    "version": "1.2",
    "creator": {"name": "Chrome", "version": "120.0"},
    "entries": [
      {
        "startedDateTime": "2024-01-01T00:00:00.000Z",
        "time": 100,
        "request": {
          "method": "GET",
          "url": "https://api.example.com/users",
          "httpVersion": "HTTP/1.1",
          "headers": [
            {"name": "Host", "value": "api.example.com"},
            {"name": "User-Agent", "value": "Chrome"},
            {"name": "Accept", "value": "application/json"}
          ],
          "queryString": [],
          "postData": null,
          "headersSize": 100,
          "bodySize": 0
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "headers": [],
          "content": {
            "size": 50,
            "mimeType": "application/json",
            "text": "{\"users\":[]}"
          },
          "redirectURL": "",
          "headersSize": 0,
          "bodySize": 50
        },
        "cache": {},
        "serverIPAddress": "127.0.0.1",
        "connection": "1234"
      },
      {
        "startedDateTime": "2024-01-01T00:00:01.000Z",
        "time": 150,
        "request": {
          "method": "POST",
          "url": "https://api.example.com/login",
          "httpVersion": "HTTP/1.1",
          "headers": [
            {"name": "Content-Type", "value": "application/json"},
            {"name": "Cookie", "value": "session=abc"},
            {"name": "User-Agent", "value": "Chrome"}
          ],
          "queryString": [],
          "postData": {
            "mimeType": "application/json",
            "text": "{\"username\":\"alice\",\"password\":\"123456\"}"
          },
          "headersSize": 100,
          "bodySize": 50
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "headers": [],
          "content": {
            "size": 80,
            "mimeType": "application/json",
            "text": "{\"code\":0,\"data\":{\"token\":\"abc123\",\"user\":{\"id\":1}}}"
          },
          "redirectURL": "",
          "headersSize": 0,
          "bodySize": 80
        },
        "cache": {},
        "serverIPAddress": "127.0.0.1",
        "connection": "1235"
      }
    ]
  }
}`

func TestParse_ValidHAR(t *testing.T) {
	har, err := Parse([]byte(sampleHAR))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("version = %q, want 1.2", har.Log.Version)
	}
	if len(har.Log.Entries) != 2 {
		t.Errorf("entries = %d, want 2", len(har.Log.Entries))
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParse_MissingVersion(t *testing.T) {
	data := `{"log": {"entries": []}}`
	_, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestToScenario_BasicConversion(t *testing.T) {
	har, err := Parse([]byte(sampleHAR))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	opts := ConvertOptions{
		IncludeBrowserHeaders: false,
		Deduplicate:           false,
		DefaultVUs:            20,
		DefaultDuration:       "2m",
	}

	scenario := har.ToScenario(opts)
	if len(scenario.Requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(scenario.Requests))
	}

	// 检查第一个请求(GET /users)
	r0 := scenario.Requests[0]
	if r0.Method != "GET" {
		t.Errorf("r0 method = %q, want GET", r0.Method)
	}
	if !strings.HasSuffix(r0.URL, "/users") {
		t.Errorf("r0 url = %q", r0.URL)
	}
	// Cookie 应该被过滤掉
	for k := range r0.Headers {
		if strings.ToLower(k) == "cookie" {
			t.Error("Cookie header should be filtered")
		}
		if strings.ToLower(k) == "user-agent" {
			t.Error("User-Agent header should be filtered")
		}
	}

	// 检查第二个请求(POST /login)
	r1 := scenario.Requests[1]
	if r1.Method != "POST" {
		t.Errorf("r1 method = %q, want POST", r1.Method)
	}
	if r1.Body == nil {
		t.Fatal("body should not be nil for JSON POST")
	}
	if r1.Body["username"] != "alice" {
		t.Errorf("body username = %v", r1.Body["username"])
	}
}

func TestToScenario_Deduplication(t *testing.T) {
	har, err := Parse([]byte(sampleHAR))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 同一 HAR 再添加一个重复请求
	har.Log.Entries = append(har.Log.Entries, har.Log.Entries[0])

	opts := ConvertOptions{
		Deduplicate: true,
	}

	scenario := har.ToScenario(opts)
	if len(scenario.Requests) != 2 {
		t.Errorf("requests = %d, want 2 (dedup)", len(scenario.Requests))
	}
}

func TestToScenario_BaseURLReplace(t *testing.T) {
	har, err := Parse([]byte(sampleHAR))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	opts := ConvertOptions{
		BaseURL: "https://api.example.com",
	}

	scenario := har.ToScenario(opts)
	for _, r := range scenario.Requests {
		if strings.HasPrefix(r.URL, "https://") {
			t.Errorf("URL not replaced: %s", r.URL)
		}
		if !strings.HasPrefix(r.URL, "/") {
			t.Errorf("URL should start with /, got: %s", r.URL)
		}
	}
}

func TestToScenario_AutoExtractors(t *testing.T) {
	har, err := Parse([]byte(sampleHAR))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	scenario := har.ToScenario(DefaultOptions())
	loginReq := scenario.Requests[1] // POST /login

	if len(loginReq.Extractors) == 0 {
		t.Error("expected auto-extractors from login response")
	}
	t.Logf("auto extractors: %v", loginReq.Extractors)
}

func TestGenerateRequestName(t *testing.T) {
	tests := []struct {
		url    string
		method string
		want   string
	}{
		{"/api/users", "GET", "GET api/users"},
		{"/api/users/123", "GET", "GET api/users/:id"},
		{"/", "POST", "POST"},
		{"/api/orders/456/items/789", "DELETE", "DELETE api/orders/:id/items/:id"},
	}

	for _, tt := range tests {
		got := generateRequestName(tt.url, tt.method, 0)
		if got != tt.want {
			t.Errorf("generateRequestName(%q, %q) = %q, want %q", tt.url, tt.method, got, tt.want)
		}
	}
}