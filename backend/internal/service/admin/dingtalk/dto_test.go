package dingtalk

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSettingsResponseKeepsStableJSONFields(t *testing.T) {
	encoded, err := json.Marshal(SettingsResponse{CorpConfigs: []CorpConfigResponse{}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, field := range []string{
		`"corpId":""`, `"appSecretSet":false`, `"corpConfigs":[]`, `"tokenExpire":""`,
		`"redisPrefix":""`, `"singleLogin":0`, `"selfBind":0`, `"notifyEnabled":0`,
		`"appName":""`, `"logoText":""`, `"logoUrl":""`, `"appUrl":""`,
	} {
		if !strings.Contains(text, field) {
			t.Fatalf("JSON %s missing %s", text, field)
		}
	}
}
