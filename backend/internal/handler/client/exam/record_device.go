package exam

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	examservice "wecheckin/backend/internal/service/client/exam"
)

func clientEnvironment(c *app.RequestContext, reportedDevice string) examservice.ClientEnvironment {
	device := strings.TrimSpace(reportedDevice)
	if device == "" {
		device = strings.TrimSpace(string(c.UserAgent()))
	}
	// exam_r_device is VARCHAR(512); truncate by characters so UTF-8 stays valid.
	runes := []rune(device)
	if len(runes) > 512 {
		device = string(runes[:512])
	}
	return examservice.ClientEnvironment{Device: device, IP: c.ClientIP()}
}
