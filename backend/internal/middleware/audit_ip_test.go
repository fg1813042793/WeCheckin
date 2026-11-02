package middleware

import (
	"net"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/network"
	"wecheckin/backend/internal/config"
)

type auditTestConn struct {
	network.Conn
	peer net.Addr
}

func (c auditTestConn) RemoteAddr() net.Addr { return c.peer }

func TestAuditIPTrustsOnlyConfiguredProxyPeers(t *testing.T) {
	old := config.Cfg
	t.Cleanup(func() { config.Cfg = old })
	for _, tt := range []struct {
		trusted                   []string
		peer, forwarded, expected string
	}{
		{nil, "198.51.100.12", "192.0.2.5", "198.51.100.12"},
		{[]string{"127.0.0.1/32"}, "127.0.0.1", "192.0.2.5, 198.51.100.12", "198.51.100.12"},
		{[]string{"127.0.0.1/32"}, "127.0.0.1", "192.0.2.5", "192.0.2.5"},
		{[]string{"127.0.0.1/32"}, "198.51.100.12", "192.0.2.5", "198.51.100.12"},
	} {
		config.Cfg = &config.Config{Server: config.ServerConfig{AuditTrustedProxyCIDRs: tt.trusted}}
		c := app.NewContext(0)
		c.SetConn(auditTestConn{peer: &net.TCPAddr{IP: net.ParseIP(tt.peer), Port: 1234}})
		c.Request.Header.Set("X-Forwarded-For", tt.forwarded)
		if got := ClientIPResolver()(c); got != tt.expected {
			t.Fatalf("peer=%s trust=%v got=%s want=%s", tt.peer, tt.trusted, got, tt.expected)
		}
	}
}
