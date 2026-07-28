package model

import "testing"

func TestDingTalkH5PerformanceUserUsesSharedUsersTable(t *testing.T) {
	cases := map[string]string{
		"user":     (DingTalkH5PerfUser{}).TableName(),
		"session":  (DingTalkH5PerfSession{}).TableName(),
		"review":   (DingTalkH5PerfReview{}).TableName(),
		"history":  (DingTalkH5PerfHistory{}).TableName(),
		"template": (DingTalkH5PerfTemplate{}).TableName(),
	}
	want := map[string]string{
		"user":     "users",
		"session":  "dingtalk_h5_perf_sessions",
		"review":   "dingtalk_h5_perf_reviews",
		"history":  "dingtalk_h5_perf_histories",
		"template": "dingtalk_h5_perf_templates",
	}
	for name, got := range cases {
		if got != want[name] {
			t.Fatalf("%s table = %q, want %q", name, got, want[name])
		}
	}
}
