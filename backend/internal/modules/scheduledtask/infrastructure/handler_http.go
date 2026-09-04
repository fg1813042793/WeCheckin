package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wecheckin/backend/internal/modules/scheduledtask/application"
	"wecheckin/backend/internal/support/outboundhttp"
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

type IPResolver = outboundhttp.IPResolver
type ContextDialer = outboundhttp.ContextDialer

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
	client      *outboundhttp.Client
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
	client, err := outboundhttp.NewClient(outboundhttp.Policy{
		AllowedHosts:         policy.AllowedHosts,
		AllowedCIDRs:         policy.AllowedCIDRs,
		AllowPrivateNetworks: policy.AllowPrivateNetworks,
		MaxRedirects:         policy.MaxRedirects,
		MaxRequestBytes:      policy.MaxRequestBytes,
		MaxResponseBytes:     policy.MaxResponseBytes,
		Resolver:             policy.Resolver,
		Dialer:               policy.Dialer,
	})
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{client: client, credentials: credentials}, nil
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
	_, err := handler.decodeAndValidate(ctx, raw)
	return err
}

func (handler *HTTPHandler) Execute(ctx context.Context, run application.RunContext) (application.HandlerResult, error) {
	config, err := handler.decodeAndValidate(ctx, json.RawMessage(run.Task.HandlerConfigJSON))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	trustedHeaders := make(map[string]string)
	if config.CredentialRef != "" {
		headers, err := handler.credentials.ResolveCredential(ctx, config.CredentialRef)
		if err != nil {
			return application.HandlerResult{}, &application.HandlerError{Code: "credential_unavailable", Summary: err.Error()}
		}
		for key, value := range headers {
			trustedHeaders[key] = value
		}
	}
	trustedHeaders["X-Scheduled-Run-ID"] = run.RunID
	response, err := handler.client.Do(ctx, outboundhttp.Request{
		Method: config.Method, URL: config.URL, Query: config.Query, Headers: config.Headers,
		TrustedHeaders: trustedHeaders, Body: config.Body,
		Timeout: time.Duration(config.TimeoutMillis) * time.Millisecond,
	})
	if err != nil {
		if outboundhttp.IsTimeout(err) {
			return application.HandlerResult{}, &application.HandlerError{Code: "timeout", Summary: "HTTP request timeout", Temporary: true}
		}
		if errors.Is(err, outboundhttp.ErrResponseTooLarge) {
			return application.HandlerResult{}, &application.HandlerError{Code: "response_too_large", Summary: err.Error()}
		}
		return application.HandlerResult{}, &application.HandlerError{Code: "request_failed", Summary: err.Error(), Temporary: true}
	}
	if !statusExpected(response.StatusCode, config.ExpectedStatus) {
		return application.HandlerResult{}, &application.HandlerError{
			Code: "unexpected_status", Summary: fmt.Sprintf("HTTP response status %d: %s", response.StatusCode, string(response.Body)),
			Temporary: response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500,
		}
	}
	return application.HandlerResult{
		Summary: fmt.Sprintf("HTTP %d: %s", response.StatusCode, string(response.Body)),
		Data:    map[string]interface{}{"statusCode": response.StatusCode},
	}, nil
}

func (handler *HTTPHandler) decodeAndValidate(ctx context.Context, raw json.RawMessage) (httpHandlerConfig, error) {
	var config httpHandlerConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, fmt.Errorf("decode HTTP task config: %w", err)
	}
	config.Method = strings.ToUpper(strings.TrimSpace(config.Method))
	if config.CredentialRef != "" && handler.credentials == nil {
		return config, errors.New("HTTP credential provider is not initialized")
	}
	if err := handler.client.Validate(ctx, outboundhttp.Request{
		Method: config.Method, URL: config.URL, Query: config.Query, Headers: config.Headers, Body: config.Body,
	}); err != nil {
		if strings.Contains(err.Error(), "trusted server configuration") {
			return config, errors.New(strings.Replace(err.Error(), "trusted server configuration", "a server credential reference", 1))
		}
		return config, err
	}
	return config, nil
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

var _ application.Handler = (*HTTPHandler)(nil)
