package survey

import (
	"strings"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/useragent"
)

func parseUA(ua string) (string, string, string) {
	return useragent.Parse(ua)
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
