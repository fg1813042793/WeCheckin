package workflowsummary

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

func TestRenderSingleWorkflowExportFormats(t *testing.T) {
	document := sampleExportDocument("OA-001")
	tests := []struct {
		format      ExportFormat
		contentType string
		extension   string
		check       func(*testing.T, []byte)
	}{
		{ExportFormatPDF, "application/pdf", ".pdf", func(t *testing.T, body []byte) {
			if !bytes.HasPrefix(body, []byte("%PDF-1.4")) || !bytes.Contains(body, []byte("%%EOF")) {
				t.Fatalf("invalid PDF body")
			}
		}},
		{ExportFormatDOCX, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", ".docx", func(t *testing.T, body []byte) {
			entries := readZipEntries(t, body)
			if !strings.Contains(entries["word/document.xml"], "OA-001") {
				t.Fatalf("DOCX document missing business key")
			}
		}},
		{ExportFormatXLSX, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".xlsx", func(t *testing.T, body []byte) {
			entries := readZipEntries(t, body)
			if !strings.Contains(entries["xl/worksheets/sheet1.xml"], "OA-001") {
				t.Fatalf("XLSX sheet missing business key")
			}
		}},
	}
	for _, test := range tests {
		t.Run(string(test.format), func(t *testing.T) {
			result, err := renderWorkflowExport([]exportDocument{document}, test.format)
			if err != nil {
				t.Fatalf("render export: %v", err)
			}
			if result.ContentType != test.contentType || !strings.HasSuffix(result.Filename, test.extension) {
				t.Fatalf("result = %#v", result)
			}
			test.check(t, result.Body)
		})
	}
}

func TestBuildExportDocumentUsesInstanceVersionAndFormSchema(t *testing.T) {
	document := buildExportDocument(&workflowapp.InstanceDetail{
		Instance: workflowapp.InstanceSummary{
			ID: "instance-1", DefinitionName: "采购审批", DefinitionVersion: 3,
			BusinessKey: "PO-001", StarterName: "张三", Status: "completed",
		},
		Form: []workflowcore.FormField{
			{Key: "reason", Label: "采购原因", Type: workflowcore.FormFieldTypeText},
			{Key: "items", Label: "采购明细", Type: workflowcore.FormFieldTypeDetailList, Columns: []workflowcore.FormField{
				{Key: "name", Label: "物品", Type: workflowcore.FormFieldTypeText},
			}},
		},
		FormData: map[string]interface{}{
			"reason": "项目使用",
			"items":  []interface{}{map[string]interface{}{"name": "显示器"}},
		},
	})
	flat := document.Title
	for _, row := range document.Metadata {
		flat += strings.Join(row, "")
	}
	for _, section := range document.Sections {
		flat += section.Title + strings.Join(section.Headers, "")
		for _, fieldRow := range section.FieldRows {
			for _, field := range fieldRow {
				flat += field.Label + field.Value
			}
		}
		for _, row := range section.Rows {
			flat += strings.Join(row, "")
		}
	}
	for _, want := range []string{"采购审批", "v3", "采购原因", "项目使用", "采购明细", "显示器"} {
		if !strings.Contains(flat, want) {
			t.Fatalf("export document missing %q: %s", want, flat)
		}
	}
}

func TestBuildExportDocumentPreservesFormGrid(t *testing.T) {
	document := buildExportDocument(&workflowapp.InstanceDetail{
		Instance: workflowapp.InstanceSummary{ID: "instance-1", DefinitionName: "绩效考评单", DefinitionVersion: 2},
		Form: []workflowcore.FormField{{
			Key: "review", Label: "价值观自评", Type: workflowcore.FormFieldTypeGroup,
			Fields: []workflowcore.FormField{
				{Key: "team", Label: "团结一心", Type: workflowcore.FormFieldTypeNumber, Span: 8},
				{Key: "innovation", Label: "开拓创新", Type: workflowcore.FormFieldTypeNumber, Span: 8},
				{Key: "persistence", Label: "坚韧不拔", Type: workflowcore.FormFieldTypeNumber, Span: 8},
				{Key: "comment", Label: "评价", Type: workflowcore.FormFieldTypeTextarea, Span: 24},
			},
		}},
		FormData: map[string]interface{}{
			"team": 50, "innovation": 60, "persistence": 70, "comment": "表现稳定",
		},
	})

	if len(document.Sections) < 1 || len(document.Sections[0].FieldRows) != 2 {
		t.Fatalf("field rows = %#v, want two layout rows", document.Sections)
	}
	firstRow := document.Sections[0].FieldRows[0]
	if len(firstRow) != 3 || firstRow[0].Span != 8 || firstRow[1].Span != 8 || firstRow[2].Span != 8 {
		t.Fatalf("first field row = %#v, want three one-third fields", firstRow)
	}
	secondRow := document.Sections[0].FieldRows[1]
	if len(secondRow) != 1 || secondRow[0].Span != 24 || secondRow[0].Value != "表现稳定" {
		t.Fatalf("second field row = %#v, want one full-width field", secondRow)
	}
}

func TestWorkflowExportFormatsContainStructuredStyles(t *testing.T) {
	document := buildExportDocument(&workflowapp.InstanceDetail{
		Instance: workflowapp.InstanceSummary{ID: "instance-1", DefinitionName: "绩效考评单", DefinitionVersion: 2},
		Form: []workflowcore.FormField{{
			Key: "review", Label: "价值观自评", Type: workflowcore.FormFieldTypeGroup,
			Fields: []workflowcore.FormField{
				{Key: "team", Label: "团结一心", Type: workflowcore.FormFieldTypeNumber, Span: 8},
				{Key: "innovation", Label: "开拓创新", Type: workflowcore.FormFieldTypeNumber, Span: 8},
				{Key: "persistence", Label: "坚韧不拔", Type: workflowcore.FormFieldTypeNumber, Span: 8},
			},
		}},
		FormData: map[string]interface{}{"team": 50, "innovation": 60, "persistence": 70},
	})

	pdf, err := renderPDF(document)
	if err != nil {
		t.Fatalf("render PDF: %v", err)
	}
	for _, want := range []string{"/BaseFont /Helvetica", "/F2", " re f", " re S"} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Fatalf("PDF missing structured style command %q", want)
		}
	}

	xlsx, err := renderXLSX([]exportDocument{document})
	if err != nil {
		t.Fatalf("render XLSX: %v", err)
	}
	xlsxEntries := readZipEntries(t, xlsx)
	for _, want := range []string{`<mergeCell ref="A1:X1"/>`, `customHeight="1"`, `<pageSetup fitToWidth="1"`} {
		if !strings.Contains(xlsxEntries["xl/worksheets/sheet1.xml"], want) {
			t.Fatalf("XLSX worksheet missing %q", want)
		}
	}
	if !strings.Contains(xlsxEntries["xl/styles.xml"], "Microsoft YaHei") {
		t.Fatalf("XLSX styles should use a readable CJK font")
	}

	docx, err := renderDOCX(document)
	if err != nil {
		t.Fatalf("render DOCX: %v", err)
	}
	docxEntries := readZipEntries(t, docx)
	for _, want := range []string{`<w:tblLayout w:type="fixed"/>`, `<w:gridSpan w:val="8"/>`, `<w:shd w:fill="E8F3F1"/>`} {
		if !strings.Contains(docxEntries["word/document.xml"], want) {
			t.Fatalf("DOCX document missing %q", want)
		}
	}
}

func TestRenderBatchWorkflowExports(t *testing.T) {
	documents := []exportDocument{sampleExportDocument("OA-001"), sampleExportDocument("OA-002")}

	xlsx, err := renderWorkflowExport(documents, ExportFormatXLSX)
	if err != nil {
		t.Fatalf("render xlsx: %v", err)
	}
	xlsxEntries := readZipEntries(t, xlsx.Body)
	if !strings.Contains(xlsxEntries["xl/workbook.xml"], "sheetId=\"2\"") {
		t.Fatalf("batch XLSX should contain one sheet per instance")
	}

	for _, format := range []ExportFormat{ExportFormatPDF, ExportFormatDOCX} {
		result, err := renderWorkflowExport(documents, format)
		if err != nil {
			t.Fatalf("render %s: %v", format, err)
		}
		if result.ContentType != "application/zip" || !strings.HasSuffix(result.Filename, ".zip") {
			t.Fatalf("batch result = %#v", result)
		}
		entries := readZipEntries(t, result.Body)
		if len(entries) != 2 {
			t.Fatalf("ZIP entries = %d, want 2", len(entries))
		}
	}
}

func sampleExportDocument(businessKey string) exportDocument {
	return exportDocument{
		Title:    "请假审批",
		FileStem: businessKey,
		Metadata: [][]string{{"申请编号", businessKey}, {"发起人", "张三"}},
		Sections: []exportSection{
			{Title: "表单内容", Headers: []string{"字段", "内容"}, Rows: [][]string{{"请假原因", "测试"}}},
			{Title: "审批记录", Headers: []string{"节点", "处理人"}, Rows: [][]string{{"主管审批", "李四"}}},
		},
	}
}

func readZipEntries(t *testing.T, body []byte) map[string]string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	entries := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", file.Name, err)
		}
		data, err := io.ReadAll(stream)
		_ = stream.Close()
		if err != nil {
			t.Fatalf("read %s: %v", file.Name, err)
		}
		entries[file.Name] = string(data)
	}
	return entries
}
