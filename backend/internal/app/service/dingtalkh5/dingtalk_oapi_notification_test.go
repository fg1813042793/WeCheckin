package dingtalkh5

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestDingTalkOAPIClientSendsWorkNotification(t *testing.T) {
	var sent map[string]interface{}
	client := defaultDingTalkIdentityClient{baseURL: "https://oapi.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/gettoken":
			if r.URL.Query().Get("appkey") != "app-key" || r.URL.Query().Get("appsecret") != "app-secret" {
				t.Fatalf("unexpected token query: %s", r.URL.RawQuery)
			}
			return jsonTestResponse(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "token-1",
			})
		case "/topapi/message/corpconversation/asyncsend_v2":
			if r.URL.Query().Get("access_token") != "token-1" {
				t.Fatalf("unexpected access token: %s", r.URL.RawQuery)
			}
			if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
				t.Fatalf("decode notification body: %v", err)
			}
			return jsonTestResponse(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"task_id": 1,
			})
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"errcode": 404, "errmsg": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:    "corp-a",
		AppKey:    "app-key",
		AppSecret: "app-secret",
		AgentID:   "123456",
	}, []string{"manager001"}, "Lip 已提交 2026-07 月度考评，请处理。")
	if err != nil {
		t.Fatalf("SendWorkNotificationContext error = %v", err)
	}
	if sent["agent_id"] != float64(123456) {
		t.Fatalf("agent_id = %#v, want 123456", sent["agent_id"])
	}
	if sent["userid_list"] != "manager001" {
		t.Fatalf("userid_list = %#v, want manager001", sent["userid_list"])
	}
	msg, _ := sent["msg"].(map[string]interface{})
	text, _ := msg["text"].(map[string]interface{})
	if text["content"] != "Lip 已提交 2026-07 月度考评，请处理。" {
		t.Fatalf("content = %#v", text["content"])
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonTestResponse(payload map[string]interface{}) (*http.Response, error) {
	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(payload)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(&body),
	}, nil
}
