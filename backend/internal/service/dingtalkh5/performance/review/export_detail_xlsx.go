package review

import (
	"archive/zip"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	xlsxStyleDefault = iota
	xlsxStyleTitle
	xlsxStyleHeader
	xlsxStyleBody
	xlsxStyleSection
	xlsxStyleSubtotal
	xlsxStyleSignature
	xlsxStyleNote
	xlsxStyleBlueGrade
	xlsxStyleGreenGrade
	xlsxStyleYellowGrade
	xlsxStyleRedGrade
)

type xlsxWorkbookSheet struct {
	Name      string
	Rows      [][]xlsxCell
	Merges    []string
	ColWidths []float64
}

type xlsxCell struct {
	Text  string
	Style int
}

func buildReviewDetailXLSX(reviews []ReviewDTO) ([]byte, error) {
	return buildWorkbookXLSX(reviewDetailWorkbookSheets(reviews))
}

func reviewDetailWorkbookSheets(reviews []ReviewDTO) []xlsxWorkbookSheet {
	if len(reviews) == 0 {
		return []xlsxWorkbookSheet{
			buildReviewPerformanceDetailSheet(ReviewDTO{}, "绩效考评表"),
			buildGradeCoefficientSheet("绩效分档与绩效工资系数"),
		}
	}
	if len(reviews) == 1 {
		review := reviews[0]
		return []xlsxWorkbookSheet{
			buildReviewPerformanceDetailSheet(review, "绩效考评表"),
			buildGradeCoefficientSheet("绩效分档与绩效工资系数"),
		}
	}

	used := map[string]bool{}
	sheets := make([]xlsxWorkbookSheet, 0, len(reviews)+1)
	for _, review := range reviews {
		base := reviewExportSheetBaseName(review)
		sheetName := uniqueReviewExportSheetName(base+"-绩效详情", used)
		sheets = append(sheets, buildReviewPerformanceDetailSheet(review, sheetName))
	}
	sheets = append(sheets, buildGradeCoefficientSheet(uniqueReviewExportSheetName("绩效分档与绩效工资系数", used)))
	return sheets
}

func buildReviewPerformanceDetailSheet(review ReviewDTO, name string) xlsxWorkbookSheet {
	targetSheet := buildTargetPerformanceSheet(review, name)
	valueSheet := buildValuePerformanceSheet(review, name)
	combined := xlsxWorkbookSheet{
		Name:      name,
		Rows:      make([][]xlsxCell, 0, len(targetSheet.Rows)+len(valueSheet.Rows)+1),
		Merges:    make([]string, 0, len(targetSheet.Merges)+len(valueSheet.Merges)),
		ColWidths: combineXLSXColWidths(targetSheet.ColWidths, valueSheet.ColWidths),
	}
	combined.Rows = append(combined.Rows, targetSheet.Rows...)
	combined.Merges = append(combined.Merges, targetSheet.Merges...)
	if len(valueSheet.Rows) > 0 {
		combined.Rows = append(combined.Rows, emptyXLSXRow(len(combined.ColWidths), xlsxStyleDefault))
		offset := len(combined.Rows)
		combined.Rows = append(combined.Rows, valueSheet.Rows...)
		for _, mergeRef := range valueSheet.Merges {
			combined.Merges = append(combined.Merges, shiftXLSXMergeRows(mergeRef, offset))
		}
	}
	return combined
}

func buildTargetPerformanceSheet(review ReviewDTO, name string) xlsxWorkbookSheet {
	year, month := reviewExportYearMonth(review.Period)
	builder := xlsxSheetBuilder{
		name:      name,
		colWidths: []float64{3, 54, 13, 17, 19, 72},
	}

	builder.addSimpleRow(xlsxStyleDefault, "", "", "", "", "", "")
	titleRow := builder.addRow(
		xlsxText("", xlsxStyleDefault),
		xlsxText(fmt.Sprintf("%s年   %s  月 绩效目标", year, month), xlsxStyleTitle),
		xlsxText("", xlsxStyleTitle),
		xlsxText("", xlsxStyleTitle),
		xlsxText("", xlsxStyleTitle),
		xlsxText("", xlsxStyleTitle),
	)
	builder.merge(titleRow, "B", "F")
	builder.addSimpleRow(xlsxStyleHeader, "", "绩效目标", "考评权重", "目标完成百分比", "绩效目标得分", "绩效目标达成结果自评")

	objectives := review.Objectives
	if len(objectives) == 0 {
		objectives = []Objective{{}}
	}
	totalWeight := 0.0
	for _, objective := range objectives {
		totalWeight += objective.Weight
		builder.addSimpleRow(
			xlsxStyleBody,
			"",
			objective.Target,
			reviewExportPercent(objective.Weight),
			reviewExportPercentValue(objective.Completion),
			reviewExportObjectiveScore(objective.Weight, objective.Completion),
			objective.Result,
		)
	}
	builder.addSimpleRow(xlsxStyleBody, "", "", "", "", "", "")
	builder.addSimpleRow(xlsxStyleSubtotal, "", "合计", reviewExportPercent(totalWeight), "", "", "")

	summaryLabelRow := builder.addSimpleRow(xlsxStyleSection, "", "绩效承诺人的思考与总结", "", "", "", "")
	builder.merge(summaryLabelRow, "B", "F")
	summaryRow := builder.addSimpleRow(xlsxStyleNote, "", review.SelfSummary, "", "", "", "")
	builder.merge(summaryRow, "B", "F")

	reviewerLabelRow := builder.addSimpleRow(xlsxStyleSection, "", "绩效考评人的评价", "", "", "", "")
	builder.merge(reviewerLabelRow, "B", "F")
	managerRow := builder.addSimpleRow(xlsxStyleNote, "", reviewExportComment(review.ManagerComment), "", "", "", "")
	builder.merge(managerRow, "B", "F")
	hrbpLabelRow := builder.addSimpleRow(xlsxStyleSection, "", "由HRBP进行评价", "", "", "", "")
	builder.merge(hrbpLabelRow, "B", "F")
	hrbpRow := builder.addSimpleRow(xlsxStyleNote, "", reviewExportComment(review.HRBPComment), "", "", "", "")
	builder.merge(hrbpRow, "B", "F")

	builder.addSimpleRow(xlsxStyleHeader, "", "绩效分档：", effectiveGrade(review), "A:优秀  B:良好  C:及格  D:较差  E:糟糕", "", "")
	signRow := builder.addSimpleRow(
		xlsxStyleSignature,
		"",
		"绩效承诺人："+reviewExportFirstText(review.EmployeeName, review.EmployeeID),
		"",
		"绩效考评人："+reviewExportFirstText(review.ManagerName, review.ManagerID, review.HRBPReviewerName, review.HRBPName, review.HRBPReviewerID, review.HRBPID),
		"",
		"",
	)
	builder.merge(signRow, "B", "C")
	builder.merge(signRow, "D", "F")

	if len(review.NextObjectives) > 0 {
		builder.addSimpleRow(xlsxStyleBody, "", "", "", "", "", "")
		nextLabelRow := builder.addSimpleRow(xlsxStyleSection, "", "下月目标", "", "", "", "")
		builder.merge(nextLabelRow, "B", "F")
		builder.addSimpleRow(xlsxStyleHeader, "", "下月绩效目标", "考评权重", "", "", "")
		for _, objective := range review.NextObjectives {
			builder.addSimpleRow(xlsxStyleBody, "", objective.Target, reviewExportPercent(objective.Weight), "", "", "")
		}
	}

	return builder.sheet()
}

func buildGradeCoefficientSheet(name string) xlsxWorkbookSheet {
	rows := [][]xlsxCell{
		xlsxStyledRow(xlsxStyleDefault, "", "", "", ""),
		xlsxStyledRow(xlsxStyleTitle, "", "绩效", "绩效分档", "绩效工资系数"),
		xlsxStyledRow(xlsxStyleBlueGrade, "", "优秀", "A+", "1.5"),
		xlsxStyledRow(xlsxStyleBlueGrade, "", "", "A-", "1.4"),
		xlsxStyledRow(xlsxStyleGreenGrade, "", "良好", "B+", "1.3"),
		xlsxStyledRow(xlsxStyleGreenGrade, "", "", "B-", "1.2"),
		xlsxStyledRow(xlsxStyleBody, "", "及格", "C+", "1.1"),
		xlsxStyledRow(xlsxStyleBody, "", "", "C", "1.0"),
		xlsxStyledRow(xlsxStyleBody, "", "", "C-", "0.9"),
		xlsxStyledRow(xlsxStyleYellowGrade, "", "较差", "D+", "0.8"),
		xlsxStyledRow(xlsxStyleYellowGrade, "", "", "D-", "0.7"),
		xlsxStyledRow(xlsxStyleRedGrade, "", "糟糕", "E+", "0.6"),
		xlsxStyledRow(xlsxStyleRedGrade, "", "", "E-", "0.5"),
		xlsxStyledRow(xlsxStyleDefault, "", "", "", ""),
		xlsxStyledRow(xlsxStyleNote, "", "备注：\n绩效工资 = 工资总额 * 30% * 绩效工资系数", "", ""),
	}
	return xlsxWorkbookSheet{
		Name:      name,
		Rows:      rows,
		Merges:    []string{"B3:B4", "B5:B6", "B7:B9", "B10:B11", "B12:B13", "B15:D15"},
		ColWidths: []float64{3, 22, 22, 22},
	}
}

func buildValuePerformanceSheet(review ReviewDTO, name string) xlsxWorkbookSheet {
	year, month := reviewExportYearMonth(review.Period)
	builder := xlsxSheetBuilder{
		name:      name,
		colWidths: []float64{12, 16, 38, 12, 12, 12, 12, 10, 86},
	}
	titleRow := builder.addSimpleRow(xlsxStyleTitle, fmt.Sprintf("%s年     %s   月 价值观绩效考评", year, month), "", "", "", "", "", "", "", "")
	builder.merge(titleRow, "A", "I")
	builder.addSimpleRow(xlsxStyleHeader, "", "价值观", "价值观定义", "个人自评", "上级评价", "HR评价", "标度", "分值", "行为定义")

	values := review.Values
	if len(values) == 0 {
		values = []ValueScore{{}}
	}
	valueStartRow := builder.nextRowNumber()
	for valueIndex, value := range values {
		rubric := reviewExportValueRubric(value.Rubric)
		groupStartRow := builder.nextRowNumber()
		for rubricIndex, item := range rubric {
			contentLabel := ""
			valueName := ""
			valueDefinition := ""
			selfScore := ""
			managerScore := ""
			hrScore := ""
			if valueIndex == 0 && rubricIndex == 0 {
				contentLabel = "考核内容"
			}
			if rubricIndex == 0 {
				valueName = reviewExportFirstText(value.Name, value.ID)
				valueDefinition = value.Definition
				selfScore = reviewExportValue(value.Self)
				managerScore = reviewExportValue(value.Manager)
				hrScore = reviewExportHRValue(value)
			}
			builder.addSimpleRow(xlsxStyleBody, contentLabel, valueName, valueDefinition, selfScore, managerScore, hrScore, item.Label, reviewExportNumber(item.Score), item.Description)
		}
		groupEndRow := builder.nextRowNumber() - 1
		if groupEndRow > groupStartRow {
			builder.mergeRange("B", groupStartRow, "B", groupEndRow)
			builder.mergeRange("C", groupStartRow, "C", groupEndRow)
			builder.mergeRange("D", groupStartRow, "D", groupEndRow)
			builder.mergeRange("E", groupStartRow, "E", groupEndRow)
			builder.mergeRange("F", groupStartRow, "F", groupEndRow)
		}
	}
	valueEndRow := builder.nextRowNumber() - 1
	if valueEndRow >= valueStartRow {
		builder.mergeRange("A", valueStartRow, "A", valueEndRow)
	}

	noteRow := builder.addSimpleRow(
		xlsxStyleNote,
		"备注：1. 直评、自评根据对应的定义标准评定分数，分数可上下浮动，可选取介于两个等级之间的任意分值；\n"+
			"      2. 若直评和自评分数相差较大，直评可暂时不评定分数，须部门经理和人事介入沟通具体评分事宜。",
		"", "", "", "", "", "", "", "",
	)
	builder.merge(noteRow, "A", "I")

	return builder.sheet()
}

func reviewExportHRValue(value ValueScore) string {
	return reviewExportFirstText(reviewExportValue(value.HRBP), reviewExportValue(value.HR))
}

type xlsxSheetBuilder struct {
	name      string
	rows      [][]xlsxCell
	merges    []string
	colWidths []float64
}

func (builder *xlsxSheetBuilder) addSimpleRow(style int, values ...string) int {
	return builder.addRow(xlsxStyledRow(style, values...)...)
}

func (builder *xlsxSheetBuilder) addRow(cells ...xlsxCell) int {
	builder.rows = append(builder.rows, cells)
	return len(builder.rows)
}

func (builder *xlsxSheetBuilder) nextRowNumber() int {
	return len(builder.rows) + 1
}

func (builder *xlsxSheetBuilder) merge(row int, fromCol string, toCol string) {
	builder.mergeRange(fromCol, row, toCol, row)
}

func (builder *xlsxSheetBuilder) mergeRange(fromCol string, fromRow int, toCol string, toRow int) {
	builder.merges = append(builder.merges, fmt.Sprintf("%s%d:%s%d", fromCol, fromRow, toCol, toRow))
}

func (builder *xlsxSheetBuilder) sheet() xlsxWorkbookSheet {
	return xlsxWorkbookSheet{
		Name:      builder.name,
		Rows:      builder.rows,
		Merges:    builder.merges,
		ColWidths: builder.colWidths,
	}
}

func xlsxStyledRow(style int, values ...string) []xlsxCell {
	row := make([]xlsxCell, 0, len(values))
	for _, value := range values {
		row = append(row, xlsxText(value, style))
	}
	return row
}

func emptyXLSXRow(length int, style int) []xlsxCell {
	if length <= 0 {
		return nil
	}
	row := make([]xlsxCell, 0, length)
	for index := 0; index < length; index++ {
		row = append(row, xlsxText("", style))
	}
	return row
}

func combineXLSXColWidths(left []float64, right []float64) []float64 {
	length := len(left)
	if len(right) > length {
		length = len(right)
	}
	result := make([]float64, length)
	for index := 0; index < length; index++ {
		if index < len(left) {
			result[index] = left[index]
		}
		if index < len(right) && right[index] > result[index] {
			result[index] = right[index]
		}
	}
	return result
}

func shiftXLSXMergeRows(ref string, offset int) string {
	if offset == 0 || ref == "" {
		return ref
	}
	var b strings.Builder
	for index := 0; index < len(ref); {
		if ref[index] < '0' || ref[index] > '9' {
			b.WriteByte(ref[index])
			index++
			continue
		}
		start := index
		for index < len(ref) && ref[index] >= '0' && ref[index] <= '9' {
			index++
		}
		rowNumber, err := strconv.Atoi(ref[start:index])
		if err != nil {
			b.WriteString(ref[start:index])
			continue
		}
		b.WriteString(strconv.Itoa(rowNumber + offset))
	}
	return b.String()
}

func xlsxText(value string, style int) xlsxCell {
	return xlsxCell{Text: value, Style: style}
}

func buildWorkbookXLSX(sheets []xlsxWorkbookSheet) ([]byte, error) {
	if len(sheets) == 0 {
		sheets = []xlsxWorkbookSheet{{Name: "Sheet1"}}
	}
	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)
	files := reviewDetailWorkbookFiles(sheets)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := writeZipTextFile(archive, name, files[name]); err != nil {
			_ = archive.Close()
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func reviewDetailWorkbookFiles(sheets []xlsxWorkbookSheet) map[string]string {
	files := map[string]string{
		"[Content_Types].xml":        reviewDetailContentTypesXML(len(sheets)),
		"_rels/.rels":                reviewDetailRootRelsXML(),
		"xl/workbook.xml":            reviewDetailWorkbookXML(sheets),
		"xl/_rels/workbook.xml.rels": reviewDetailWorkbookRelsXML(len(sheets)),
		"xl/styles.xml":              reviewDetailStylesXML(),
	}
	for index, sheet := range sheets {
		files[fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)] = reviewDetailWorksheetXML(sheet)
	}
	return files
}

func reviewDetailContentTypesXML(sheetCount int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	b.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for index := 1; index <= sheetCount; index++ {
		b.WriteString(fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, index))
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func reviewDetailRootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`
}

func reviewDetailWorkbookXML(sheets []xlsxWorkbookSheet) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for index, sheet := range sheets {
		b.WriteString(fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, escapeXMLAttr(sheet.Name), index+1, index+1))
	}
	b.WriteString(`</sheets></workbook>`)
	return b.String()
}

func reviewDetailWorkbookRelsXML(sheetCount int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for index := 1; index <= sheetCount; index++ {
		b.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, index, index))
	}
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheetCount+1))
	b.WriteString(`</Relationships>`)
	return b.String()
}

func reviewDetailWorksheetXML(sheet xlsxWorkbookSheet) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	b.WriteString(`<sheetViews><sheetView showGridLines="0" workbookViewId="0"/></sheetViews>`)
	if len(sheet.ColWidths) > 0 {
		b.WriteString(`<cols>`)
		for index, width := range sheet.ColWidths {
			b.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%.2f" customWidth="1"/>`, index+1, index+1, width))
		}
		b.WriteString(`</cols>`)
	}
	b.WriteString(`<sheetData>`)
	for rowIndex, row := range sheet.Rows {
		rowNumber := rowIndex + 1
		b.WriteString(fmt.Sprintf(`<row r="%d">`, rowNumber))
		for colIndex, cell := range row {
			cellRef := fmt.Sprintf("%s%d", spreadsheetColumnName(colIndex+1), rowNumber)
			b.WriteString(`<c r="`)
			b.WriteString(cellRef)
			b.WriteString(`" s="`)
			b.WriteString(strconv.Itoa(cell.Style))
			b.WriteString(`" t="inlineStr"><is><t xml:space="preserve">`)
			b.WriteString(escapeXMLText(cell.Text))
			b.WriteString(`</t></is></c>`)
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData>`)
	if len(sheet.Merges) > 0 {
		b.WriteString(fmt.Sprintf(`<mergeCells count="%d">`, len(sheet.Merges)))
		for _, ref := range sheet.Merges {
			b.WriteString(`<mergeCell ref="`)
			b.WriteString(escapeXMLAttr(ref))
			b.WriteString(`"/>`)
		}
		b.WriteString(`</mergeCells>`)
	}
	b.WriteString(`</worksheet>`)
	return b.String()
}

func reviewDetailStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="4">
    <font><sz val="11"/><name val="Arial"/></font>
    <font><b/><sz val="11"/><name val="Arial"/></font>
    <font><b/><sz val="14"/><name val="Arial"/></font>
    <font><b/><color rgb="FFFFFFFF"/><sz val="11"/><name val="Arial"/></font>
  </fonts>
  <fills count="7">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFFFF00"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFD9D9D9"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF4472C4"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF70AD47"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFFC000"/><bgColor indexed="64"/></patternFill></fill>
  </fills>
  <borders count="2">
    <border><left/><right/><top/><bottom/><diagonal/></border>
    <border><left style="thin"><color rgb="FF404040"/></left><right style="thin"><color rgb="FF404040"/></right><top style="thin"><color rgb="FF404040"/></top><bottom style="thin"><color rgb="FF404040"/></bottom><diagonal/></border>
  </borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="12">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="2" fillId="2" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="0" borderId="1" xfId="0" applyAlignment="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1" applyBorder="1"><alignment horizontal="left" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="0" borderId="1" xfId="0" applyAlignment="1" applyBorder="1"><alignment horizontal="left" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="3" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1" applyBorder="1"><alignment horizontal="left" vertical="top" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="3" fillId="4" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="3" fillId="5" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="6" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="3" fillId="6" borderId="1" xfId="0" applyAlignment="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
  </cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
}

func reviewExportSheetBaseName(review ReviewDTO) string {
	return reviewExportFirstText(review.EmployeeName, review.EmployeeID, "考评单") + "-" + reviewExportFirstText(review.Period, "未设置月份")
}

func uniqueReviewExportSheetName(name string, used map[string]bool) string {
	base := sanitizeReviewExportSheetName(name)
	if base == "" {
		base = "Sheet"
	}
	if !used[base] {
		used[base] = true
		return base
	}
	for index := 2; ; index++ {
		suffix := fmt.Sprintf("-%d", index)
		candidate := trimReviewExportSheetName(base, 31-utf8.RuneCountInString(suffix)) + suffix
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}

func sanitizeReviewExportSheetName(name string) string {
	replacer := strings.NewReplacer("[", "-", "]", "-", ":", "-", "*", "-", "?", "-", "/", "-", "\\", "-")
	return trimReviewExportSheetName(strings.TrimSpace(replacer.Replace(name)), 31)
}

func trimReviewExportSheetName(name string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(name)
	if len(runes) <= limit {
		return name
	}
	return string(runes[:limit])
}

func reviewExportYearMonth(period string) (string, string) {
	parts := strings.Split(strings.TrimSpace(period), "-")
	if len(parts) >= 2 {
		month := strings.TrimLeft(parts[1], "0")
		if month == "" {
			month = parts[1]
		}
		return parts[0], month
	}
	return "", ""
}

func reviewExportPercent(value float64) string {
	if value == 0 {
		return ""
	}
	if value <= 1 {
		return reviewExportNumber(value*100) + "%"
	}
	return reviewExportNumber(value) + "%"
}

func reviewExportPercentValue(value interface{}) string {
	text := reviewExportValue(value)
	if text == "" {
		return ""
	}
	if strings.HasSuffix(text, "%") {
		return text
	}
	if number, err := strconv.ParseFloat(text, 64); err == nil {
		return reviewExportPercent(number)
	}
	return text
}

func reviewExportObjectiveScore(weight float64, completion interface{}) string {
	if weight == 0 {
		return ""
	}
	weightRatio := weight
	if weightRatio > 1 {
		weightRatio = weightRatio / 100
	}
	completionRatio := 1.0
	if text := reviewExportValue(completion); text != "" {
		normalized := strings.TrimSuffix(text, "%")
		if number, err := strconv.ParseFloat(normalized, 64); err == nil {
			completionRatio = number
			if completionRatio > 1 {
				completionRatio = completionRatio / 100
			}
		}
	}
	return reviewExportNumber(weightRatio * completionRatio)
}

func reviewExportValueRubric(items []ValueRubric) []ValueRubric {
	if len(items) > 0 {
		return items
	}
	return []ValueRubric{
		{Label: "卓越", Score: 50},
		{Label: "优秀", Score: 40},
		{Label: "良好", Score: 30},
		{Label: "及格", Score: 20},
		{Label: "较差", Score: 10},
	}
}

func escapeXMLAttr(value string) string {
	escaped := escapeXMLText(value)
	escaped = strings.ReplaceAll(escaped, `"`, "&quot;")
	return strings.ReplaceAll(escaped, "'", "&apos;")
}
