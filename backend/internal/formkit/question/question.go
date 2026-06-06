package question

import (
	"sync"

	"wecheckin-backend/backend/internal/formkit/schema"
)

// ValidationError 单个字段校验失败详情
type ValidationError struct {
	QuestionID string
	Field      string
	Message    string
}

func (e *ValidationError) Error() string {
	if e.QuestionID != "" {
		return e.QuestionID + ": " + e.Message
	}
	return e.Message
}

// Question 题型接口。
//
// 题型是「题的类型」的语义单元。一个题在 schema 里是一段配置（id/type/title/...），
// 题型负责：
//   - 校验用户填的答案是否合法；
//   - 从表单原始提交数据中提取该题答案；
//   - 把答案格式化为展示串（列表/报表里用）；
//   - 提供默认 props，供前端编辑器预填；
//   - 标识自身（如 Type），便于注册表检索。
type Question interface {
	// Type 题型唯一标识，与 schema.Question.Type 一致。
	Type() string

	// DisplayName 题型在前端编辑器里的中文显示名。
	DisplayName() string

	// Category 题型分类，用于编辑器分组。常见值：base/select/media/advanced。
	Category() string

	// Validate 校验答案 value 是否符合 q 的约束。
	// 默认会先跑 schema 里声明的 Validate 规则；题型只做类型特定的检查。
	Validate(value interface{}, q schema.Question) error

	// ExtractValue 从表单提交 data 中按 q.ID 取出该题答案。
	// 若不存在返回 nil。
	ExtractValue(data map[string]interface{}, q schema.Question) interface{}

	// FormatValue 把值格式化为展示串。nil/空返回 ""。
	FormatValue(value interface{}, q schema.Question) string

	// DefaultProps 返回该题型的默认 props (placeholder/options/maxLength 等)。
	DefaultProps() map[string]interface{}
}

// Registry 题型注册表（单例）。
type Registry struct {
	mu sync.RWMutex
	m  map[string]Question
}

var defaultRegistry = &Registry{m: map[string]Question{}}

// Register 注册一个题型。同 Type 重复注册会覆盖。
func Register(q Question) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.m[q.Type()] = q
}

// Get 取题型。找不到返回 nil。
func Get(t string) Question {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	return defaultRegistry.m[t]
}

// Has 题型是否已注册。
func Has(t string) bool {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	_, ok := defaultRegistry.m[t]
	return ok
}

// Types 返回所有已注册题型 Type，按注册顺序。
func Types() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	out := make([]string, 0, len(defaultRegistry.m))
	for k := range defaultRegistry.m {
		out = append(out, k)
	}
	return out
}

// All 返回所有已注册题型实例。
func All() []Question {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	out := make([]Question, 0, len(defaultRegistry.m))
	for _, q := range defaultRegistry.m {
		out = append(out, q)
	}
	return out
}
