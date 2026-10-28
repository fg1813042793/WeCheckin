package survey

import (
	"strings"

	"wecheckin/backend/internal/model"
)

// parseUA 解析 User-Agent 获取浏览器、设备类型、平台类型
func parseUA(ua string) (browser, deviceType, platformType string) {
	u := strings.ToLower(strings.TrimSpace(ua))
	isNativeApp := strings.HasPrefix(u, "uni-app")
	isIOS := strings.Contains(u, "iphone") || strings.Contains(u, "ipad") || strings.Contains(u, "ipod") ||
		(strings.Contains(u, "macintosh") && strings.Contains(u, "mobile/")) ||
		(isNativeApp && strings.Contains(u, "ios"))
	isAndroid := strings.Contains(u, "android")

	// Mobile OS identifiers precede their desktop compatibility tokens.
	switch {
	case strings.Contains(u, "openharmony") || strings.Contains(u, "harmonyos") || strings.Contains(u, "hmos"):
		platformType = "OpenHarmony"
	case isIOS:
		platformType = "iOS"
	case isAndroid:
		platformType = "Android"
	case strings.Contains(u, "windows") || strings.Contains(u, "win32") || strings.Contains(u, "win64"):
		platformType = "Windows"
	case strings.Contains(u, "macintosh") || strings.Contains(u, "mac os") || strings.Contains(u, "macos"):
		platformType = "macOS"
	case strings.Contains(u, "linux"):
		platformType = "Linux"
	default:
		platformType = "-"
	}

	switch {
	case strings.Contains(u, "micromessenger"):
		browser = "微信"
	case strings.Contains(u, "dingtalk") || strings.Contains(u, "aliapp(ding/"):
		browser = "钉钉"
	case isNativeApp:
		browser = "App（uni-app）"
	case strings.Contains(u, "edg/") || strings.Contains(u, "edge/") || strings.Contains(u, "edga/") || strings.Contains(u, "edgios/"):
		browser = "Edge"
	case isAndroid && strings.Contains(u, "; wv"):
		browser = "Android WebView"
	case strings.Contains(u, "chrome/") || strings.Contains(u, "crios/") || strings.Contains(u, "chromium/"):
		browser = "Chrome"
	case strings.Contains(u, "firefox/") || strings.Contains(u, "fxios/"):
		browser = "Firefox"
	case strings.Contains(u, "safari/"):
		browser = "Safari"
	case isIOS && strings.Contains(u, "applewebkit/"):
		browser = "iOS WebView"
	case strings.Contains(u, "applewebkit/"):
		browser = "WebKit"
	default:
		browser = "-"
	}

	switch {
	case isIOS || isAndroid || platformType == "OpenHarmony" || strings.Contains(u, "mobile") || strings.Contains(u, "phone") || strings.Contains(u, "tablet"):
		deviceType = "移动端"
	case platformType == "Windows" || platformType == "macOS" || platformType == "Linux":
		deviceType = "PC端"
	default:
		deviceType = "-"
	}
	return
}

func normalizeResponseDeviceMetadata(response *model.SurveyResponse) {
	if response == nil || strings.TrimSpace(response.Device) == "" {
		return
	}
	browser, deviceType, platformType := parseUA(response.Device)
	switch strings.TrimSpace(response.Browser) {
	case "", "-":
		response.Browser = browser
	case "Safari", "Chrome":
		// Refine compatibility tokens, but preserve specific names if a stored UA is truncated.
		switch browser {
		case "微信", "钉钉", "Edge", "Firefox", "Android WebView":
			response.Browser = browser
		case "Chrome":
			if response.Browser == "Safari" {
				response.Browser = browser
			}
		}
	}
	if response.DeviceType == "" || response.DeviceType == "-" || (response.DeviceType == "PC端" && deviceType == "移动端") {
		response.DeviceType = deviceType
	}
	if response.PlatformType == "" || response.PlatformType == "-" ||
		(response.PlatformType == "macOS" && platformType == "iOS") ||
		(response.PlatformType == "Linux" && platformType == "Android") {
		response.PlatformType = platformType
	}
}
