package review

import (
	"context"

	"wecheckin/backend/internal/model"
	templatesvc "wecheckin/backend/internal/service/dingtalkh5/performance/template"
)

func TemplateContext(ctx context.Context) (TemplateDTO, error) {
	return templatesvc.TemplateContext(ctx)
}

func LoadTemplateContext(ctx context.Context) (TemplateDTO, error) {
	return templatesvc.LoadTemplateContext(ctx)
}

func SaveTemplateContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload TemplateDTO) (TemplateDTO, error) {
	return templatesvc.SaveTemplateContext(ctx, user, payload)
}

func EnsureSeedContext(ctx context.Context) error {
	return templatesvc.EnsureSeedContext(ctx)
}
