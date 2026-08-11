package review

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReviewDetailExportWorkbookMatchesMonthlyReviewTemplate(t *testing.T) {
	body, err := buildReviewDetailXLSX([]ReviewDTO{sampleDetailExportReview()})
	if err != nil {
		t.Fatalf("build detail xlsx: %v", err)
	}

	entries := readXLSXEntries(t, body)
	workbook := entries["xl/workbook.xml"]
	for _, want := range []string{
		`name="绩效考评表"`,
		`name="绩效分档与绩效工资系数"`,
	} {
		if !strings.Contains(workbook, want) {
			t.Fatalf("detail workbook should contain sheet %q, got:\n%s", want, workbook)
		}
	}
	if got := strings.Count(workbook, "<sheet "); got != 2 {
		t.Fatalf("detail workbook should keep review detail and grade sheets, got %d sheets:\n%s", got, workbook)
	}
	assertOrderedText(t, workbook,
		`name="绩效考评表"`,
		`name="绩效分档与绩效工资系数"`,
	)

	flat := strings.Join([]string{
		entries["xl/worksheets/sheet1.xml"],
		entries["xl/worksheets/sheet2.xml"],
	}, "\n")
	for _, want := range []string{
		"2026年   7  月 绩效目标",
		"绩效目标",
		"考评权重",
		"目标完成百分比",
		"绩效目标得分",
		"绩效目标达成结果自评",
		"完成 H5 绩效详情导出",
		"绩效承诺人的思考与总结",
		"本月交付核心功能并完成复盘",
		"绩效考评人的评价",
		"上级评价内容",
		"HRBP评价内容",
		"绩效分档",
		"绩效工资系数",
		"2026年     7   月 价值观绩效考评",
		"价值观",
		"上级评价",
		"HR评价",
		"团结一心",
		"协作共赢",
	} {
		if !strings.Contains(flat, want) {
			t.Fatalf("detail xlsx should contain %q, got:\n%s", want, flat)
		}
	}
	assertOrderedText(t, entries["xl/worksheets/sheet1.xml"],
		"2026年   7  月 绩效目标",
		"绩效考评人的评价",
		"上级评价内容",
		"由HRBP进行评价",
		"HRBP评价内容",
		"2026年     7   月 价值观绩效考评",
	)
	valueSectionIndex := strings.Index(entries["xl/worksheets/sheet1.xml"], "2026年     7   月 价值观绩效考评")
	if valueSectionIndex == -1 {
		t.Fatalf("detail xlsx should contain value performance section, got:\n%s", entries["xl/worksheets/sheet1.xml"])
	}
	valueSection := entries["xl/worksheets/sheet1.xml"][valueSectionIndex:]
	assertOrderedText(t, valueSection,
		"2026年     7   月 价值观绩效考评",
		"个人自评",
		"上级评价",
		"HR评价",
		"标度",
		"分值",
		"行为定义",
	)
	if strings.Contains(entries["xl/worksheets/sheet1.xml"], "由直接上级主管进行评价") {
		t.Fatalf("manager comment row should not repeat legacy manager label, got:\n%s", entries["xl/worksheets/sheet1.xml"])
	}
	for _, unwanted := range []string{"部门直评", "流程记录", "提交给HRBP"} {
		if strings.Contains(flat, unwanted) {
			t.Fatalf("detail xlsx should not contain flow history %q, got:\n%s", unwanted, flat)
		}
	}
}

func TestReviewDetailExportKeepsTargetAndValueSheetsInOneWorkbook(t *testing.T) {
	first := sampleDetailExportReview()
	second := sampleDetailExportReview()
	second.ID = "foster-2026-06"
	second.EmployeeID = "foster"
	second.EmployeeName = "Foster"
	second.Period = "2026-06"
	second.NextPeriod = "2026-07"

	body, err := buildReviewDetailXLSX([]ReviewDTO{first, second})
	if err != nil {
		t.Fatalf("build detail xlsx: %v", err)
	}

	entries := readXLSXEntries(t, body)
	workbook := entries["xl/workbook.xml"]
	for _, want := range []string{
		`name="Lip-2026-07-绩效详情"`,
		`name="Foster-2026-06-绩效详情"`,
		`name="绩效分档与绩效工资系数"`,
	} {
		if !strings.Contains(workbook, want) {
			t.Fatalf("detail workbook should contain sheet %q, got:\n%s", want, workbook)
		}
	}
	if got := strings.Count(workbook, "<sheet "); got != 3 {
		t.Fatalf("detail workbook should keep one detail sheet per review plus grade sheet, got %d sheets:\n%s", got, workbook)
	}
	for sheetFile, wants := range map[string][]string{
		"xl/worksheets/sheet1.xml": {"2026年   7  月 绩效目标", "2026年     7   月 价值观绩效考评", "上级评价内容", "HRBP评价内容"},
		"xl/worksheets/sheet2.xml": {"2026年   6  月 绩效目标", "2026年     6   月 价值观绩效考评", "上级评价内容", "HRBP评价内容"},
	} {
		for _, want := range wants {
			if !strings.Contains(entries[sheetFile], want) {
				t.Fatalf("%s should contain %q, got:\n%s", sheetFile, want, entries[sheetFile])
			}
		}
	}
}

func sampleDetailExportReview() ReviewDTO {
	return ReviewDTO{
		ID:             "lip-2026-07",
		EmployeeID:     "lip",
		EmployeeName:   "Lip",
		ManagerID:      "cube",
		ManagerName:    "Cube",
		HRBPID:         "nick",
		HRBPName:       "Nick",
		Department:     "M/H业务 / 研发部 / Java开发一组",
		Period:         "2026-07",
		NextPeriod:     "2026-08",
		Status:         ReviewStatusCompleted,
		SelfSummary:    "本月交付核心功能并完成复盘",
		ManagerComment: "上级评价内容",
		ManagerGrade:   "B+",
		HRBPComment:    "HRBP评价内容",
		HRBPGrade:      "A-",
		FinalGrade:     "A-",
		FinalNote:      "归档备注",
		Objectives: []Objective{
			{ID: "obj-1", Target: "完成 H5 绩效详情导出", Weight: 60, Completion: 100, Result: "已上线"},
		},
		Values: []ValueScore{
			{
				ID:         "team",
				Name:       "团结一心",
				Definition: "协作共赢",
				Self:       45,
				Manager:    42,
				HRBP:       43,
				HR:         44,
				Rubric: []ValueRubric{
					{Label: "优秀", Score: 40, Description: "主动协作"},
				},
			},
		},
		NextObjectives: []NextObjective{
			{ID: "next-1", Target: "推进下月 OKR 目标", Weight: 100},
		},
		EmployeeConfirmResult:  "confirmed",
		EmployeeConfirmComment: "确认无异议",
		EmployeeConfirmedAt:    1785545348000,
		History: []HistoryDTO{
			{At: 1785545348000, By: "Nick", Action: "提交给HRBP"},
		},
	}
}

func readXLSXEntries(t *testing.T, body []byte) map[string]string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("read xlsx zip: %v", err)
	}
	entries := make(map[string]string)
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open xlsx entry %s: %v", file.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read xlsx entry %s: %v", file.Name, err)
		}
		entries[file.Name] = string(data)
	}
	return entries
}

func assertOrderedText(t *testing.T, haystack string, values ...string) {
	t.Helper()
	lastIndex := -1
	for _, value := range values {
		index := strings.Index(haystack, value)
		if index == -1 {
			t.Fatalf("expected %q to be present in:\n%s", value, haystack)
		}
		if index <= lastIndex {
			t.Fatalf("expected %q after previous text in order %v, got:\n%s", value, values, haystack)
		}
		lastIndex = index
	}
}
