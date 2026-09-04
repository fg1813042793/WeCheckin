package notification

import (
	"context"
	"net/url"
	"testing"

	"wecheckin/backend/internal/model"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
)

func TestBuildTransitionPayloadCarriesAppName(t *testing.T) {
	review := reviewModelForURL("lip-2026-07", "2026-07", reviewStatusManagerReview)
	got := buildTransitionPayloadContext(context.Background(), configsvc.DingTalkH5CorpConfig{
		CorpID:     "ding-corp",
		NotifyMode: "robot",
		AgentID:    "123456",
	}, review, nil, "")
	if got.SourceName != configsvc.DefaultDingTalkH5AppName {
		t.Fatalf("sourceName = %q, want %q", got.SourceName, configsvc.DefaultDingTalkH5AppName)
	}
}

func TestBuildOperationURLAppendsDeepLinkParams(t *testing.T) {
	got := buildOperationURL("https://oa.example.com/dingtalk-h5/?corpId=corp-a", reviewModelForURL("lip-2026-07", "2026-07", reviewStatusManagerReview))
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

func TestBuildNotificationURLWrapsDingTalkOpenAppLink(t *testing.T) {
	review := reviewModelForURL("lip-2026-07", "2026-07", reviewStatusManagerReview)
	got := buildNotificationURL("https://oa.example.com/dingtalk-h5/?corpId=corp-a", configsvc.DingTalkH5CorpConfig{
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

func TestBuildNotificationPayloadUsesEnterpriseAppURL(t *testing.T) {
	review := reviewModelForURL("lip-2026-07", "2026-07", reviewStatusHRBPReview)
	got := buildTransitionPayloadContext(context.Background(), configsvc.DingTalkH5CorpConfig{
		CorpID:       "ding-corp",
		NotifyMode:   "robot",
		AgentID:      "123456",
		UnifiedAppID: "dingmi-okr-app",
		AppURL:       "https://okr.example.com/dingtalk-h5/?corpId=ding-corp",
	}, review, nil, "")

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

func TestOpenAppIDPrefersUnifiedAppIDForRobotNotification(t *testing.T) {
	got := openAppID(configsvc.DingTalkH5CorpConfig{
		NotifyMode:   "robot",
		AgentID:      "123456",
		UnifiedAppID: "dingmi-okr-app",
	})
	if got != "dingmi-okr-app" {
		t.Fatalf("open app id = %q, want unified app id", got)
	}
}

func TestOpenAppIDPrefersUnifiedAppIDForAgentNotification(t *testing.T) {
	got := openAppID(configsvc.DingTalkH5CorpConfig{
		NotifyMode:   "agent",
		AgentID:      "123456",
		UnifiedAppID: "dingmi-okr-app",
	})
	if got != "dingmi-okr-app" {
		t.Fatalf("open app id = %q, want unified app id", got)
	}
}

func TestBuildNotificationURLFallsBackToRawDeepLinkWithoutDingTalkAppID(t *testing.T) {
	review := reviewModelForURL("lip-2026-07", "2026-07", reviewStatusManagerReview)
	got := buildNotificationURL("https://oa.example.com/dingtalk-h5/", configsvc.DingTalkH5CorpConfig{CorpID: "ding-corp"}, review)

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

func TestBuildTransitionContentUsesResolvedEmployeeName(t *testing.T) {
	review := reviewModelForURL("114-2026-09", "2026-09", reviewStatusHRBPReview)
	review.EmployeeAccount = "114"
	actor := &model.DingTalkH5PerfUser{Account: "manager-1", Name: "Alice"}

	got := buildTransitionContent(review, actor, "Phoebe")
	want := "Alice 已将 Phoebe 的 2026-09 月度考评流转到「HRBP评价」，请及时处理。"
	if got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func reviewModelForURL(reviewNo, period, status string) model.DingTalkH5PerfReview {
	return model.DingTalkH5PerfReview{
		ReviewNo: reviewNo,
		Period:   period,
		Status:   status,
	}
}
