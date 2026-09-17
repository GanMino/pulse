package engine

import (
	"fmt"
	"math/rand"
	"regexp"
)

// VariableResolver 负责变量解析与替换
// 变量语法: {{variable_name}}
// 支持从全局变量 + VU 本地变量(extractors)中查找
type VariableResolver struct {
	// globalVars 全局变量(从场景配置中)
	globalVars map[string][]string
}

// NewVariableResolver 创建变量 Resolver
func NewVariableResolver(global map[string][]string) *VariableResolver {
	if global == nil {
		global = make(map[string][]string)
	}
	return &VariableResolver{
		globalVars: global,
	}
}

// 匹配 {{var_name}} 格式
var varRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)

// Substitute 在字符串中替换 {{var_name}} 占位符
// localVars 是当前 VU 的本地变量(extractors 提取的值)
func (r *VariableResolver) Substitute(input string, localVars map[string]string) (string, error) {
	var firstErr error
	result := varRegex.ReplaceAllStringFunc(input, func(match string) string {
		// 提取变量名
		name := match[2 : len(match)-2] // 去掉 {{ 和 }}

		// 优先查找 localVars
		if localVars != nil {
			if v, ok := localVars[name]; ok {
				return v
			}
		}

		// 然后查找 globalVars
		if vs, ok := r.globalVars[name]; ok && len(vs) > 0 {
			// 多个值:随机选一个(数据驱动)
			return vs[rand.Intn(len(vs))]
		}

		if firstErr == nil {
			firstErr = fmt.Errorf("variable %q not found", name)
		}
		return match // 保留原样
	})
	return result, firstErr
}

// SubstituteBytes bytes 版本
func (r *VariableResolver) SubstituteBytes(input []byte, localVars map[string]string) ([]byte, error) {
	out, err := r.Substitute(string(input), localVars)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// NextValueForGlobal 获取全局变量的下一个值(数据驱动)
func (r *VariableResolver) NextValueForGlobal(name string) (string, error) {
	vs, ok := r.globalVars[name]
	if !ok {
		return "", fmt.Errorf("variable %q not found", name)
	}
	if len(vs) == 0 {
		return "", fmt.Errorf("variable %q has no values", name)
	}
	if len(vs) == 1 {
		return vs[0], nil
	}
	return vs[rand.Intn(len(vs))], nil
}

// AllGlobalVars 返回所有全局变量的副本(调试用)
func (r *VariableResolver) AllGlobalVars() map[string][]string {
	out := make(map[string][]string, len(r.globalVars))
	for k, v := range r.globalVars {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}