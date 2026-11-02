package workflowsummary

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode/utf16"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

func TestExportContentFormatsAndHistoricalForms(t *testing.T) {
	for _, content := range []string{"", "full", "form"} {
		for _, format := range []ExportFormat{ExportFormatPDF, ExportFormatDOCX, ExportFormatXLSX} {
			for _, count := range []int{1, 2} {
				t.Run(fmt.Sprintf("%s/%s/%d", content, format, count), func(t *testing.T) {
					runtime, ids := contentExportFixture(count)
					visibility := workflowapp.InstanceVisibility{Ready: true, UserIDs: []uint{7}}
					service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: visibility})
					result, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, contentRequest(t, ids, format, content))
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(runtime.query.Visibility, &visibility) || runtime.detailCalls != count || runtime.query.DefinitionID != 0 {
						t.Fatalf("export must authorize selection and load record-bound details: query=%+v calls=%d", runtime.query, runtime.detailCalls)
					}
					bodies := exportedDocuments(t, result, format, count, content)
					for index, body := range bodies {
						text := exportedText(t, body, format)
						for _, want := range []string{"历史考评表", "历史字段" + []string{"甲", "乙"}[index], "历史值" + []string{"甲", "乙"}[index], "考评分组", "评分", "87", "HRBP意见", "建议晋升", "HRBP负责人", "王老师", "项目明细", "历史项目", "明细评分", "91", "计算总分", "178", "流程状态", "表单状态保留"} {
							if !strings.Contains(text, want) {
								t.Errorf("export missing actual form content %q: %s", want, text)
							}
						}
						for _, metadata := range []string{"流程名称", "申请编号", "流程版本", "发起人", "操作人", "发起时间", "完成时间", "BUSINESS-SECRET", "OPERATOR-SECRET", "已完成", "审批记录", "TASK-SECRET", "APPROVER-SECRET", "COMMENT-SECRET"} {
							if got := strings.Contains(text, metadata); got != (content != "form") {
								t.Errorf("metadata %q present=%v for content=%q", metadata, got, content)
							}
						}
						if content == "form" {
							if strings.Count(text, "历史考评表") != 1 || strings.Contains(text, "流程信息") || strings.Contains(text, "INSTANCE-SECRET") || strings.Contains(text, "2026-10-30") {
								t.Errorf("form must have one title and no metadata: %s", text)
							}
							assertFormStartsAfterTitle(t, body, format)
						}
						otherVersion := "历史字段" + []string{"乙", "甲"}[index]
						if strings.Contains(text, otherVersion) {
							t.Errorf("historical schema mixed with another record: %s", text)
						}
					}
				})
			}
		}
	}
}

func TestExportRejectsInvalidContentBeforeLoadingRecords(t *testing.T) {
	for _, content := range []string{"other", "FORM", " form "} {
		runtime, ids := contentExportFixture(1)
		service := NewServiceWithAccess(runtime, summaryAccessStub{})
		result, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, contentRequest(t, ids, ExportFormatPDF, content))
		if !errors.Is(err, ErrExportContentInvalid) || result != nil {
			t.Errorf("content %q: expected readable parameter error, got result=%v err=%v", content, result != nil, err)
		}
		if runtime.detailCalls != 0 || runtime.query.Visibility != nil {
			t.Errorf("invalid content loaded records")
		}
	}
}

func TestExportEmptyContentMatchesExplicitFull(t *testing.T) {
	for _, format := range []ExportFormat{ExportFormatPDF, ExportFormatDOCX, ExportFormatXLSX} {
		for _, count := range []int{1, 2} {
			runtime, ids := contentExportFixture(count)
			service := NewServiceWithAccess(runtime, summaryAccessStub{})
			user := &model.DingTalkH5PerfUser{ID: 7}
			legacy, err := service.Export(context.Background(), user, contentRequest(t, ids, format, ""))
			if err != nil {
				t.Fatal(err)
			}
			full, err := service.Export(context.Background(), user, contentRequest(t, ids, format, "full"))
			if err != nil {
				t.Fatal(err)
			}
			if legacy.Filename != full.Filename || legacy.ContentType != full.ContentType || !bytes.Equal(legacy.Body, full.Body) {
				t.Errorf("empty content differs from full for %s/%d", format, count)
			}
		}
	}
}

func TestExportContentPreservesAccessDenial(t *testing.T) {
	for _, content := range []string{"", "full", "form"} {
		for _, denied := range []string{"partial selection", "scope error", "anonymous"} {
			t.Run(content+"/"+denied, func(t *testing.T) {
				runtime, ids := contentExportFixture(2)
				access := summaryAccessStub{visibility: workflowapp.InstanceVisibility{Ready: true, UserIDs: []uint{7}}}
				user := &model.DingTalkH5PerfUser{ID: 7}
				switch denied {
				case "partial selection":
					runtime.list.List = runtime.list.List[:1]
				case "scope error":
					access.err = ErrSummaryAccessDenied
				case "anonymous":
					user = nil
				}
				result, err := NewServiceWithAccess(runtime, access).Export(context.Background(), user, contentRequest(t, ids, ExportFormatDOCX, content))
				if !errors.Is(err, ErrSummaryAccessDenied) || result != nil || runtime.detailCalls != 0 {
					t.Fatalf("denied export result=%v err=%v detailCalls=%d", result != nil, err, runtime.detailCalls)
				}
			})
		}
	}
}

func TestExportTitleCollisionsRemainUnique(t *testing.T) {
	documents := []exportDocument{{FileStem: "考评表"}, {FileStem: "考评表"}, {FileStem: "考评表-2"}, {FileStem: "考评表"}, {FileStem: "考评表-2"}}
	for _, names := range [][]string{uniqueFileStems(documents), uniqueSheetNames(documents)} {
		seen := make(map[string]bool)
		for _, name := range names {
			if seen[strings.ToLower(name)] {
				t.Errorf("duplicate title-based name %q in %v", name, names)
			}
			seen[strings.ToLower(name)] = true
		}
	}
	longTitle := strings.Repeat("考评表", 15)
	documents = []exportDocument{{FileStem: longTitle}, {FileStem: longTitle + "-2"}, {FileStem: "Review"}, {FileStem: "review"}, {FileStem: "Review-2"}}
	seen := make(map[string]bool)
	for _, name := range uniqueSheetNames(documents) {
		if seen[strings.ToLower(name)] || len([]rune(name)) > 31 {
			t.Errorf("invalid/duplicate worksheet name %q", name)
		}
		seen[strings.ToLower(name)] = true
	}
}

func TestFormExportBatchesKeepEveryCollidingTitle(t *testing.T) {
	for _, format := range []ExportFormat{ExportFormatPDF, ExportFormatDOCX, ExportFormatXLSX} {
		t.Run(string(format), func(t *testing.T) {
			runtime, ids := contentExportFixture(2)
			third := *runtime.details[ids[0]]
			third.Instance.ID = "third-record"
			third.Instance.DefinitionName = "历史考评表-2"
			runtime.list.List = append(runtime.list.List, third.Instance)
			runtime.details[third.Instance.ID] = &third
			ids = append(ids, third.Instance.ID)
			result, err := NewServiceWithAccess(runtime, summaryAccessStub{}).Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, contentRequest(t, ids, format, "form"))
			if err != nil {
				t.Fatal(err)
			}
			entries := readZipEntries(t, result.Body)
			if format == ExportFormatXLSX {
				for _, name := range []string{"张三-历史考评表", "张三-历史考评表-2", "张三-历史考评表-2-2"} {
					if strings.Count(entries["xl/workbook.xml"], `name="`+name+`"`) != 1 {
						t.Errorf("missing unique sheet %q", name)
					}
				}
				if !strings.Contains(entries["xl/worksheets/sheet3.xml"], "历史考评表-2") {
					t.Error("third record was lost")
				}
			} else {
				if len(entries) != 3 {
					t.Fatalf("batch has %d unique entries, want 3", len(entries))
				}
				for _, name := range []string{"张三-历史考评表", "张三-历史考评表-2", "张三-历史考评表-2-2"} {
					if len(entries[name+"."+string(format)]) == 0 {
						t.Errorf("missing unique file %q", name)
					}
				}
			}
		})
	}
}

func contentRequest(t *testing.T, ids []string, format ExportFormat, content string) ExportRequest {
	t.Helper()
	return ExportRequest{InstanceIDs: ids, Format: format, Content: ExportContent(content)}
}

func TestFormExportPreservesNestedAndDetailOnlyGroups(t *testing.T) {
	for _, format := range []ExportFormat{ExportFormatPDF, ExportFormatDOCX, ExportFormatXLSX} {
		t.Run(string(format), func(t *testing.T) {
			runtime, ids := contentExportFixture(1)
			detail := runtime.details[ids[0]]
			detail.Form = []workflowcore.FormField{{Key: "outer", Label: "考评外层分组", Type: workflowcore.FormFieldTypeGroup, Fields: []workflowcore.FormField{
				{Key: "inner", Label: "考评内层分组", Type: workflowcore.FormFieldTypeGroup, Fields: []workflowcore.FormField{detail.Form[2]}},
				{Key: "hrbp_group", Label: "HRBP分组", Type: workflowcore.FormFieldTypeGroup, Fields: []workflowcore.FormField{detail.Form[1].Fields[1]}},
			}}}
			result, err := NewServiceWithAccess(runtime, summaryAccessStub{}).Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, contentRequest(t, ids, format, "form"))
			if err != nil {
				t.Fatal(err)
			}
			body := exportedDocuments(t, result, format, 1, "form")[0]
			text := exportedText(t, body, format)
			for _, want := range []string{"考评外层分组", "考评内层分组", "项目明细", "历史项目", "HRBP分组", "HRBP意见", "建议晋升"} {
				if strings.Count(text, want) != 1 {
					t.Errorf("nested form should preserve %q once: %s", want, text)
				}
			}
			if strings.Contains(text, "内容内容") {
				t.Error("heading-only groups must not gain empty content tables")
			}
		})
	}
}

func contentExportFixture(count int) (*summaryRuntimeStub, []string) {
	runtime := &summaryRuntimeStub{list: &workflowapp.InstanceList{}, details: make(map[string]*workflowapp.InstanceDetail)}
	ids := make([]string, 0, count)
	for index := 0; index < count; index++ {
		id := fmt.Sprintf("INSTANCE-SECRET-%d", index)
		version := []string{"甲", "乙"}[index]
		instance := workflowapp.InstanceSummary{
			ID: id, DefinitionID: uint(index + 1), DefinitionVersion: index + 2,
			DefinitionName: "历史考评表", DefinitionKey: "DEFINITION-SECRET", BusinessKey: fmt.Sprintf("BUSINESS-SECRET-%d", index),
			StarterID: "STARTER-ID", StarterName: "张三", OperatorID: "OPERATOR-ID", OperatorName: "OPERATOR-SECRET",
			Status: "completed", StartTime: 1793318400000, EndTime: 1793322000000,
		}
		runtime.list.List = append(runtime.list.List, instance)
		runtime.details[id] = &workflowapp.InstanceDetail{
			Instance: instance,
			Form: []workflowcore.FormField{
				{Key: "historic", Label: "历史字段" + version, Type: workflowcore.FormFieldTypeText},
				{Key: "review", Label: "考评分组", Type: workflowcore.FormFieldTypeGroup, Fields: []workflowcore.FormField{
					{Key: "score", Label: "评分", Type: workflowcore.FormFieldTypeNumber, Span: 8},
					{Key: "hrbp", Label: "HRBP意见", Type: workflowcore.FormFieldTypeTextarea, Span: 8},
					{Key: "hrbp_user", Label: "HRBP负责人", Type: workflowcore.FormFieldTypeUser, Span: 8},
				}},
				{Key: "items", Label: "项目明细", Type: workflowcore.FormFieldTypeDetailList, Columns: []workflowcore.FormField{
					{Key: "name", Label: "项目名称", Type: workflowcore.FormFieldTypeText},
					{Key: "score", Label: "明细评分", Type: workflowcore.FormFieldTypeNumber},
				}},
				{Key: "total", Label: "计算总分", Type: workflowcore.FormFieldTypeCalculation, Span: 12},
				{Key: "status", Label: "流程状态", Type: workflowcore.FormFieldTypeText, Span: 12},
			},
			FormData: map[string]interface{}{
				"historic": "历史值" + version, "score": 87, "hrbp": "建议晋升", "hrbp_user": "user-42",
				"items": []map[string]interface{}{{"name": "历史项目", "score": 91}}, "total": 178, "status": "表单状态保留",
			},
			UserNames: map[string]string{"user-42": "王老师"},
			Tasks:     []workflowapp.TaskSummary{{NodeName: "TASK-SECRET", HandledByName: "APPROVER-SECRET", Comment: "COMMENT-SECRET", Action: "approve", HandledAt: 1793322000000}},
		}
		ids = append(ids, id)
	}
	return runtime, ids
}

func exportedDocuments(t *testing.T, result *ExportResult, format ExportFormat, count int, content string) [][]byte {
	t.Helper()
	bodies := make([][]byte, 0, count)
	if format == ExportFormatXLSX {
		if result.ContentType != exportContentType(format) || !strings.HasSuffix(result.Filename, ".xlsx") {
			t.Fatalf("expected XLSX workbook, got %s %s", result.ContentType, result.Filename)
		}
		entries := readZipEntries(t, result.Body)
		for index := 0; index < count; index++ {
			bodies = append(bodies, []byte(entries[fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)]))
			if content == "form" {
				name := "张三-历史考评表"
				if index > 0 {
					name += fmt.Sprintf("-%d", index+1)
				}
				if !strings.Contains(entries["xl/workbook.xml"], `name="`+name+`"`) || strings.Contains(entries["xl/workbook.xml"], "SECRET") {
					t.Errorf("form sheet names must use title and ordinal: %s", entries["xl/workbook.xml"])
				}
			}
		}
	} else if count == 1 {
		if result.ContentType != exportContentType(format) {
			t.Fatalf("single export type: %s", result.ContentType)
		}
		if content == "form" && result.Filename != "张三-历史考评表."+string(format) {
			t.Errorf("form filename = %s", result.Filename)
		}
		bodies = append(bodies, result.Body)
	} else {
		if result.ContentType != "application/zip" || !strings.HasSuffix(result.Filename, ".zip") {
			t.Fatalf("batch export must be ZIP: %s %s", result.ContentType, result.Filename)
		}
		entries := readZipEntries(t, result.Body)
		if len(entries) != count {
			t.Fatalf("batch has %d entries, want %d", len(entries), count)
		}
		for index := 0; index < count; index++ {
			stem := fmt.Sprintf("BUSINESS-SECRET-%d", index)
			if content == "form" {
				stem = "张三-历史考评表"
				if index > 0 {
					stem += fmt.Sprintf("-%d", index+1)
				}
			}
			body, exists := entries[stem+"."+string(format)]
			if !exists {
				t.Fatalf("missing batch document %s.%s", stem, format)
			}
			bodies = append(bodies, []byte(body))
		}
	}
	if content == "form" && (strings.Contains(result.Filename, "SECRET") || !strings.HasPrefix(result.Filename, "张三-历史考评表")) {
		t.Errorf("form container filename should use form title: %s", result.Filename)
	}
	return bodies
}

func exportedText(t *testing.T, body []byte, format ExportFormat) string {
	t.Helper()
	if format == ExportFormatPDF {
		var text strings.Builder
		for _, match := range regexp.MustCompile(`(?s)BT\n/F([12]) .*?<([0-9A-F]+)> Tj\nET`).FindAllSubmatch(body, -1) {
			data, err := hex.DecodeString(string(match[2]))
			if string(match[1]) == "2" && err == nil {
				text.Write(data)
				continue
			}
			if err != nil || len(data)%2 != 0 {
				t.Fatalf("invalid PDF text %q", match[2])
			}
			codes := make([]uint16, len(data)/2)
			for index := range codes {
				codes[index] = uint16(data[index*2])<<8 | uint16(data[index*2+1])
			}
			text.WriteString(string(utf16.Decode(codes)))
		}
		return text.String()
	}
	if format == ExportFormatDOCX {
		body = []byte(readZipEntries(t, body)["word/document.xml"])
	}
	var text strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(string(body)))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if data, ok := token.(xml.CharData); ok {
			text.Write(data)
		}
	}
	return text.String()
}

func assertFormStartsAfterTitle(t *testing.T, body []byte, format ExportFormat) {
	t.Helper()
	switch format {
	case ExportFormatDOCX:
		xml := readZipEntries(t, body)["word/document.xml"]
		// The first table must be the first form heading, with no metadata table or spacer before it.
		prefix := strings.SplitN(strings.SplitN(xml, "</w:p>", 2)[1], "</w:tbl>", 2)[0]
		if !strings.HasPrefix(prefix, "<w:tbl>") || !strings.Contains(prefix, "表单内容") {
			t.Error("empty metadata table/gap precedes form")
		}
	case ExportFormatXLSX:
		row := regexp.MustCompile(`<row r="2"[^>]*>.*?</row>`).Find(body)
		if !strings.Contains(string(row), "表单内容") {
			t.Errorf("second row should start form, got %s", row)
		}
	case ExportFormatPDF:
		if !strings.Contains(string(body), "42.00 740.00 511.00 26.00 re") {
			t.Error("first form section must occupy the first band below title")
		}
	}
}
