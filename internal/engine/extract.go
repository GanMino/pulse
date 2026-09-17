package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// JSONPath 极简版 JSONPath 实现
// 支持语法:
//   $                  - 根对象
//   $.field            - 取字段
//   $.field.subfield   - 多级取字段
//   $.array[0]         - 数组索引
//   $.array[*]         - 数组所有元素(返回数组)
//
// 复杂度:O(n) 遍历
// 不支持:过滤、通配符、递归

// ExtractJSONPath 从 JSON 字符串中按 JSONPath 提取值
func ExtractJSONPath(jsonStr, path string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return navigate(data, path)
}

// navigate 按 JSONPath 导航
func navigate(data interface{}, path string) (interface{}, error) {
	if path == "$" || path == "" {
		return data, nil
	}

	// 去掉前导 $
	if !strings.HasPrefix(path, "$") {
		return nil, fmt.Errorf("JSONPath must start with $")
	}
	path = path[1:]
	if path == "" {
		return data, nil
	}

	// 去掉前导 .
	if strings.HasPrefix(path, ".") {
		path = path[1:]
	}

	if path == "" {
		return data, nil
	}

	// 分割路径段
	segments := splitPath(path)

	current := data
	for _, seg := range segments {
		// 处理数组索引 seg[0] 或 seg[*]
		if strings.HasSuffix(seg, "]") {
			openBracket := strings.Index(seg, "[")
			if openBracket < 0 {
				return nil, fmt.Errorf("invalid array syntax: %s", seg)
			}
			fieldName := seg[:openBracket]
			indexPart := seg[openBracket+1 : len(seg)-1]

			// 先取字段
			if fieldName != "" {
				child, err := getField(current, fieldName)
				if err != nil {
					return nil, err
				}
				current = child
			}

			// 处理索引
			if indexPart == "*" {
				// 返回所有元素(作为数组)
				arr, ok := current.([]interface{})
				if !ok {
					return nil, fmt.Errorf("expected array at %s", seg)
				}
				return arr, nil
			}

			idx, err := parseIndex(indexPart)
			if err != nil {
				return nil, err
			}
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at %s", seg)
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("index %d out of range (len=%d)", idx, len(arr))
			}
			current = arr[idx]
			continue
		}

		// 普通字段
		child, err := getField(current, seg)
		if err != nil {
			return nil, err
		}
		current = child
	}

	return current, nil
}

// splitPath 分割 JSONPath 段,考虑 [index] 的特殊情况
func splitPath(path string) []string {
	var segments []string
	var current strings.Builder
	i := 0
	for i < len(path) {
		ch := path[i]
		if ch == '.' && current.Len() > 0 {
			segments = append(segments, current.String())
			current.Reset()
		} else if ch == '[' {
			// 把 [ 加入当前段
			current.WriteByte(ch)
			// 找到匹配的 ]
			for i < len(path) && path[i] != ']' {
				i++
				current.WriteByte(path[i])
			}
			if i < len(path) {
				current.WriteByte(']') // 关闭 ]
				i++
			}
			segments = append(segments, current.String())
			current.Reset()
			continue
		} else {
			current.WriteByte(ch)
		}
		i++
	}
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}
	return segments
}

// parseIndex 解析数组索引(支持负数)
func parseIndex(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty index")
	}
	idx := 0
	negative := false
	for i, ch := range s {
		if i == 0 && ch == '-' {
			negative = true
			continue
		}
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid index: %s", s)
		}
		idx = idx*10 + int(ch-'0')
	}
	if negative {
		idx = -idx
	}
	return idx, nil
}

// getField 从 map 中取字段
func getField(data interface{}, field string) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot access field %q on nil", field)
	}
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Map {
		// JSON unmarshal 默认用 map[string]interface{}
		key := reflect.ValueOf(field)
		val := v.MapIndex(key)
		if !val.IsValid() {
			return nil, fmt.Errorf("field %q not found", field)
		}
		return val.Interface(), nil
	}
	if v.Kind() == reflect.Struct {
		val := v.FieldByName(field)
		if !val.IsValid() {
			return nil, fmt.Errorf("field %q not found on struct", field)
		}
		return val.Interface(), nil
	}
	return nil, fmt.Errorf("cannot access field %q on %s", field, v.Kind())
}

// ExtractFirstString 提取并转为字符串(取第一个匹配的元素)
func ExtractFirstString(data interface{}) (string, bool) {
	if data == nil {
		return "", false
	}
	switch v := data.(type) {
	case string:
		return v, true
	case float64:
		return fmt.Sprintf("%g", v), true
	case bool:
		return fmt.Sprintf("%t", v), true
	default:
		// 复杂类型,JSON 序列化
		b, err := json.Marshal(data)
		if err != nil {
			return "", false
		}
		return string(b), true
	}
}