package outboundhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrResponseTooLarge = errors.New("HTTP response body exceeds configured limit")

type IPResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type ContextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

type Policy struct {
	AllowedHosts         []string
	AllowAnyPublicHosts  bool
	AllowedCIDRs         []string
	AllowPrivateNetworks bool
	MaxRedirects         int
	MaxRequestBytes      int64
	MaxResponseBytes     int64
	Resolver             IPResolver
	Dialer               ContextDialer
}

type Request struct {
	Method         string
	URL            string
	Query          map[string]string
	Headers        map[string]string
	TrustedHeaders map[string]string
	Body           []byte
	Timeout        time.Duration
}

type Response struct {
	StatusCode int
	Body       []byte
}

type Client struct {
	policy      Policy
	allowedNets []*net.IPNet
}

func NewClient(policy Policy) (*Client, error) {
	if policy.Resolver == nil {
		policy.Resolver = net.DefaultResolver
	}
	if policy.Dialer == nil {
		policy.Dialer = &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	}
	if policy.MaxRequestBytes <= 0 {
		policy.MaxRequestBytes = 1 << 20
	}
	if policy.MaxResponseBytes <= 0 {
		policy.MaxResponseBytes = 1 << 20
	}
	client := &Client{policy: policy}
	for _, raw := range policy.AllowedCIDRs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid outbound HTTP CIDR %q: %w", raw, err)
		}
		client.allowedNets = append(client.allowedNets, network)
	}
	return client, nil
}

func (client *Client) Validate(ctx context.Context, request Request) error {
	_, err := client.validate(ctx, request)
	return err
}

func (client *Client) Do(ctx context.Context, request Request) (Response, error) {
	target, err := client.validate(ctx, request)
	if err != nil {
		return Response{}, err
	}
	query := target.Query()
	for key, value := range request.Query {
		query.Set(key, value)
	}
	target.RawQuery = query.Encode()
	requestCtx := ctx
	cancel := func() {}
	if request.Timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, request.Timeout)
	}
	defer cancel()
	httpRequest, err := http.NewRequestWithContext(requestCtx, strings.ToUpper(strings.TrimSpace(request.Method)), target.String(), bytes.NewReader(request.Body))
	if err != nil {
		return Response{}, err
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}
	for key, value := range request.TrustedHeaders {
		httpRequest.Header.Set(key, value)
	}
	if len(request.Body) > 0 && httpRequest.Header.Get("Content-Type") == "" {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	response, err := (&http.Client{Transport: client.transport(), CheckRedirect: client.checkRedirect}).Do(httpRequest)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, client.policy.MaxResponseBytes+1))
	if err != nil {
		return Response{}, err
	}
	if int64(len(body)) > client.policy.MaxResponseBytes {
		return Response{}, ErrResponseTooLarge
	}
	return Response{StatusCode: response.StatusCode, Body: body}, nil
}

func (client *Client) validate(ctx context.Context, request Request) (*url.URL, error) {
	if client == nil {
		return nil, errors.New("outbound HTTP client is not initialized")
	}
	method := strings.ToUpper(strings.TrimSpace(request.Method))
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return nil, errors.New("HTTP method is not allowed")
	}
	target, err := url.Parse(strings.TrimSpace(request.URL))
	if err != nil || target.Hostname() == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil {
		return nil, errors.New("HTTP URL must be an absolute http or https URL without user info")
	}
	if int64(len(request.Body)) > client.policy.MaxRequestBytes {
		return nil, errors.New("HTTP request body exceeds configured limit")
	}
	for key := range request.Headers {
		if IsSensitiveHeader(key) || strings.EqualFold(key, "Host") {
			return nil, fmt.Errorf("HTTP header %s must be provided by trusted server configuration", key)
		}
	}
	if _, err := client.validateHost(ctx, target.Hostname()); err != nil {
		return nil, err
	}
	return target, nil
}

func (client *Client) transport() *http.Transport {
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			addresses, err := client.validateHost(ctx, host)
			if err != nil {
				return nil, err
			}
			var dialErrors []error
			for _, target := range addresses {
				connection, err := client.policy.Dialer.DialContext(ctx, network, net.JoinHostPort(target.IP.String(), port))
				if err == nil {
					return connection, nil
				}
				dialErrors = append(dialErrors, err)
			}
			return nil, errors.Join(dialErrors...)
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

func (client *Client) checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) > client.policy.MaxRedirects {
		return errors.New("HTTP redirect limit exceeded")
	}
	_, err := client.validateHost(request.Context(), request.URL.Hostname())
	return err
}

func (client *Client) validateHost(ctx context.Context, host string) ([]net.IPAddr, error) {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if !client.policy.AllowAnyPublicHosts && !hostAllowed(host, client.policy.AllowedHosts) {
		return nil, fmt.Errorf("HTTP host %q is not allowlisted", host)
	}
	var addresses []net.IPAddr
	if parsed := net.ParseIP(host); parsed != nil {
		addresses = []net.IPAddr{{IP: parsed}}
	} else {
		resolved, err := client.policy.Resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve HTTP host %q: %w", host, err)
		}
		addresses = resolved
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("HTTP host %q did not resolve", host)
	}
	for _, address := range addresses {
		if err := client.validateIP(address.IP); err != nil {
			return nil, err
		}
	}
	return addresses, nil
}

func (client *Client) validateIP(ip net.IP) error {
	if ip == nil {
		return errors.New("HTTP target IP is invalid")
	}
	if ip.Equal(net.ParseIP("169.254.169.254")) || ip.Equal(net.ParseIP("fd00:ec2::254")) {
		return errors.New("cloud metadata address is blocked")
	}
	explicitlyAllowed := false
	for _, network := range client.allowedNets {
		if network.Contains(ip) {
			explicitlyAllowed = true
			break
		}
	}
	if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("HTTP target address %s is blocked", ip)
	}
	if (ip.IsLoopback() || ip.IsPrivate()) && !client.policy.AllowPrivateNetworks && !explicitlyAllowed {
		return fmt.Errorf("HTTP private target address %s is blocked", ip)
	}
	return nil
}

func hostAllowed(host string, allowed []string) bool {
	for _, pattern := range allowed {
		pattern = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(pattern)), ".")
		if pattern == host {
			return true
		}
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(host, suffix) && host != strings.TrimPrefix(suffix, ".") {
				return true
			}
		}
	}
	return false
}

func IsSensitiveHeader(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	default:
		return false
	}
}

func IsTimeout(err error) bool {
	var netError net.Error
	return errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netError) && netError.Timeout())
}
