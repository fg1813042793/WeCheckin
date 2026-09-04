package exam

import (
	"context"

	"wecheckin/backend/internal/model"
	examservice "wecheckin/backend/internal/service/client/exam"
)

func checkExamLimit(e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	return checkExamLimitContext(context.Background(), e, uidStr, device, deviceId, ip)
}

func checkExamLimitContext(ctx context.Context, e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	msg, _ := examservice.NewService().CheckLimitContext(ctx, e, uidStr, device, deviceId, ip)
	return msg
}
