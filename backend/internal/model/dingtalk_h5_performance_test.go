package model

import (
	"reflect"
	"testing"
)

func TestDingTalkH5PerformanceUserUsesSharedUsersTable(t *testing.T) {
	cases := map[string]string{
		"user":        (DingTalkH5PerfUser{}).TableName(),
		"review":      (DingTalkH5PerfReview{}).TableName(),
		"history":     (DingTalkH5PerfHistory{}).TableName(),
		"template":    (DingTalkH5PerfTemplate{}).TableName(),
		"corpConfig":  (DingTalkH5CorpConfig{}).TableName(),
		"userBinding": (DingTalkH5UserBinding{}).TableName(),
	}
	want := map[string]string{
		"user":        "users",
		"review":      "dingtalk_h5_perf_reviews",
		"history":     "dingtalk_h5_perf_histories",
		"template":    "dingtalk_h5_perf_templates",
		"corpConfig":  "dingtalk_h5_corp_configs",
		"userBinding": "dingtalk_h5_user_bindings",
	}
	for name, got := range cases {
		if got != want[name] {
			t.Fatalf("%s table = %q, want %q", name, got, want[name])
		}
	}
}

func TestDingTalkH5MultiCorpModelsCarryRequiredFields(t *testing.T) {
	cases := map[string]struct {
		typ    reflect.Type
		fields []string
	}{
		"corp config": {
			typ: reflect.TypeOf(DingTalkH5CorpConfig{}),
			fields: []string{
				"CorpID",
				"CorpName",
				"AppKey",
				"AppSecret",
				"AgentID",
				"UnifiedAppID",
				"AppURL",
				"NotifyEnabled",
				"NotifyMode",
				"RobotCode",
				"Enabled",
				"AddTime",
				"EditTime",
			},
		},
		"user binding": {
			typ: reflect.TypeOf(DingTalkH5UserBinding{}),
			fields: []string{
				"CorpID",
				"DingTalkUserID",
				"UnionID",
				"UserID",
				"Enabled",
				"AddTime",
				"EditTime",
			},
		},
	}
	for name, item := range cases {
		for _, field := range item.fields {
			if _, ok := item.typ.FieldByName(field); !ok {
				t.Fatalf("%s missing field %s", name, field)
			}
		}
	}
}

func TestDingTalkH5PerformanceTablesCarryAuditAndDataScopeFields(t *testing.T) {
	requiredFields := []string{
		"CreateBy",
		"UpdateBy",
		"CreateDeptID",
		"UpdateDeptID",
		"DeleteBy",
		"DeleteDeptID",
		"DeletedAt",
		"AddTime",
		"EditTime",
		"CreatedAt",
		"UpdatedAt",
	}
	cases := map[string]reflect.Type{
		"review":   reflect.TypeOf(DingTalkH5PerfReview{}),
		"history":  reflect.TypeOf(DingTalkH5PerfHistory{}),
		"template": reflect.TypeOf(DingTalkH5PerfTemplate{}),
	}
	for name, typ := range cases {
		for _, field := range requiredFields {
			if _, ok := typ.FieldByName(field); !ok {
				t.Fatalf("%s missing audit/data-scope field %s", name, field)
			}
		}
	}
}
