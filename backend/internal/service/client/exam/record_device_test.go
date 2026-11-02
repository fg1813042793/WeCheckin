package exam

import (
	"encoding/json"
	"sync"
	"testing"

	"gorm.io/gorm/schema"

	"wecheckin/backend/internal/model"
)

func TestRecordDetailIncludesClientMetadataFromStoredUserAgent(t *testing.T) {
	tests := []struct {
		name, ua, browser, deviceType, platform string
	}{
		{"desktop", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/130.0 Safari/537.36 Edg/130.0", "Edge", "PC端", "Windows"},
		{"dingtalk iphone", "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 AliApp(DingTalk/7.6.0)", "钉钉", "移动端", "iOS"},
		{"wechat harmony", "Mozilla/5.0 (Phone; OpenHarmony 6.1) AppleWebKit/537.36 Chrome/132.0 Mobile MicroMessenger/8.0", "微信", "移动端", "OpenHarmony"},
		{"legacy missing", "", "-", "-", "-"},
		{"unknown", "CustomClient/1.0", "-", "-", "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := model.ExamRecord{ID: 19, Device: tt.ua, AddIP: "192.0.2.10"}
			result := decodeRecordDetailPayload(record, model.Exam{})
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			var payload struct {
				Record map[string]interface{} `json:"record"`
			}
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]string{"browser": tt.browser, "deviceType": tt.deviceType, "platformType": tt.platform, "device": tt.ua, "addIp": "192.0.2.10"} {
				if got := payload.Record[key]; got != want {
					t.Errorf("record.%s = %v, want %q", key, got, want)
				}
			}
		})
	}
}

func TestRecordClientMetadataDoesNotRequireDatabaseColumns(t *testing.T) {
	parsed, err := schema.Parse(&model.ExamRecord{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Browser", "DeviceType", "PlatformType"} {
		field := parsed.LookUpField(name)
		if field == nil || field.DBName != "" || field.Creatable || field.Updatable {
			t.Fatalf("%s must be derived and never included in persistence", name)
		}
	}
	records := []model.ExamRecord{{Device: "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 Version/18 Safari/605.1.15"}, {}}
	normalizeRecordListDeviceMetadata(records)
	if records[0].Browser != "Safari" || records[0].DeviceType != "PC端" || records[0].PlatformType != "macOS" || records[1].Browser != "-" {
		t.Fatalf("list metadata not populated: %+v", records)
	}
}
