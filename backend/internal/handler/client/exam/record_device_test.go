package exam

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestClientEnvironmentUsesReportedDeviceOrUserAgent(t *testing.T) {
	for _, tt := range []struct{ name, reported, header, want string }{
		{"reported native client", " uni-app/1.0 (ios; iPhone) ", "generic HTTP client", "uni-app/1.0 (ios; iPhone)"},
		{"missing device", "", "Mozilla/5.0 Chrome/130.0", "Mozilla/5.0 Chrome/130.0"},
		{"blank device", " \t ", "Mozilla/5.0 Safari/605.1", "Mozilla/5.0 Safari/605.1"},
		{"no metadata", "", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := app.NewContext(0)
			c.Request.Header.Set("User-Agent", tt.header)
			c.Request.SetRequestURI("/api/v2/exams/5/submissions?ip=203.0.113.99")
			c.SetClientIPFunc(func(*app.RequestContext) string { return "192.0.2.10" })
			got := clientEnvironment(c, tt.reported)
			if got.Device != tt.want || got.IP != "192.0.2.10" {
				t.Fatalf("client environment = %+v, want device %q and server-resolved IP", got, tt.want)
			}
		})
	}
}

func TestClientEnvironmentBoundsDeviceWithoutBreakingUTF8(t *testing.T) {
	c := app.NewContext(0)
	for _, reported := range []bool{true, false} {
		ua := strings.Repeat("测", 520)
		body := ""
		if reported {
			body = ua
		} else {
			c.Request.Header.Set("User-Agent", ua)
		}
		got := clientEnvironment(c, body).Device
		if !utf8.ValidString(got) || utf8.RuneCountInString(got) != 512 {
			t.Fatalf("device must fit the existing column and remain UTF-8: bytes=%d", len(got))
		}
	}
}
