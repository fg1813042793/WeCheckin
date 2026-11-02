package exam

import (
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/useragent"
)

// ClientEnvironment is captured at the HTTP boundary, not read from submitted IP fields.
type ClientEnvironment struct {
	IP     string
	Device string
}

func normalizeRecordDeviceMetadata(record *model.ExamRecord) {
	if record == nil {
		return
	}
	record.Browser, record.DeviceType, record.PlatformType = useragent.Parse(record.Device)
}

func normalizeRecordListDeviceMetadata(records []model.ExamRecord) {
	for i := range records {
		normalizeRecordDeviceMetadata(&records[i])
	}
}
