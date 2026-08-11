package review

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"wecheckin/backend/internal/model"
)

type ExportResult struct {
	Filename    string
	ContentType string
	Body        []byte
}

func ExportReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters) (*ExportResult, error) {
	if filters.Detail {
		filters.SkipHistory = true
	}
	reviewList, err := listReviewsContext(ctx, user, filters, false)
	if err != nil {
		return nil, err
	}
	filenamePrefix := "dingtalk-h5-performance"
	var body []byte
	if filters.Detail {
		filenamePrefix = "dingtalk-h5-performance-detail"
		body, err = buildReviewDetailXLSX(reviewList.List)
	} else {
		rows := reviewSummaryExportRows(reviewList.List)
		body, err = buildPlainXLSX(rows)
	}
	if err != nil {
		return nil, err
	}
	return &ExportResult{
		Filename:    fmt.Sprintf("%s-%s.xlsx", filenamePrefix, time.Now().Format("2006-01-02")),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Body:        body,
	}, nil
}

func reviewSummaryExportRows(reviews []ReviewDTO) [][]string {
	rows := [][]string{{"考评月份", "目标月份", "员工账号", "部门", "直属上级", "HRBP", "上级分档", "HRBP分档", "员工总结", "上级评价", "HRBP评价"}}
	for _, review := range reviews {
		rows = append(rows, []string{
			review.Period,
			review.NextPeriod,
			review.EmployeeID,
			review.Department,
			review.ManagerID,
			review.HRBPID,
			review.ManagerGrade,
			review.HRBPGrade,
			review.SelfSummary,
			review.ManagerComment,
			review.HRBPComment,
		})
	}
	return rows
}

func reviewExportFirstText(values ...string) string {
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text != "" {
			return text
		}
	}
	return ""
}

func reviewExportNumber(value float64) string {
	text := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
	if text == "-0" {
		return "0"
	}
	return text
}

func reviewExportValue(value interface{}) string {
	if value == nil {
		return ""
	}
	text := strings.TrimSpace(toString(value))
	if text == "null" {
		return ""
	}
	return text
}

func reviewExportComment(value string) string {
	return reviewExportValue(value)
}

func buildPlainXLSX(rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`,
		"xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="HRBP汇总" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`,
		"xl/worksheets/sheet1.xml": worksheetXML(rows),
	}
	for name, content := range files {
		if err := writeZipTextFile(archive, name, content); err != nil {
			_ = archive.Close()
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeZipTextFile(archive *zip.Writer, name string, content string) error {
	writer, err := archive.Create(name)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(content))
	return err
}

func worksheetXML(rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for rowIndex, row := range rows {
		rowNumber := rowIndex + 1
		b.WriteString(fmt.Sprintf(`<row r="%d">`, rowNumber))
		for colIndex, cell := range row {
			cellRef := fmt.Sprintf("%s%d", spreadsheetColumnName(colIndex+1), rowNumber)
			b.WriteString(`<c r="`)
			b.WriteString(cellRef)
			b.WriteString(`" t="inlineStr"><is><t>`)
			b.WriteString(escapeXMLText(normalizeSpreadsheetCell(cell)))
			b.WriteString(`</t></is></c>`)
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.String()
}

func spreadsheetColumnName(index int) string {
	if index <= 0 {
		return ""
	}
	var name []byte
	for index > 0 {
		index--
		name = append([]byte{byte('A' + index%26)}, name...)
		index /= 26
	}
	return string(name)
}

func escapeXMLText(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}

func normalizeSpreadsheetCell(value string) string {
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.TrimSpace(value)
}
