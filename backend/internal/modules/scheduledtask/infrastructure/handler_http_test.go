package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestHTTPHandlerSendsAllowedRequestWithCredentialReferenceAndRunID(t *testing.T) {
	var authorization, runID string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		runID = request.Header.Get("X-Scheduled-Run-ID")
		response.WriteHeader(http.StatusCreated)
		_, _ = response.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	handler := newLocalHTTPHandler(t, StaticHTTPCredentials{"service-a": {"Authorization": "Bearer secret"}})
	result, err := handler.Execute(context.Background(), application.RunContext{
		RunID: "run-42", Task: application.TaskSnapshot{HandlerConfigJSON: mustJSON(t, map[string]interface{}{
			"method": "POST", "url": server.URL, "body": map[string]interface{}{"hello": "world"},
			"credentialRef": "service-a", "expectedStatus": []int{201},
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer secret" || runID != "run-42" || !strings.Contains(result.Summary, `"ok":true`) {
		t.Fatalf("headers/result = %q / %q / %#v", authorization, runID, result)
	}
}

func TestHTTPHandlerRejectsLoopbackAndCloudMetadataByDefault(t *testing.T) {
	handler, err := NewHTTPHandler(HTTPHandlerPolicy{AllowedHosts: []string{"127.0.0.1", "169.254.169.254"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"http://127.0.0.1/internal", "http://169.254.169.254/latest/meta-data"} {
		err := handler.ValidateConfig(context.Background(), json.RawMessage(mustJSON(t, map[string]interface{}{"method": "GET", "url": target})))
		if err == nil {
			t.Fatalf("target %s must be rejected", target)
		}
	}
}

func TestHTTPHandlerRevalidatesRedirectTargets(t *testing.T) {
	redirect := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "http://169.254.169.254/latest/meta-data", http.StatusFound)
	}))
	defer redirect.Close()
	handler, err := NewHTTPHandler(HTTPHandlerPolicy{
		AllowedHosts: []string{"127.0.0.1", "169.254.169.254"}, AllowedCIDRs: []string{"127.0.0.0/8"}, MaxRedirects: 1,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.Execute(context.Background(), application.RunContext{
		RunID: "run-redirect", Task: application.TaskSnapshot{HandlerConfigJSON: mustJSON(t, map[string]interface{}{"method": "GET", "url": redirect.URL})},
	})
	if err == nil || !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("redirect error = %v", err)
	}
}

func TestHTTPHandlerLimitsRequestAndResponseBodies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(strings.Repeat("x", 64)))
	}))
	defer server.Close()
	handler, err := NewHTTPHandler(HTTPHandlerPolicy{
		AllowedHosts: []string{"127.0.0.1"}, AllowedCIDRs: []string{"127.0.0.0/8"},
		MaxRequestBytes: 16, MaxResponseBytes: 16,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	tooLargeRequest := json.RawMessage(mustJSON(t, map[string]interface{}{
		"method": "POST", "url": server.URL, "body": strings.Repeat("x", 32),
	}))
	if err := handler.ValidateConfig(context.Background(), tooLargeRequest); err == nil {
		t.Fatal("oversized request body must fail validation")
	}
	_, err = handler.Execute(context.Background(), application.RunContext{
		RunID: "run-response", Task: application.TaskSnapshot{HandlerConfigJSON: mustJSON(t, map[string]interface{}{"method": "GET", "url": server.URL})},
	})
	if err == nil || !strings.Contains(err.Error(), "response body") {
		t.Fatalf("response limit error = %v", err)
	}
}

func TestHTTPHandlerHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	handler := newLocalHTTPHandler(t, nil)
	_, err := handler.Execute(context.Background(), application.RunContext{
		RunID: "run-timeout", Task: application.TaskSnapshot{HandlerConfigJSON: mustJSON(t, map[string]interface{}{
			"method": "GET", "url": server.URL, "timeoutMillis": 30,
		})},
	})
	var handlerError *application.HandlerError
	if err == nil || !strings.Contains(err.Error(), "timeout") || !errorsAs(err, &handlerError) || !handlerError.Temporary {
		t.Fatalf("timeout error = %#v", err)
	}
}

func newLocalHTTPHandler(t *testing.T, credentials StaticHTTPCredentials) *HTTPHandler {
	t.Helper()
	handler, err := NewHTTPHandler(HTTPHandlerPolicy{
		AllowedHosts: []string{"127.0.0.1"}, AllowedCIDRs: []string{"127.0.0.0/8"},
		MaxRequestBytes: 1024, MaxResponseBytes: 1024,
	}, credentials)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func mustJSON(t *testing.T, value interface{}) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func errorsAs(err error, target interface{}) bool {
	return errors.As(err, target)
}
