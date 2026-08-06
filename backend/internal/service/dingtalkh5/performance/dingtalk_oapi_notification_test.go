package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
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

func TestBuildReviewTransitionNotificationPayloadCarriesAppName(t *testing.T) {
	review := reviewModelForNotificationURL("lip-2026-07", "2026-07", ReviewStatusManagerReview)
	got := buildReviewTransitionNotificationPayloadContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:     "ding-corp",
		NotifyMode: "robot",
		AgentID:    "123456",
	}, review, nil)
	if got.SourceName != defaultDingTalkH5AppName {
		t.Fatalf("sourceName = %q, want %q", got.SourceName, defaultDingTalkH5AppName)
	}
}

func TestBuildReviewOperationURLAppendsDeepLinkParams(t *testing.T) {
	got := buildReviewOperationURL("https://oa.example.com/dingtalk-h5/?corpId=corp-a", reviewModelForNotificationURL("lip-2026-07", "2026-07", ReviewStatusManagerReview))
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse operation url: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "oa.example.com" || parsed.Path != "/dingtalk-h5/" {
		t.Fatalf("operation url = %q", got)
	}
	query := parsed.Query()
	if query.Get("corpId") != "corp-a" {
		t.Fatalf("corpId = %q, want corp-a", query.Get("corpId"))
	}
	if query.Get("view") != "performance:manager" {
		t.Fatalf("view = %q, want performance:manager", query.Get("view"))
	}
	if query.Get("reviewNo") != "lip-2026-07" {
		t.Fatalf("reviewNo = %q, want lip-2026-07", query.Get("reviewNo"))
	}
	if query.Get("period") != "2026-07" {
		t.Fatalf("period = %q, want 2026-07", query.Get("period"))
	}
}

func TestBuildReviewNotificationURLWrapsDingTalkOpenAppLink(t *testing.T) {
	review := reviewModelForNotificationURL("lip-2026-07", "2026-07", ReviewStatusManagerReview)
	got := buildReviewNotificationURL("https://oa.example.com/dingtalk-h5/?corpId=corp-a", DingTalkH5CorpConfig{
		CorpID:  "ding-corp",
		AgentID: "123456",
	}, review)

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse notification url: %v", err)
	}
	if parsed.Scheme != "dingtalk" || parsed.Host != "dingtalkclient" || parsed.Path != "/action/openapp" {
		t.Fatalf("notification url = %q, want DingTalk openapp scheme", got)
	}
	query := parsed.Query()
	if query.Get("corpid") != "ding-corp" {
		t.Fatalf("corpid = %q, want ding-corp", query.Get("corpid"))
	}
	if query.Get("container_type") != "work_platform" {
		t.Fatalf("container_type = %q, want work_platform", query.Get("container_type"))
	}
	if query.Get("app_id") != "0_123456" {
		t.Fatalf("app_id = %q, want 0_123456", query.Get("app_id"))
	}
	if query.Get("redirect_type") != "jump" {
		t.Fatalf("redirect_type = %q, want jump", query.Get("redirect_type"))
	}

	redirect, err := url.Parse(query.Get("redirect_url"))
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	if redirect.Scheme != "https" || redirect.Host != "oa.example.com" || redirect.Path != "/dingtalk-h5/" {
		t.Fatalf("redirect_url = %q", query.Get("redirect_url"))
	}
	if redirect.Query().Get("view") != "performance:manager" || redirect.Query().Get("reviewNo") != "lip-2026-07" {
		t.Fatalf("redirect query = %s", redirect.RawQuery)
	}
}

func TestBuildReviewNotificationPayloadUsesEnterpriseAppURL(t *testing.T) {
	review := reviewModelForNotificationURL("lip-2026-07", "2026-07", ReviewStatusHRBPReview)
	got := buildReviewTransitionNotificationPayloadContext(context.Background(), DingTalkH5CorpConfig{
		CorpID:       "ding-corp",
		NotifyMode:   "robot",
		AgentID:      "123456",
		UnifiedAppID: "dingmi-okr-app",
		AppURL:       "https://okr.example.com/dingtalk-h5/?corpId=ding-corp",
	}, review, nil)

	parsed, err := url.Parse(got.URL)
	if err != nil {
		t.Fatalf("parse notification url: %v", err)
	}
	if parsed.Scheme != "dingtalk" || parsed.Host != "dingtalkclient" {
		t.Fatalf("notification url = %q, want DingTalk openapp link", got.URL)
	}
	query := parsed.Query()
	if query.Get("app_id") != "dingmi-okr-app" {
		t.Fatalf("app_id = %q, want dingmi-okr-app", query.Get("app_id"))
	}
	redirect, err := url.Parse(query.Get("redirect_url"))
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	if redirect.Host != "okr.example.com" || redirect.Path != "/dingtalk-h5/" {
		t.Fatalf("redirect_url = %q, want enterprise app url", query.Get("redirect_url"))
	}
	if redirect.Query().Get("view") != "performance:hrbp" || redirect.Query().Get("reviewNo") != "lip-2026-07" {
		t.Fatalf("redirect query = %s", redirect.RawQuery)
	}
}

func TestDingTalkH5OpenAppIDPrefersUnifiedAppIDForRobotNotification(t *testing.T) {
	got := dingtalkH5OpenAppID(DingTalkH5CorpConfig{
		NotifyMode:   "robot",
		AgentID:      "123456",
		UnifiedAppID: "dingmi-okr-app",
	})
	if got != "dingmi-okr-app" {
		t.Fatalf("open app id = %q, want unified app id", got)
	}
}

func TestDingTalkH5OpenAppIDPrefersUnifiedAppIDForAgentNotification(t *testing.T) {
	got := dingtalkH5OpenAppID(DingTalkH5CorpConfig{
		NotifyMode:   "agent",
		AgentID:      "123456",
		UnifiedAppID: "dingmi-okr-app",
	})
	if got != "dingmi-okr-app" {
		t.Fatalf("open app id = %q, want unified app id", got)
	}
}

func TestBuildReviewNotificationURLFallsBackToRawDeepLinkWithoutDingTalkAppID(t *testing.T) {
	review := reviewModelForNotificationURL("lip-2026-07", "2026-07", ReviewStatusManagerReview)
	got := buildReviewNotificationURL("https://oa.example.com/dingtalk-h5/", DingTalkH5CorpConfig{CorpID: "ding-corp"}, review)

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse notification url: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "oa.example.com" || parsed.Path != "/dingtalk-h5/" {
		t.Fatalf("notification url = %q, want raw H5 deep link", got)
	}
	if parsed.Query().Get("reviewNo") != "lip-2026-07" {
		t.Fatalf("reviewNo = %q, want lip-2026-07", parsed.Query().Get("reviewNo"))
	}
}

func reviewModelForNotificationURL(reviewNo, period, status string) model.DingTalkH5PerfReview {
	return model.DingTalkH5PerfReview{
		ReviewNo: reviewNo,
		Period:   period,
		Status:   status,
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
