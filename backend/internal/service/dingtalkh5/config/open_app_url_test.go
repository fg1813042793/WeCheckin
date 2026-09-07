package config

import (
	"net/url"
	"testing"
)

func TestWrapDingTalkOpenAppURLPrefersUnifiedAppID(t *testing.T) {
	got := WrapDingTalkOpenAppURL("https://oa.example.com/dingtalk-h5/", DingTalkH5CorpConfig{
		CorpID: "ding-corp", AgentID: "123456", UnifiedAppID: "ding-unified",
	})
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse openapp url: %v", err)
	}
	if parsed.Query().Get("app_id") != "ding-unified" {
		t.Fatalf("app_id = %q, want unified app id", parsed.Query().Get("app_id"))
	}
}

func TestWrapDingTalkOpenAppURLKeepsWebURLWithoutAppIdentity(t *testing.T) {
	const operationURL = "https://oa.example.com/dingtalk-h5/"
	if got := WrapDingTalkOpenAppURL(operationURL, DingTalkH5CorpConfig{CorpID: "ding-corp"}); got != operationURL {
		t.Fatalf("url = %q, want web fallback", got)
	}
}
