// Package builtin 注册 formkit 内置题型。导入此包即可一次性注册所有内置题型。
package builtin

// 各题型在自身文件内通过 init() 自动注册。导入 builtin 即可触发。

import (
	"wecheckin-backend/backend/internal/formkit/question"
)

// Registered 列出所有已注册题型（仅供测试 / 文档生成时使用）。
func Registered() []question.Question {
	return question.All()
}
