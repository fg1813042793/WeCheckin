package config

import (
	"strings"
	"testing"
)

func TestAuditProxyConfigDefaultsAndEnvironment(t *testing.T) {
	withTempConfigDir(t, testConfigYAML)
	t.Setenv("WECHECKIN_SERVER_AUDIT_TRUSTED_PROXY_CIDRS", "")
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Server.AuditTrustedProxyCIDRs) != 0 {
		t.Fatal("audit must not trust forwarded headers by default")
	}
	t.Setenv("WECHECKIN_SERVER_AUDIT_TRUSTED_PROXY_CIDRS", "127.0.0.1/32, ::1/128")
	cfg, err = LoadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Server.AuditTrustedProxyCIDRs) != 2 || cfg.Server.AuditTrustedProxyCIDRs[1] != "::1/128" {
		t.Fatalf("proxies=%v", cfg.Server.AuditTrustedProxyCIDRs)
	}
}

func TestAuditProxyConfigRejectsInvalidOrUnrestrictedTrust(t *testing.T) {
	for _, cidr := range []string{"bad", "0.0.0.0/0", "::/0"} {
		cfg := Config{Server: ServerConfig{AuditTrustedProxyCIDRs: []string{cidr}}}
		if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "audit_trusted_proxy_cidrs") {
			t.Fatalf("%s: %v", cidr, err)
		}
	}
}
