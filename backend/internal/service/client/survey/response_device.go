package survey

import "strings"

// parseUA 解析 User-Agent 获取浏览器、设备类型、平台类型
func parseUA(ua string) (browser, deviceType, platformType string) {
	u := strings.ToLower(ua)
	switch {
	case strings.Contains(u, "micromessenger"):
		browser = "微信"
	case strings.Contains(u, "edg"):
		browser = "Edge"
	case strings.Contains(u, "chrome"):
		browser = "Chrome"
	case strings.Contains(u, "firefox"):
		browser = "Firefox"
	case strings.Contains(u, "safari"):
		browser = "Safari"
	default:
		browser = "-"
	}
	if strings.Contains(u, "mobile") || strings.Contains(u, "android") || strings.Contains(u, "iphone") {
		deviceType = "移动端"
	} else {
		deviceType = "PC端"
	}
	switch {
	case strings.Contains(u, "win"):
		platformType = "Windows"
	case strings.Contains(u, "mac"):
		platformType = "macOS"
	case strings.Contains(u, "linux"):
		platformType = "Linux"
	case strings.Contains(u, "android"):
		platformType = "Android"
	case strings.Contains(u, "iphone") || strings.Contains(u, "ipad"):
		platformType = "iOS"
	default:
		platformType = "-"
	}
	return
}
