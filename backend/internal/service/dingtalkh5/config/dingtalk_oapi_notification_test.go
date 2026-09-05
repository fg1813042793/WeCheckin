package config

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{
		Content: "Lip 已提交 2026-07 月度考评，请处理。",
	})
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

func TestDingTalkOAPIClientSendsAgentOANotification(t *testing.T) {
	var sent map[string]interface{}
	client := defaultDingTalkIdentityClient{baseURL: "https://oapi.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/gettoken":
			return jsonTestResponse(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "token-1",
			})
		case "/topapi/message/corpconversation/asyncsend_v2":
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
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{
		Title:      "绩效流程待办",
		Content:    "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。",
		URL:        "dingtalk://dingtalkclient/action/openapp?corpid=ding-corp&app_id=dingmi-okr-app",
		SourceName: "钉米-OKR",
		PicURL:     "@media-logo",
	})
	if err != nil {
		t.Fatalf("SendWorkNotificationContext error = %v", err)
	}
	msg, _ := sent["msg"].(map[string]interface{})
	if msg["msgtype"] != "oa" {
		t.Fatalf("msgtype = %#v, want oa", msg["msgtype"])
	}
	if _, ok := msg["link"]; ok {
		t.Fatalf("agent work notification must not use link payload because dingtalk:// messageUrl is displayed as dingtalkclient")
	}
	oa, _ := msg["oa"].(map[string]interface{})
	if oa["message_url"] != "dingtalk://dingtalkclient/action/openapp?corpid=ding-corp&app_id=dingmi-okr-app" {
		t.Fatalf("oa message_url = %#v", oa["message_url"])
	}
	if oa["pc_message_url"] != "dingtalk://dingtalkclient/action/openapp?corpid=ding-corp&app_id=dingmi-okr-app" {
		t.Fatalf("oa pc_message_url = %#v", oa["pc_message_url"])
	}
	head, _ := oa["head"].(map[string]interface{})
	if head["text"] != "钉米-OKR" {
		t.Fatalf("oa head text = %#v", head["text"])
	}
	body, _ := oa["body"].(map[string]interface{})
	if body["title"] != "绩效流程待办" {
		t.Fatalf("oa body title = %#v", body["title"])
	}
	if body["content"] != "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。" {
		t.Fatalf("oa body content = %#v", body["content"])
	}
	if body["image"] != "@media-logo" {
		t.Fatalf("oa body image = %#v", body["image"])
	}
}

func TestDingTalkAgentNotificationMessageUsesActionCardWhenRequested(t *testing.T) {
	message := dingTalkAgentNotificationMessage(DingTalkWorkNotificationPayload{
		MessageType: DingTalkMessageTypeActionCard,
		Title:       "《绩效考评单》有新评论",
		Content:     "David 在“上级评价”评论：请关注本次评分",
		URL:         "https://oa.example.com/h5?view=workflow%3Ainstance%3Ainstance-1",
	})
	if message["msgtype"] != DingTalkMessageTypeActionCard {
		t.Fatalf("msgtype = %#v, want action_card", message["msgtype"])
	}
	card, _ := message["action_card"].(map[string]string)
	if card["title"] != "《绩效考评单》有新评论" || card["markdown"] == "" {
		t.Fatalf("action_card = %#v", card)
	}
	if card["single_title"] != "查看流程" || card["single_url"] == "" {
		t.Fatalf("action_card button = %#v", card)
	}
}

func TestDingTalkAgentNotificationMessageSupportsAllWorkNotificationTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload DingTalkWorkNotificationPayload
		msgType string
		bodyKey string
	}{
		{name: "text", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeText, Content: "正文"}, msgType: "text", bodyKey: "text"},
		{name: "image", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeImage, MediaID: "@image"}, msgType: "image", bodyKey: "image"},
		{name: "voice", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeVoice, MediaID: "@voice", Duration: 12}, msgType: "voice", bodyKey: "voice"},
		{name: "file", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeFile, MediaID: "@file"}, msgType: "file", bodyKey: "file"},
		{name: "link", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeLink, Title: "标题", Content: "正文", URL: "https://example.test", PicURL: "https://example.test/p.png"}, msgType: "link", bodyKey: "link"},
		{name: "oa", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeOA, Title: "标题", Content: "正文", URL: "https://example.test"}, msgType: "oa", bodyKey: "oa"},
		{name: "markdown", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeMarkdown, Title: "标题", Content: "## 正文"}, msgType: "markdown", bodyKey: "markdown"},
		{name: "action card", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeActionCard, Title: "标题", Content: "## 正文", URL: "https://example.test"}, msgType: "action_card", bodyKey: "action_card"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := dingTalkAgentNotificationMessage(test.payload)
			if message["msgtype"] != test.msgType {
				t.Fatalf("msgtype = %#v, want %q", message["msgtype"], test.msgType)
			}
			if _, ok := message[test.bodyKey]; !ok {
				t.Fatalf("message body %q missing: %#v", test.bodyKey, message)
			}
		})
	}
}

func TestValidateDingTalkWorkNotificationPayloadChecksTypeSpecificFields(t *testing.T) {
	tests := []struct {
		name    string
		payload DingTalkWorkNotificationPayload
	}{
		{name: "image media", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeImage}},
		{name: "voice duration", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeVoice, MediaID: "@voice"}},
		{name: "file media", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeFile}},
		{name: "link url", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeLink, Title: "标题", Content: "正文"}},
		{name: "oa url", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeOA, Title: "标题", Content: "正文"}},
		{name: "card url", payload: DingTalkWorkNotificationPayload{MessageType: DingTalkMessageTypeActionCard, Title: "标题", Content: "正文"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateDingTalkWorkNotificationPayload(test.payload); err == nil {
				t.Fatalf("validateDingTalkWorkNotificationPayload() error = nil")
			}
		})
	}
}

func TestDingTalkWorkNotificationRejectsActionCardWithoutURL(t *testing.T) {
	client := defaultDingTalkIdentityClient{}
	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		AppKey: "app-key", AppSecret: "app-secret", AgentID: "123",
	}, []string{"ding-user-1"}, DingTalkWorkNotificationPayload{
		MessageType: DingTalkMessageTypeActionCard,
		Title:       "《绩效考评单》有新评论",
		Content:     "David 添加评论：请关注本次评分",
	})
	if err == nil || !strings.Contains(err.Error(), "ActionCard 跳转地址不能为空") {
		t.Fatalf("SendWorkNotificationContext() error = %v", err)
	}
}

func TestDingTalkOAPIClientDoesNotFallbackToRobotWhenAgentModeIsStrict(t *testing.T) {
	agentAttempts := 0
	robotAttempts := 0
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/gettoken":
			return jsonTestResponse(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "old-token",
			})
		case "/topapi/message/corpconversation/asyncsend_v2":
			agentAttempts++
			return jsonTestResponse(map[string]interface{}{
				"errcode": 40035,
				"errmsg":  "agentId【342080997】不合法",
			})
		case "/v1.0/oauth2/accessToken", "/v1.0/robot/oToMessages/batchSend":
			robotAttempts++
			t.Fatalf("strict agent mode must not call robot path: %s", r.URL.Path)
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"errcode": 404, "errmsg": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:     "corp-a",
		AppKey:     "app-key",
		AppSecret:  "app-secret",
		AgentID:    "342080997",
		NotifyMode: "agent",
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{
		Title:   "绩效流程待办",
		Content: "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。",
		URL:     "https://okr.example.com/dingtalk-h5/?view=performance%3Amanager",
	})
	if err == nil {
		t.Fatalf("SendWorkNotificationContext error = nil, want agent invalid error")
	}
	if !strings.Contains(err.Error(), "agentId【342080997】不合法") {
		t.Fatalf("error = %v, want agent invalid error", err)
	}
	if agentAttempts != 1 {
		t.Fatalf("agent attempts = %d, want 1", agentAttempts)
	}
	if robotAttempts != 0 {
		t.Fatalf("robot attempts = %d, want 0", robotAttempts)
	}
}

func TestDingTalkOAPIClientFallsBackToRobotWhenAgentIDInvalidForUnifiedApp(t *testing.T) {
	var robotSent map[string]interface{}
	agentAttempts := 0
	robotAttempts := 0
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/gettoken":
			return jsonTestResponse(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "old-token",
			})
		case "/topapi/message/corpconversation/asyncsend_v2":
			agentAttempts++
			return jsonTestResponse(map[string]interface{}{
				"errcode": 40035,
				"errmsg":  "agentId【342080997】不合法",
			})
		case "/v1.0/oauth2/accessToken":
			return jsonTestResponse(map[string]interface{}{
				"accessToken": "new-token",
				"expireIn":    7200,
			})
		case "/v1.0/robot/oToMessages/batchSend":
			robotAttempts++
			if r.Header.Get("x-acs-dingtalk-access-token") != "new-token" {
				t.Fatalf("unexpected robot access token: %s", r.Header.Get("x-acs-dingtalk-access-token"))
			}
			if err := json.NewDecoder(r.Body).Decode(&robotSent); err != nil {
				t.Fatalf("decode robot body: %v", err)
			}
			return jsonTestResponse(map[string]interface{}{
				"processQueryKey": "query-1",
			})
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"errcode": 404, "errmsg": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:       "corp-a",
		AppKey:       "app-key",
		AppSecret:    "app-secret",
		AgentID:      "342080997",
		UnifiedAppID: "dingmi-okr-app",
		NotifyMode:   "agent_fallback",
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{
		Title:      "绩效流程待办",
		Content:    "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。",
		URL:        "dingtalk://dingtalkclient/action/openapp?corpid=ding-corp&app_id=dingmi-okr-app",
		SourceName: "钉米-OKR",
	})
	if err != nil {
		t.Fatalf("SendWorkNotificationContext error = %v", err)
	}
	if agentAttempts != 1 {
		t.Fatalf("agent attempts = %d, want 1", agentAttempts)
	}
	if robotAttempts != 1 {
		t.Fatalf("robot attempts = %d, want 1", robotAttempts)
	}
	if robotSent["robotCode"] != "app-key" {
		t.Fatalf("robotCode = %#v, want app-key fallback", robotSent["robotCode"])
	}
	if robotSent["msgKey"] != "sampleLink" {
		t.Fatalf("msgKey = %#v, want sampleLink", robotSent["msgKey"])
	}
}

func TestDingTalkOAPIClientFallsBackToRobotWhenAgentIDInvalidWithoutUnifiedApp(t *testing.T) {
	agentAttempts := 0
	robotAttempts := 0
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/gettoken":
			return jsonTestResponse(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "old-token",
			})
		case "/topapi/message/corpconversation/asyncsend_v2":
			agentAttempts++
			return jsonTestResponse(map[string]interface{}{
				"errcode": 40035,
				"errmsg":  "agentId【342080997】不合法",
			})
		case "/v1.0/oauth2/accessToken":
			return jsonTestResponse(map[string]interface{}{
				"accessToken": "new-token",
				"expireIn":    7200,
			})
		case "/v1.0/robot/oToMessages/batchSend":
			robotAttempts++
			return jsonTestResponse(map[string]interface{}{
				"processQueryKey": "query-1",
			})
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"errcode": 404, "errmsg": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:     "corp-a",
		AppKey:     "app-key",
		AppSecret:  "app-secret",
		AgentID:    "342080997",
		NotifyMode: "agent_fallback",
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{
		Title:   "绩效流程待办",
		Content: "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。",
		URL:     "https://okr.example.com/dingtalk-h5/?view=performance%3Amanager",
	})
	if err != nil {
		t.Fatalf("SendWorkNotificationContext error = %v", err)
	}
	if agentAttempts != 1 {
		t.Fatalf("agent attempts = %d, want 1", agentAttempts)
	}
	if robotAttempts != 1 {
		t.Fatalf("robot attempts = %d, want 1", robotAttempts)
	}
}

func TestDingTalkOAPIClientSendsRobotWorkNotification(t *testing.T) {
	var sent map[string]interface{}
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v1.0/oauth2/accessToken":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode token body: %v", err)
			}
			if body["appKey"] != "app-key" || body["appSecret"] != "app-secret" {
				t.Fatalf("unexpected token body: %#v", body)
			}
			return jsonTestResponse(map[string]interface{}{
				"accessToken": "token-v1",
				"expireIn":    7200,
			})
		case "/v1.0/robot/oToMessages/batchSend":
			if r.Header.Get("x-acs-dingtalk-access-token") != "token-v1" {
				t.Fatalf("unexpected access token header: %s", r.Header.Get("x-acs-dingtalk-access-token"))
			}
			if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
				t.Fatalf("decode robot notification body: %v", err)
			}
			return jsonTestResponse(map[string]interface{}{
				"processQueryKey": "query-1",
			})
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"code": "not_found", "message": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:     "corp-a",
		AppKey:     "app-key",
		AppSecret:  "app-secret",
		NotifyMode: "robot",
		RobotCode:  "robot-code",
	}, []string{"manager001", "manager002"}, DingTalkWorkNotificationPayload{
		Content: "Lip 已提交 2026-07 月度考评，请处理。",
	})
	if err != nil {
		t.Fatalf("SendWorkNotificationContext error = %v", err)
	}
	if sent["robotCode"] != "robot-code" {
		t.Fatalf("robotCode = %#v, want robot-code", sent["robotCode"])
	}
	userIDs, ok := sent["userIds"].([]interface{})
	if !ok {
		t.Fatalf("userIds = %#v, want array", sent["userIds"])
	}
	if len(userIDs) != 2 || userIDs[0] != "manager001" || userIDs[1] != "manager002" {
		t.Fatalf("userIds = %#v, want [manager001 manager002]", sent["userIds"])
	}
	if sent["msgKey"] != "sampleText" {
		t.Fatalf("msgKey = %#v, want sampleText", sent["msgKey"])
	}
	var msgParam map[string]string
	if raw, ok := sent["msgParam"].(string); ok {
		if err := json.Unmarshal([]byte(raw), &msgParam); err != nil {
			t.Fatalf("decode msgParam: %v", err)
		}
	}
	if msgParam["content"] != "Lip 已提交 2026-07 月度考评，请处理。" {
		t.Fatalf("msgParam content = %#v", msgParam["content"])
	}
}

func TestDingTalkOAPIClientSendsRobotLinkNotification(t *testing.T) {
	var sent map[string]interface{}
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v1.0/oauth2/accessToken":
			return jsonTestResponse(map[string]interface{}{
				"accessToken": "token-v1",
				"expireIn":    7200,
			})
		case "/v1.0/robot/oToMessages/batchSend":
			if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
				t.Fatalf("decode robot notification body: %v", err)
			}
			return jsonTestResponse(map[string]interface{}{
				"processQueryKey": "query-1",
			})
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"code": "not_found", "message": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:     "corp-a",
		AppKey:     "app-key",
		AppSecret:  "app-secret",
		NotifyMode: "robot",
		RobotCode:  "robot-code",
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{
		Title:      "绩效流程待办",
		Content:    "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。",
		URL:        "https://oa.example.com/dingtalk-h5/?view=performance%3Amanager&reviewNo=lip-2026-07",
		PicURL:     "https://oa.example.com/logo.png",
		SourceName: "钉米-OKR",
	})
	if err != nil {
		t.Fatalf("SendWorkNotificationContext error = %v", err)
	}
	if sent["msgKey"] != "sampleLink" {
		t.Fatalf("msgKey = %#v, want sampleLink", sent["msgKey"])
	}
	var msgParam map[string]string
	if raw, ok := sent["msgParam"].(string); ok {
		if err := json.Unmarshal([]byte(raw), &msgParam); err != nil {
			t.Fatalf("decode msgParam: %v", err)
		}
	}
	if msgParam["title"] != "绩效流程待办" || msgParam["text"] == "" || msgParam["messageUrl"] == "" {
		t.Fatalf("msgParam = %#v, want title/text/messageUrl", msgParam)
	}
	if msgParam["picUrl"] != "https://oa.example.com/logo.png" {
		t.Fatalf("picUrl = %q, want configured logo url", msgParam["picUrl"])
	}
	if msgParam["sourceName"] != "钉米-OKR" {
		t.Fatalf("sourceName = %q, want 钉米-OKR", msgParam["sourceName"])
	}
}

func TestDingTalkRobotNotificationMessageUsesActionCardWhenRequested(t *testing.T) {
	msgKey, rawMsgParam := dingTalkRobotNotificationMessage(DingTalkWorkNotificationPayload{
		MessageType: DingTalkMessageTypeActionCard,
		Title:       "《绩效考评单》有新评论",
		Content:     "David 在“上级评价”评论：请关注本次评分",
		URL:         "https://oa.example.com/h5?view=workflow%3Ainstance%3Ainstance-1",
	})
	if msgKey != "sampleActionCard" {
		t.Fatalf("msgKey = %q, want sampleActionCard", msgKey)
	}
	var msgParam map[string]string
	if err := json.Unmarshal([]byte(rawMsgParam), &msgParam); err != nil {
		t.Fatalf("decode msgParam: %v", err)
	}
	if msgParam["title"] == "" || msgParam["text"] == "" {
		t.Fatalf("action card content = %#v", msgParam)
	}
	if msgParam["singleTitle"] != "查看流程" || msgParam["singleURL"] == "" {
		t.Fatalf("action card button = %#v", msgParam)
	}
}

func TestDingTalkRobotLinkNotificationCarriesConfiguredAppName(t *testing.T) {
	_, rawMsgParam := dingTalkRobotNotificationMessage(DingTalkWorkNotificationPayload{
		Title:      "绩效流程待办",
		Content:    "Lip 的 2026-07 月度考评已流转到「上级评价」，请及时处理。",
		URL:        "dingtalk://dingtalkclient/action/openapp?corpid=ding-corp",
		SourceName: "钉米-OKR",
		PicURL:     "https://oa.example.com/logo.png",
	})
	var msgParam map[string]string
	if err := json.Unmarshal([]byte(rawMsgParam), &msgParam); err != nil {
		t.Fatalf("decode msgParam: %v", err)
	}
	if msgParam["sourceName"] != "钉米-OKR" {
		t.Fatalf("sourceName = %q, want 钉米-OKR", msgParam["sourceName"])
	}
	if msgParam["picUrl"] != "https://oa.example.com/logo.png" {
		t.Fatalf("picUrl = %q, want configured logo url", msgParam["picUrl"])
	}
}

func TestDingTalkOAPIClientIncludesHTTPErrorBody(t *testing.T) {
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v1.0/oauth2/accessToken":
			return jsonTestResponse(map[string]interface{}{
				"accessToken": "token-v1",
				"expireIn":    7200,
			})
		case "/v1.0/robot/oToMessages/batchSend":
			var body bytes.Buffer
			_ = json.NewEncoder(&body).Encode(map[string]interface{}{
				"code":    "invalidParameter.userIds.empty",
				"message": "用户userId列表为空",
			})
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     make(http.Header),
				Body:       io.NopCloser(&body),
			}, nil
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{"code": "not_found", "message": "not found"})
	})}}

	err := client.SendWorkNotificationContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:     "corp-a",
		AppKey:     "app-key",
		AppSecret:  "app-secret",
		NotifyMode: "robot",
		RobotCode:  "robot-code",
	}, []string{"manager001"}, DingTalkWorkNotificationPayload{Content: "通知内容"})
	if err == nil {
		t.Fatalf("SendWorkNotificationContext should fail")
	}
	text := err.Error()
	for _, want := range []string{"HTTP 400", "invalidParameter.userIds.empty", "用户userId列表为空"} {
		if !strings.Contains(text, want) {
			t.Fatalf("error = %q, want %q", text, want)
		}
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
