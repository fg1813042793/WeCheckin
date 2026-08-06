package builtin

import "regexp"

// compileRegexp 包装 regexp.Compile，处理 nil。
// 拆分出来方便单测覆盖（如果以后要加缓存）。
func compileRegexp(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}
