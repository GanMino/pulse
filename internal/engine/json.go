package engine

import "encoding/json"

// jsonMarshal 是 encoding/json 的薄封装,便于以后替换(如 jsoniter)
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}