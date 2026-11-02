package middleware

import (
	"net"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/config"
)

// ClientIPResolver only accepts forwarded IP headers from explicitly trusted proxy peers.
func ClientIPResolver() app.ClientIP {
	var cidrs []*net.IPNet
	if config.Cfg != nil {
		for _, raw := range config.Cfg.Server.AuditTrustedProxyCIDRs {
			_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
			if err == nil {
				cidrs = append(cidrs, network)
			}
		}
	}
	return app.ClientIPWithOption(app.ClientIPOptions{TrustedCIDRs: cidrs, RemoteIPHeaders: []string{"X-Forwarded-For", "X-Real-IP"}})
}
