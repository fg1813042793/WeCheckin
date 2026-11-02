package workflowsummary_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
)

func TestRenderDocumentsSupportsPublicPDFDOCXAndXLSX(t *testing.T) {
	document := workflowsummary.Document{
		Title:    "考试答卷",
		FileStem: "exam-record-1",
		Metadata: [][]string{{"考生", "张三"}},
		Sections: []workflowsummary.Section{{
			Title:   "答题明细",
			Headers: []string{"题目", "用户答案", "正确答案", "判题结果"},
			Rows:    [][]string{{"1. 首都", "北京", "北京", "正确"}},
		}},
	}

	for _, format := range []workflowsummary.ExportFormat{
		workflowsummary.ExportFormatPDF,
		workflowsummary.ExportFormatDOCX,
		workflowsummary.ExportFormatXLSX,
	} {
		t.Run(string(format), func(t *testing.T) {
			result, err := workflowsummary.RenderDocuments([]workflowsummary.Document{document}, format)
			if err != nil {
				t.Fatalf("render %s: %v", format, err)
			}
			if !strings.HasSuffix(result.Filename, "."+string(format)) {
				t.Fatalf("filename = %q", result.Filename)
			}
			switch format {
			case workflowsummary.ExportFormatPDF:
				if !bytes.HasPrefix(result.Body, []byte("%PDF-1.4")) || !bytes.Contains(result.Body, []byte("%%EOF")) {
					t.Fatal("invalid PDF body")
				}
			case workflowsummary.ExportFormatDOCX:
				if !strings.Contains(readExportZipEntry(t, result.Body, "word/document.xml"), "考试答卷") {
					t.Fatal("DOCX missing title")
				}
			case workflowsummary.ExportFormatXLSX:
				if !strings.Contains(readExportZipEntry(t, result.Body, "xl/worksheets/sheet1.xml"), "考试答卷") {
					t.Fatal("XLSX missing title")
				}
			}
		})
	}
}

func readExportZipEntry(t *testing.T, body []byte, name string) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		data, err := io.ReadAll(stream)
		_ = stream.Close()
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(data)
	}
	t.Fatalf("zip entry %s not found", name)
	return ""
}
