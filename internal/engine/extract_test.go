package engine

import (
 "testing"
)

func TestExtractJSONPath_Root(t *testing.T) {
	jsonStr := `{"name": "alice", "age": 30}`
	result, err := ExtractJSONPath(jsonStr, "$")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map")
	}
	if m["name"] != "alice" {
		t.Errorf("name = %v", m["name"])
	}
}

func TestExtractJSONPath_SimpleField(t *testing.T) {
	jsonStr := `{"name": "alice", "age": 30}`
	result, err := ExtractJSONPath(jsonStr, "$.name")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if result != "alice" {
		t.Errorf("got %v, want alice", result)
	}
}

func TestExtractJSONPath_NestedField(t *testing.T) {
	jsonStr := `{"data": {"user": {"id": 123, "name": "alice"}}}`
	result, err := ExtractJSONPath(jsonStr, "$.data.user.name")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if result != "alice" {
		t.Errorf("got %v", result)
	}
}

func TestExtractJSONPath_ArrayIndex(t *testing.T) {
	jsonStr := `{"users": [{"id": 1},{"id": 2},{"id": 3}]}`
	result, err := ExtractJSONPath(jsonStr, "$.users[1].id")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	// JSON numbers unmarshal to float64
	if result != float64(2) {
		t.Errorf("got %v (%T), want 2", result, result)
	}
}

func TestExtractJSONPath_ArrayAll(t *testing.T) {
	jsonStr := `{"items": [1, 2, 3]}`
	result, err := ExtractJSONPath(jsonStr, "$.items[*]")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("len = %d, want 3", len(arr))
	}
}

func TestExtractJSONPath_NotFound(t *testing.T) {
	jsonStr := `{"name": "alice"}`
	_, err := ExtractJSONPath(jsonStr, "$.foo")
	if err == nil {
		t.Error("expected error for missing field")
	}
}

func TestExtractJSONPath_InvalidJSON(t *testing.T) {
	_, err := ExtractJSONPath("not json", "$.foo")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExtractJSONPath_TokenExample(t *testing.T) {
	jsonStr := `{"code":0,"data":{"token":"abc123","user":{"id":1}}}`
	token, err := ExtractJSONPath(jsonStr, "$.data.token")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if token != "abc123" {
		t.Errorf("token = %v", token)
	}
}

func TestVariableResolver_Substitute(t *testing.T) {
	resolver := NewVariableResolver(map[string][]string{
		"user": {"alice", "bob"},
		"pass": {"123456"},
	})

	tests := []struct {
		input    string
		localVars map[string]string
		wantErr  bool
	}{
		{"Hello {{user}}", nil, false},
		{"User: {{user}}, Pass: {{pass}}", nil, false},
		{"Token: {{token}}", nil, true}, // 不存在的变量
		{"Token: {{token}}", map[string]string{"token": "abc"}, false}, // local 优先
	}

	for _, tt := range tests {
		_, err := resolver.Substitute(tt.input, tt.localVars)
		if (err != nil) != tt.wantErr {
			t.Errorf("Substitute(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestVariableResolver_LocalPriority(t *testing.T) {
	resolver := NewVariableResolver(map[string][]string{
		"token": {"global_token"},
	})

	result, err := resolver.Substitute("{{token}}", map[string]string{
		"token": "local_token",
	})
	if err != nil {
		t.Fatalf("substitute: %v", err)
	}
	if result != "local_token" {
		t.Errorf("got %q, want local_token (local should take priority)", result)
	}
}

func TestVariableResolver_RandomForMultiple(t *testing.T) {
	resolver := NewVariableResolver(map[string][]string{
		"id": {"1", "2", "3", "4", "5"},
	})

	// 多次调用,应该随机返回
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		result, err := resolver.Substitute("{{id}}", nil)
		if err != nil {
			t.Fatalf("substitute: %v", err)
		}
		seen[result] = true
	}

	// 至少应该出现多个不同的值(由于随机性)
	if len(seen) < 2 {
		t.Errorf("expected random values, got only %d unique: %v", len(seen), seen)
	}
}