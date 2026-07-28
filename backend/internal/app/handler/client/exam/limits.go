package exam

import (
	"context"

	examservice "wecheckin-backend/backend/internal/app/service/exam"
	"wecheckin-backend/backend/internal/model"
)

func checkExamLimit(e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	return checkExamLimitContext(context.Background(), e, uidStr, device, deviceId, ip)
}

func checkExamLimitContext(ctx context.Context, e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	msg, _ := examservice.NewService().CheckLimitContext(ctx, e, uidStr, device, deviceId, ip)
	return msg
}
