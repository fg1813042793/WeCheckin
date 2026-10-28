package survey

import (
	"testing"

	"wecheckin/backend/internal/model"
)

func TestParseResponseUserAgent(t *testing.T) {
	const iphone = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko)"
	const android = "Mozilla/5.0 (Linux; Android 14; TestDevice) AppleWebKit/537.36 (KHTML, like Gecko)"
	tests := []struct {
		name, ua, browser, device, platform string
	}{
		{"iPhone Safari", iphone + " Version/18.0 Mobile/15E148 Safari/604.1", "Safari", "移动端", "iOS"},
		{"iPhone unnamed webview", iphone + " Mobile/15E148", "iOS WebView", "移动端", "iOS"},
		{"iPhone DingTalk", iphone + " Mobile/15E148 AliApp(DingTalk/7.6.0)", "钉钉", "移动端", "iOS"},
		{"iPhone Chrome", iphone + " CriOS/130.0.0.0 Mobile/15E148 Safari/604.1", "Chrome", "移动端", "iOS"},
		{"iPhone Firefox", iphone + " FxiOS/130.0 Mobile/15E148 Safari/605.1.15", "Firefox", "移动端", "iOS"},
		{"iPhone Edge", iphone + " EdgiOS/130.0 Mobile/15E148 Safari/605.1.15", "Edge", "移动端", "iOS"},
		{"iPad without mobile token", "Mozilla/5.0 (iPad; CPU OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Version/18.0 Safari/604.1", "Safari", "移动端", "iOS"},
		{"iPad mobile Mac UA", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15) AppleWebKit/605.1.15 Mobile/15E148", "iOS WebView", "移动端", "iOS"},
		{"Android Chrome", android + " Chrome/130.0.0.0 Mobile Safari/537.36", "Chrome", "移动端", "Android"},
		{"Android webview", "Mozilla/5.0 (Linux; Android 14; TestDevice; wv) AppleWebKit/537.36 Version/4.0 Chrome/130.0.0.0 Mobile Safari/537.36", "Android WebView", "移动端", "Android"},
		{"Android Edge", android + " Chrome/130.0.0.0 Mobile Safari/537.36 EdgA/130.0", "Edge", "移动端", "Android"},
		{"Android WeChat", android + " Chrome/130.0.0.0 Mobile Safari/537.36 MicroMessenger/8.0", "微信", "移动端", "Android"},
		{"Android DingTalk", android + " Chrome/130.0.0.0 Mobile Safari/537.36 AliApp(DingTalk/7.6.0)", "钉钉", "移动端", "Android"},
		{"macOS Safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 Version/18.0 Safari/605.1.15", "Safari", "PC端", "macOS"},
		{"Windows Edge", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/130.0.0.0 Safari/537.36 Edg/130.0", "Edge", "PC端", "Windows"},
		{"Linux Firefox", "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0", "Firefox", "PC端", "Linux"},
		{"OpenHarmony", "Mozilla/5.0 (Phone; OpenHarmony 6.1) AppleWebKit/537.36 Chrome/132.0 Mobile MicroMessenger/8.0", "微信", "移动端", "OpenHarmony"},
		{"native iOS app", "uni-app/1.0 (ios; iPhone; apple; zh-CN)", "App（uni-app）", "移动端", "iOS"},
		{"native Android app", "uni-app/1.0 (android; TestDevice; zh-CN)", "App（uni-app）", "移动端", "Android"},
		{"missing", " \t ", "-", "-", "-"},
		{"unknown client", "CustomClient/1.0", "-", "-", "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			browser, device, platform := parseUA(tt.ua)
			if browser != tt.browser || device != tt.device || platform != tt.platform {
				t.Fatalf("metadata = (%q, %q, %q), want (%q, %q, %q)", browser, device, platform, tt.browser, tt.device, tt.platform)
			}
		})
	}
}

func TestNormalizeResponseDeviceMetadataRepairsLegacyValues(t *testing.T) {
	tests := []struct {
		name                      string
		response                  model.SurveyResponse
		browser, device, platform string
	}{
		{
			name:     "missing iOS browser and misclassified platform",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148", Browser: "-", DeviceType: "移动端", PlatformType: "macOS"},
			browser:  "iOS WebView", device: "移动端", platform: "iOS",
		},
		{
			name:     "iPad legacy PC classification",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (iPad; CPU OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Version/18.0 Safari/604.1", Browser: "Safari", DeviceType: "PC端", PlatformType: "macOS"},
			browser:  "Safari", device: "移动端", platform: "iOS",
		},
		{
			name:     "iOS Chrome previously Safari",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 CriOS/130.0 Mobile/15E148 Safari/604.1", Browser: "Safari", DeviceType: "移动端", PlatformType: "macOS"},
			browser:  "Chrome", device: "移动端", platform: "iOS",
		},
		{
			name:     "Android previously Linux",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (Linux; Android 14) Chrome/130.0 Mobile Safari/537.36", Browser: "Chrome", DeviceType: "移动端", PlatformType: "Linux"},
			browser:  "Chrome", device: "移动端", platform: "Android",
		},
		{
			name:     "retain known browser when stored UA lacks its suffix",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148", Browser: "微信", DeviceType: "移动端", PlatformType: "iOS"},
			browser:  "微信", device: "移动端", platform: "iOS",
		},
		{
			name:     "retain metadata for unknown UA",
			response: model.SurveyResponse{Device: "CustomClient/1.0", Browser: "企业客户端", DeviceType: "移动端", PlatformType: "iOS"},
			browser:  "企业客户端", device: "移动端", platform: "iOS",
		},
		{
			name:     "retain Safari when only webview markers remain",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148", Browser: "Safari", DeviceType: "移动端", PlatformType: "macOS"},
			browser:  "Safari", device: "移动端", platform: "iOS",
		},
		{
			name:     "retain Edge when its suffix is missing",
			response: model.SurveyResponse{Device: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/130.0.0.0 Safari/537.36", Browser: "Edge", DeviceType: "PC端", PlatformType: "Windows"},
			browser:  "Edge", device: "PC端", platform: "Windows",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalUA := tt.response.Device
			normalizeResponseDeviceMetadata(&tt.response)
			if tt.response.Browser != tt.browser || tt.response.DeviceType != tt.device || tt.response.PlatformType != tt.platform {
				t.Fatalf("metadata = (%q, %q, %q), want (%q, %q, %q)", tt.response.Browser, tt.response.DeviceType, tt.response.PlatformType, tt.browser, tt.device, tt.platform)
			}
			if tt.response.Device != originalUA {
				t.Fatal("normalization must preserve the original UA")
			}
		})
	}
	normalizeResponseDeviceMetadata(nil)
}
