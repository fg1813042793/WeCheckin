package workflowsummary

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const maxExportBodyBytes = 64 << 20

type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatDOCX ExportFormat = "docx"
)

var (
	ErrExportFormatInvalid = errors.New("导出格式仅支持 pdf、xlsx、docx")
	ErrExportBodyTooLarge  = errors.New("导出文件过大，请减少本次导出数量")
)

type ExportResult struct {
	Filename    string
	ContentType string
	Body        []byte
}

func (format ExportFormat) Valid() bool {
	return format == ExportFormatPDF || format == ExportFormatXLSX || format == ExportFormatDOCX
}

func renderWorkflowExport(documents []exportDocument, format ExportFormat) (*ExportResult, error) {
	if len(documents) == 0 {
		return nil, ErrExportInstancesEmpty
	}
	if !format.Valid() {
		return nil, ErrExportFormatInvalid
	}
	date := time.Now().Format("2006-01-02")
	if format == ExportFormatXLSX {
		body, err := renderXLSX(documents)
		if err != nil {
			return nil, err
		}
		return checkedExportResult("workflow-summary-"+date+".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", body)
	}
	if len(documents) == 1 {
		body, err := renderSingleDocument(documents[0], format)
		if err != nil {
			return nil, err
		}
		return checkedExportResult(
			documents[0].FileStem+"."+string(format),
			exportContentType(format),
			body,
		)
	}

	stems := uniqueFileStems(documents)
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for index, document := range documents {
		body, err := renderSingleDocument(document, format)
		if err != nil {
			_ = archive.Close()
			return nil, err
		}
		entry, err := archive.Create(filepath.ToSlash(stems[index] + "." + string(format)))
		if err != nil {
			_ = archive.Close()
			return nil, err
		}
		if _, err := entry.Write(body); err != nil {
			_ = archive.Close()
			return nil, err
		}
		if buffer.Len() > maxExportBodyBytes {
			_ = archive.Close()
			return nil, ErrExportBodyTooLarge
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return checkedExportResult("workflow-summary-"+date+".zip", "application/zip", buffer.Bytes())
}

func renderSingleDocument(document exportDocument, format ExportFormat) ([]byte, error) {
	switch format {
	case ExportFormatPDF:
		return renderPDF(document)
	case ExportFormatDOCX:
		return renderDOCX(document)
	default:
		return nil, ErrExportFormatInvalid
	}
}

func exportContentType(format ExportFormat) string {
	switch format {
	case ExportFormatPDF:
		return "application/pdf"
	case ExportFormatDOCX:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ExportFormatXLSX:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

func checkedExportResult(filename, contentType string, body []byte) (*ExportResult, error) {
	if len(body) > maxExportBodyBytes {
		return nil, ErrExportBodyTooLarge
	}
	return &ExportResult{Filename: filename, ContentType: contentType, Body: body}, nil
}

func writeZipText(archive *zip.Writer, name, content string) error {
	entry, err := archive.Create(name)
	if err != nil {
		return err
	}
	_, err = entry.Write([]byte(content))
	return err
}

func xmlText(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case '&':
			builder.WriteString("&amp;")
		case '<':
			builder.WriteString("&lt;")
		case '>':
			builder.WriteString("&gt;")
		case '"':
			builder.WriteString("&quot;")
		case '\'':
			builder.WriteString("&apos;")
		default:
			if r == '\t' || r == '\n' || r == '\r' || r >= 32 {
				builder.WriteRune(r)
			}
		}
	}
	return builder.String()
}

func exportError(format ExportFormat, err error) error {
	return fmt.Errorf("生成 %s 文件失败: %w", format, err)
}
