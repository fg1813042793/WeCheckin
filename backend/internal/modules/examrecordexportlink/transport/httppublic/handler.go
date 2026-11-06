package httppublic

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	exportlinkapp "wecheckin/backend/internal/modules/examrecordexportlink/application"
	examservice "wecheckin/backend/internal/service/client/exam"
	"wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
	"wecheckin/backend/pkg/logger"
)

type Resolver interface {
	Resolve(context.Context, string) (exportlinkapp.Grant, error)
}

type Exporter interface {
	ExportRecordForLinkContext(context.Context, uint, uint, workflowsummary.ExportFormat) (*workflowsummary.ExportResult, error)
}

type Handler struct {
	resolver Resolver
	exporter Exporter
}

func NewHandler(resolver Resolver, exporter Exporter) *Handler {
	return &Handler{resolver: resolver, exporter: exporter}
}

func (handler *Handler) Download(ctx context.Context, c *app.RequestContext) {
	c.Header("Cache-Control", "no-store")
	if handler == nil || handler.resolver == nil || handler.exporter == nil {
		downloadInternalError("exam record export link handler unavailable")
		c.String(http.StatusInternalServerError, "答卷下载失败，请稍后重试")
		return
	}
	grant, err := handler.resolver.Resolve(ctx, strings.TrimSpace(c.Query("token")))
	if err != nil {
		if errors.Is(err, exportlinkapp.ErrGrantNotFound) || errors.Is(err, exportlinkapp.ErrGrantExpired) {
			c.String(http.StatusGone, "答卷下载链接已失效")
			return
		}
		downloadInternalError(err.Error())
		c.String(http.StatusInternalServerError, "答卷下载失败，请稍后重试")
		return
	}
	result, err := handler.exporter.ExportRecordForLinkContext(ctx, grant.ExamID, grant.RecordID, workflowsummary.ExportFormat(grant.Format))
	if errors.Is(err, examservice.ErrRecordExportUnavailable) {
		c.String(http.StatusGone, "答卷下载链接已失效")
		return
	}
	if err != nil {
		downloadInternalError(err.Error())
		c.String(http.StatusInternalServerError, "答卷下载失败，请稍后重试")
		return
	}
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", `attachment; filename="`+result.Filename+`"`)
	c.Write(result.Body)
}

func downloadInternalError(message string) {
	if logger.Logger != nil {
		logger.Logger.Printf("[EXAM_RECORD_EXPORT_LINK_ERROR] %s", message)
		return
	}
	log.Printf("[EXAM_RECORD_EXPORT_LINK_ERROR] %s", message)
}
