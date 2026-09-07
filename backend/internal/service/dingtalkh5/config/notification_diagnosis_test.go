package config

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestDiagnoseOldAgentWorkNotificationReportsNumericAgentID(t *testing.T) {
	var sent map[string]interface{}
	client := defaultDingTalkIdentityClient{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/topapi/message/corpconversation/asyncsend_v2" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			body, _ := json.Marshal(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"task_id": 7,
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		})},
	}
	step := client.diagnoseOldAgentWorkNotificationContext(context.Background(), "token-1", "342080997", "manager001")
	if step.Status != "success" {
		t.Fatalf("diagnosis status = %q, error = %q", step.Status, step.Error)
	}
	if got := step.Request["agent_id"]; got != int64(342080997) {
		t.Fatalf("diagnosis request agent_id = %#v, want int64 342080997", got)
	}
	if got := sent["agent_id"]; got != float64(342080997) {
		t.Fatalf("sent request agent_id = %#v, want numeric 342080997", got)
	}
}

func TestDiagnoseWorkNotificationUsesRobotModeEndpoints(t *testing.T) {
	paths := make([]string, 0, 2)
	client := defaultDingTalkIdentityClient{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			paths = append(paths, r.URL.Path)
			var body []byte
			switch r.URL.Path {
			case "/v1.0/oauth2/accessToken":
				body, _ = json.Marshal(map[string]interface{}{"accessToken": "token-1", "expireIn": 7200})
			case "/v1.0/robot/oToMessages/batchSend":
				body = []byte(`{}`)
			default:
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	steps, success := client.diagnoseWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		AppKey: "app-key", AppSecret: "app-secret", NotifyMode: "robot", RobotCode: "robot-code",
	}, "manager001")
	if !success {
		t.Fatalf("diagnosis failed: %#v", steps)
	}
	if len(paths) != 2 || paths[0] != "/v1.0/oauth2/accessToken" || paths[1] != "/v1.0/robot/oToMessages/batchSend" {
		t.Fatalf("diagnosis paths = %#v", paths)
	}
}
