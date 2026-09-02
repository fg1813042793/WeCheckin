package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type HTTPHandlerPolicy struct {
	AllowedHosts         []string
	AllowedCIDRs         []string
	AllowPrivateNetworks bool
	MaxRedirects         int
	MaxRequestBytes      int64
	MaxResponseBytes     int64
	Resolver             IPResolver
	Dialer               ContextDialer
}

type IPResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type ContextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

type HTTPCredentialProvider interface {
	ResolveCredential(context.Context, string) (map[string]string, error)
}

type StaticHTTPCredentials map[string]map[string]string

func (credentials StaticHTTPCredentials) ResolveCredential(_ context.Context, reference string) (map[string]string, error) {
	headers, ok := credentials[strings.TrimSpace(reference)]
	if !ok {
		return nil, fmt.Errorf("HTTP credential reference %q is not registered", reference)
	}
	result := make(map[string]string, len(headers))
	for key, value := range headers {
		result[key] = value
	}
	return result, nil
}

type HTTPHandler struct {
	policy      HTTPHandlerPolicy
	allowedNets []*net.IPNet
	credentials HTTPCredentialProvider
}

type httpHandlerConfig struct {
	Method         string            `json:"method"`
	URL            string            `json:"url"`
	Query          map[string]string `json:"query"`
	Headers        map[string]string `json:"headers"`
	Body           json.RawMessage   `json:"body"`
	TimeoutMillis  int               `json:"timeoutMillis"`
	ExpectedStatus []int             `json:"expectedStatus"`
	CredentialRef  string            `json:"credentialRef"`
}

func NewHTTPHandler(policy HTTPHandlerPolicy, credentials HTTPCredentialProvider) (*HTTPHandler, error) {
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
	handler := &HTTPHandler{policy: policy, credentials: credentials}
	for _, raw := range policy.AllowedCIDRs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled task HTTP CIDR %q: %w", raw, err)
		}
		handler.allowedNets = append(handler.allowedNets, network)
	}
	return handler, nil
}

func (handler *HTTPHandler) Type() string { return "http" }

func (handler *HTTPHandler) Metadata() application.HandlerMetadata {
	return application.HandlerMetadata{
		Type: "http", Name: "HTTP/Webhook", Description: "Sends an allowlisted HTTP request",
		RiskLevel: "high", ConfigSchema: json.RawMessage(`{
			"type":"object","required":["method","url"],
			"properties":{
				"method":{"type":"string","enum":["GET","POST","PUT","PATCH","DELETE"]},
				"url":{"type":"string"},"query":{"type":"object"},"headers":{"type":"object"},
				"body":{},"timeoutMillis":{"type":"integer","minimum":1},
				"expectedStatus":{"type":"array","items":{"type":"integer"}},
				"credentialRef":{"type":"string"}
			}
		}`),
	}
}

func (handler *HTTPHandler) ValidateConfig(ctx context.Context, raw json.RawMessage) error {
	_, _, err := handler.decodeAndValidate(ctx, raw)
	return err
}

func (handler *HTTPHandler) Execute(ctx context.Context, run application.RunContext) (application.HandlerResult, error) {
	config, target, err := handler.decodeAndValidate(ctx, json.RawMessage(run.Task.HandlerConfigJSON))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	requestURL := *target
	query := requestURL.Query()
	for key, value := range config.Query {
		query.Set(key, value)
	}
	requestURL.RawQuery = query.Encode()

	requestCtx := ctx
	cancel := func() {}
	if config.TimeoutMillis > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, time.Duration(config.TimeoutMillis)*time.Millisecond)
	}
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, config.Method, requestURL.String(), bytes.NewReader(config.Body))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_request", Summary: err.Error()}
	}
	for key, value := range config.Headers {
		request.Header.Set(key, value)
	}
	if len(config.Body) > 0 && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if config.CredentialRef != "" {
		headers, err := handler.credentials.ResolveCredential(ctx, config.CredentialRef)
		if err != nil {
			return application.HandlerResult{}, &application.HandlerError{Code: "credential_unavailable", Summary: err.Error()}
		}
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}
	request.Header.Set("X-Scheduled-Run-ID", run.RunID)

	client := &http.Client{Transport: handler.transport(), CheckRedirect: handler.checkRedirect}
	response, err := client.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) || isTimeoutError(err) {
			return application.HandlerResult{}, &application.HandlerError{Code: "timeout", Summary: "HTTP request timeout", Temporary: true}
		}
		return application.HandlerResult{}, &application.HandlerError{Code: "request_failed", Summary: err.Error(), Temporary: true}
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, handler.policy.MaxResponseBytes+1))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "response_read_failed", Summary: err.Error(), Temporary: true}
	}
	if int64(len(body)) > handler.policy.MaxResponseBytes {
		return application.HandlerResult{}, &application.HandlerError{Code: "response_too_large", Summary: "HTTP response body exceeds configured limit"}
	}
	if !statusExpected(response.StatusCode, config.ExpectedStatus) {
		return application.HandlerResult{}, &application.HandlerError{
			Code: "unexpected_status", Summary: fmt.Sprintf("HTTP response status %d: %s", response.StatusCode, string(body)),
			Temporary: response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500,
		}
	}
	return application.HandlerResult{
		Summary: fmt.Sprintf("HTTP %d: %s", response.StatusCode, string(body)),
		Data:    map[string]interface{}{"statusCode": response.StatusCode},
	}, nil
}

func (handler *HTTPHandler) decodeAndValidate(ctx context.Context, raw json.RawMessage) (httpHandlerConfig, *url.URL, error) {
	var config httpHandlerConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, nil, fmt.Errorf("decode HTTP task config: %w", err)
	}
	config.Method = strings.ToUpper(strings.TrimSpace(config.Method))
	switch config.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return config, nil, errors.New("HTTP method is not allowed")
	}
	target, err := url.Parse(strings.TrimSpace(config.URL))
	if err != nil || target.Hostname() == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil {
		return config, nil, errors.New("HTTP URL must be an absolute http or https URL without user info")
	}
	if len(config.Body) > 0 && int64(len(config.Body)) > handler.policy.MaxRequestBytes {
		return config, nil, errors.New("HTTP request body exceeds configured limit")
	}
	for key := range config.Headers {
		if isSensitiveTaskHeader(key) || strings.EqualFold(key, "Host") {
			return config, nil, fmt.Errorf("HTTP header %s must be provided by a server credential reference", key)
		}
	}
	if config.CredentialRef != "" && handler.credentials == nil {
		return config, nil, errors.New("HTTP credential provider is not initialized")
	}
	if _, err := handler.validateHost(ctx, target.Hostname()); err != nil {
		return config, nil, err
	}
	return config, target, nil
}

func (handler *HTTPHandler) transport() *http.Transport {
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			addresses, err := handler.validateHost(ctx, host)
			if err != nil {
				return nil, err
			}
			var dialErrors []error
			for _, address := range addresses {
				connection, err := handler.policy.Dialer.DialContext(ctx, network, net.JoinHostPort(address.IP.String(), port))
				if err == nil {
					return connection, nil
				}
				dialErrors = append(dialErrors, err)
			}
			return nil, errors.Join(dialErrors...)
		},
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  false,
	}
}

func (handler *HTTPHandler) checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) > handler.policy.MaxRedirects {
		return errors.New("HTTP redirect limit exceeded")
	}
	_, err := handler.validateHost(request.Context(), request.URL.Hostname())
	return err
}

func (handler *HTTPHandler) validateHost(ctx context.Context, host string) ([]net.IPAddr, error) {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if !hostAllowed(host, handler.policy.AllowedHosts) {
		return nil, fmt.Errorf("HTTP host %q is not allowlisted", host)
	}
	var addresses []net.IPAddr
	if parsed := net.ParseIP(host); parsed != nil {
		addresses = []net.IPAddr{{IP: parsed}}
	} else {
		resolved, err := handler.policy.Resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve HTTP host %q: %w", host, err)
		}
		addresses = resolved
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("HTTP host %q did not resolve", host)
	}
	for _, address := range addresses {
		if err := handler.validateIP(address.IP); err != nil {
			return nil, err
		}
	}
	return addresses, nil
}

func (handler *HTTPHandler) validateIP(ip net.IP) error {
	if ip == nil {
		return errors.New("HTTP target IP is invalid")
	}
	if ip.Equal(net.ParseIP("169.254.169.254")) || ip.Equal(net.ParseIP("fd00:ec2::254")) {
		return errors.New("cloud metadata address is blocked")
	}
	explicitlyAllowed := false
	for _, network := range handler.allowedNets {
		if network.Contains(ip) {
			explicitlyAllowed = true
			break
		}
	}
	if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("HTTP target address %s is blocked", ip)
	}
	if (ip.IsLoopback() || ip.IsPrivate()) && !handler.policy.AllowPrivateNetworks && !explicitlyAllowed {
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

func isSensitiveTaskHeader(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	default:
		return false
	}
}

func statusExpected(status int, expected []int) bool {
	if len(expected) == 0 {
		return status >= 200 && status < 300
	}
	for _, value := range expected {
		if status == value {
			return true
		}
	}
	return false
}

func isTimeoutError(err error) bool {
	var netError net.Error
	return errors.As(err, &netError) && netError.Timeout()
}

var _ application.Handler = (*HTTPHandler)(nil)
