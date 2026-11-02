package workflowsummary

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
)

func TestExportTitleAndSheetsIdentifyEachStarter(t *testing.T) {
	for _, content := range []ExportContent{ExportContentFull, ExportContentForm} {
		t.Run(string(content), func(t *testing.T) {
			var documents []exportDocument
			for index, name := range []string{"Foster", "Amy", "Foster"} {
				detail := &workflowapp.InstanceDetail{Instance: workflowapp.InstanceSummary{
					DefinitionName: "绩效考评单", StarterName: name, OperatorName: "代办管理员",
					BusinessKey: fmt.Sprintf("record-%d", index),
				}}
				document := buildExportDocument(detail, content)
				if document.Title != name+"-绩效考评单" {
					t.Fatalf("title = %q, want starter-prefixed title", document.Title)
				}
				if content == ExportContentFull && document.Metadata[0][1] != "绩效考评单" {
					t.Fatal("workflow metadata must retain the original workflow name")
				}
				documents = append(documents, document)
			}
			body, err := renderXLSX(documents)
			if err != nil {
				t.Fatal(err)
			}
			entries := readZipEntries(t, body)
			var workbook struct {
				Sheets []struct {
					Name string `xml:"name,attr"`
				} `xml:"sheets>sheet"`
			}
			if err := xml.Unmarshal([]byte(entries["xl/workbook.xml"]), &workbook); err != nil {
				t.Fatal(err)
			}
			for index, expected := range []string{"Foster-绩效考评单", "Amy-绩效考评单", "Foster-绩效考评单-2"} {
				if workbook.Sheets[index].Name != expected {
					t.Errorf("sheet %d = %q, want %q", index, workbook.Sheets[index].Name, expected)
				}
				sheet := entries[fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)]
				firstRow := strings.SplitN(sheet, "</row>", 2)[0]
				if !strings.Contains(firstRow, documents[index].Title) {
					t.Errorf("first row missing starter title: %s", firstRow)
				}
			}
			for _, format := range []ExportFormat{ExportFormatPDF, ExportFormatDOCX} {
				data, err := renderSingleDocument(documents[0], format)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(strings.TrimSpace(exportedText(t, data, format)), "Foster-绩效考评单") {
					t.Errorf("%s title does not identify starter", format)
				}
			}
		})
	}
}

func TestExportStarterTitleFallbackAndSheetConstraints(t *testing.T) {
	for _, test := range []struct{ name, resolved, expected string }{
		{"  张三  ", "李四", "张三-绩效考评单"},
		{" ", "李四", "李四-绩效考评单"},
		{"", "", "绩效考评单"},
		{"张/三[测试]", "", "张/三[测试]-绩效考评单"},
		{strings.Repeat("长名称", 15), "", strings.Repeat("长名称", 15) + "-绩效考评单"},
	} {
		document := buildExportDocument(&workflowapp.InstanceDetail{
			Instance:  workflowapp.InstanceSummary{DefinitionName: "绩效考评单", StarterName: test.name, StarterID: "internal-id"},
			UserNames: map[string]string{"internal-id": test.resolved},
		}, ExportContentForm)
		if document.Title != test.expected {
			t.Errorf("title = %q, want %q", document.Title, test.expected)
		}
		names := uniqueSheetNames([]exportDocument{document, document})
		if names[0] == names[1] {
			t.Error("duplicate sheet names")
		}
		for _, name := range names {
			if utf8.RuneCountInString(name) > 31 || strings.ContainsAny(name, `[]:*?/\`) || strings.Contains(name, "internal-id") {
				t.Errorf("invalid worksheet name %q", name)
			}
		}
	}
}
