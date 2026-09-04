package outboundhttp

import (
	"context"
	"net"
	"strings"
	"testing"
)

type staticResolver map[string][]net.IPAddr

func (resolver staticResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	return resolver[host], nil
}

func TestClientRejectsPrivateTargetsUnlessPolicyAllowsThem(t *testing.T) {
	resolver := staticResolver{
		"internal.example": {{IP: net.ParseIP("10.0.0.8")}},
		"metadata.example": {{IP: net.ParseIP("169.254.169.254")}},
	}
	client, err := NewClient(Policy{AllowedHosts: []string{"internal.example", "metadata.example"}, Resolver: resolver})
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"http://internal.example/hook", "http://metadata.example/latest/meta-data"} {
		if err := client.Validate(context.Background(), Request{Method: "POST", URL: target}); err == nil {
			t.Fatalf("target %s must be rejected", target)
		}
	}
}

func TestClientAllowsConfiguredCIDRButStillRejectsSensitiveUserHeaders(t *testing.T) {
	client, err := NewClient(Policy{
		AllowedHosts: []string{"hook.example"}, AllowedCIDRs: []string{"10.0.0.0/8"},
		Resolver: staticResolver{"hook.example": {{IP: net.ParseIP("10.0.0.8")}}}, MaxRequestBytes: 16,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Validate(context.Background(), Request{Method: "POST", URL: "http://hook.example/hook", Body: []byte(`{"ok":true}`)}); err != nil {
		t.Fatalf("configured CIDR should be allowed: %v", err)
	}
	for _, request := range []Request{
		{Method: "POST", URL: "http://hook.example/hook", Headers: map[string]string{"Authorization": "secret"}},
		{Method: "POST", URL: "http://hook.example/hook", Body: []byte(strings.Repeat("x", 17))},
	} {
		if err := client.Validate(context.Background(), request); err == nil {
			t.Fatalf("unsafe request must be rejected: %#v", request)
		}
	}
}

func TestClientCanAllowAnyPublicHostWithoutAllowingPrivateAddresses(t *testing.T) {
	client, err := NewClient(Policy{
		AllowAnyPublicHosts: true,
		Resolver: staticResolver{
			"hooks.example.com":    {{IP: net.ParseIP("203.0.113.10")}},
			"internal.example.com": {{IP: net.ParseIP("10.0.0.8")}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Validate(context.Background(), Request{Method: "POST", URL: "https://hooks.example.com/path"}); err != nil {
		t.Fatalf("public host rejected: %v", err)
	}
	if err := client.Validate(context.Background(), Request{Method: "POST", URL: "https://internal.example.com/path"}); err == nil {
		t.Fatal("private host must remain blocked")
	}
}
