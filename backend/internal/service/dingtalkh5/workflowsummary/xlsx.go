package workflowsummary

import (
	"archive/zip"
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

func renderXLSX(documents []exportDocument) ([]byte, error) {
	if len(documents) == 0 {
		return nil, ErrExportInstancesEmpty
	}
	names := uniqueSheetNames(documents)
	files := map[string]string{
		"[Content_Types].xml": xlsxContentTypes(len(documents)),
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`,
		"xl/workbook.xml":            xlsxWorkbook(names),
		"xl/_rels/workbook.xml.rels": xlsxWorkbookRelationships(len(documents)),
		"xl/styles.xml":              xlsxStyles,
	}
	for index, document := range documents {
		files[fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)] = xlsxWorksheet(document)
	}
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writeZipText(archive, key, files[key]); err != nil {
			_ = archive.Close()
			return nil, exportError(ExportFormatXLSX, err)
		}
	}
	if err := archive.Close(); err != nil {
		return nil, exportError(ExportFormatXLSX, err)
	}
	return buffer.Bytes(), nil
}

func xlsxContentTypes(sheetCount int) string {
	var sheets strings.Builder
	for index := 1; index <= sheetCount; index++ {
		fmt.Fprintf(&sheets, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, index)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` + sheets.String() + `
</Types>`
}

func xlsxWorkbook(names []string) string {
	var sheets strings.Builder
	for index, name := range names {
		fmt.Fprintf(&sheets, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlText(name), index+1, index+1)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>` + sheets.String() + `</sheets></workbook>`
}

func xlsxWorkbookRelationships(sheetCount int) string {
	var relationships strings.Builder
	for index := 1; index <= sheetCount; index++ {
		fmt.Fprintf(&relationships, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, index, index)
	}
	fmt.Fprintf(&relationships, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheetCount+1)
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + relationships.String() + `</Relationships>`
}

const xlsxStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="4">
    <font><sz val="10.5"/><name val="Microsoft YaHei"/><color rgb="FF1F2329"/></font>
    <font><b/><sz val="18"/><name val="Microsoft YaHei"/><color rgb="FF1F2329"/></font>
    <font><b/><sz val="10.5"/><name val="Microsoft YaHei"/><color rgb="FF1F2329"/></font>
    <font><sz val="9.5"/><name val="Microsoft YaHei"/><color rgb="FF4E5969"/></font>
  </fonts>
  <fills count="4">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFE8F3F1"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFF5F7FA"/><bgColor indexed="64"/></patternFill></fill>
  </fills>
  <borders count="2"><border/><border><left style="thin"><color rgb="FFD9E1E8"/></left><right style="thin"><color rgb="FFD9E1E8"/></right><top style="thin"><color rgb="FFD9E1E8"/></top><bottom style="thin"><color rgb="FFD9E1E8"/></bottom></border></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="9">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"><alignment vertical="center"/></xf>
    <xf numFmtId="0" fontId="2" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment vertical="center"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"><alignment vertical="top" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="2" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="2" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="3" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"><alignment vertical="top" wrapText="1"/></xf>
  </cellXfs>
</styleSheet>`

const xlsxGridColumns = 24

type xlsxExportCell struct {
	column int
	style  int
	value  string
}

type xlsxExportRow struct {
	height float64
	cells  []xlsxExportCell
}

type xlsxMerge struct {
	row      int
	from, to int
}

type xlsxSheetLayout struct {
	rows   []xlsxExportRow
	merges []xlsxMerge
}

func xlsxWorksheet(document exportDocument) string {
	layout := buildXLSXSheetLayout(document)
	var body strings.Builder
	for rowIndex, row := range layout.rows {
		fmt.Fprintf(&body, `<row r="%d" ht="%.1f" customHeight="1">`, rowIndex+1, row.height)
		for _, cell := range row.cells {
			fmt.Fprintf(&body, `<c r="%s%d" t="inlineStr" s="%d"><is><t xml:space="preserve">%s</t></is></c>`, xlsxColumnName(cell.column), rowIndex+1, cell.style, xmlText(cell.value))
		}
		body.WriteString(`</row>`)
	}
	var mergeXML strings.Builder
	if len(layout.merges) > 0 {
		fmt.Fprintf(&mergeXML, `<mergeCells count="%d">`, len(layout.merges))
		for _, merge := range layout.merges {
			fmt.Fprintf(&mergeXML, `<mergeCell ref="%s%d:%s%d"/>`, xlsxColumnName(merge.from), merge.row, xlsxColumnName(merge.to), merge.row)
		}
		mergeXML.WriteString(`</mergeCells>`)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetViews><sheetView workbookViewId="0" showGridLines="0"/></sheetViews><sheetFormatPr defaultRowHeight="20"/><cols><col min="1" max="24" width="4.5" customWidth="1"/></cols><sheetData>` + body.String() + `</sheetData>` + mergeXML.String() + `<pageMargins left="0.35" right="0.35" top="0.5" bottom="0.5" header="0.2" footer="0.2"/><pageSetup fitToWidth="1" fitToHeight="0" orientation="portrait"/></worksheet>`
}

func buildXLSXSheetLayout(document exportDocument) xlsxSheetLayout {
	layout := xlsxSheetLayout{}
	addRow := func(height float64, cells ...xlsxExportCell) int {
		layout.rows = append(layout.rows, xlsxExportRow{height: height, cells: cells})
		return len(layout.rows)
	}
	merge := func(row, from, to int) {
		if to > from {
			layout.merges = append(layout.merges, xlsxMerge{row: row, from: from, to: to})
		}
	}

	titleRow := addRow(34, xlsxExportCell{column: 1, style: 1, value: document.Title})
	merge(titleRow, 1, xlsxGridColumns)
	for _, metadata := range document.Metadata {
		label, value := exportRowPair(metadata)
		row := addRow(24,
			xlsxExportCell{column: 1, style: 5, value: label},
			xlsxExportCell{column: 6, style: 6, value: value},
		)
		merge(row, 1, 5)
		merge(row, 6, xlsxGridColumns)
	}

	for _, section := range document.Sections {
		addRow(9, xlsxExportCell{column: 1, style: 0, value: ""})
		sectionRow := addRow(26, xlsxExportCell{column: 1, style: 2, value: section.Title})
		merge(sectionRow, 1, xlsxGridColumns)
		if len(section.FieldRows) > 0 {
			for _, fields := range section.FieldRows {
				labelCells := make([]xlsxExportCell, 0, len(fields))
				valueCells := make([]xlsxExportCell, 0, len(fields))
				column := 1
				valueHeight := 32.0
				for _, field := range fields {
					span := normalizedExportSpan(field.Span)
					to := column + span - 1
					labelCells = append(labelCells, xlsxExportCell{column: column, style: 7, value: field.Label})
					valueCells = append(valueCells, xlsxExportCell{column: column, style: 8, value: field.Value})
					merge(len(layout.rows)+1, column, to)
					merge(len(layout.rows)+2, column, to)
					if height := xlsxValueRowHeight(field.Value, span); height > valueHeight {
						valueHeight = height
					}
					column = to + 1
				}
				addRow(22, labelCells...)
				addRow(valueHeight, valueCells...)
			}
			continue
		}

		headers := section.Headers
		if len(headers) == 0 {
			headers = []string{"内容"}
		}
		ranges := xlsxTableColumnRanges(len(headers))
		headerCells := make([]xlsxExportCell, 0, len(headers))
		headerRow := len(layout.rows) + 1
		for index, header := range headers {
			headerCells = append(headerCells, xlsxExportCell{column: ranges[index][0], style: 4, value: header})
			merge(headerRow, ranges[index][0], ranges[index][1])
		}
		addRow(25, headerCells...)
		for _, values := range section.Rows {
			cells := make([]xlsxExportCell, 0, len(headers))
			height := 28.0
			rowNumber := len(layout.rows) + 1
			for index := range headers {
				value := ""
				if index < len(values) {
					value = values[index]
				}
				cells = append(cells, xlsxExportCell{column: ranges[index][0], style: 3, value: value})
				merge(rowNumber, ranges[index][0], ranges[index][1])
				if valueHeight := xlsxValueRowHeight(value, ranges[index][1]-ranges[index][0]+1); valueHeight > height {
					height = valueHeight
				}
			}
			addRow(height, cells...)
		}
	}
	return layout
}

func exportRowPair(values []string) (string, string) {
	label, value := "", ""
	if len(values) > 0 {
		label = values[0]
	}
	if len(values) > 1 {
		value = strings.Join(values[1:], " ")
	}
	return label, value
}

func normalizedExportSpan(span int) int {
	if span <= 0 || span > xlsxGridColumns {
		return xlsxGridColumns
	}
	return span
}

func xlsxTableColumnRanges(count int) [][2]int {
	if count <= 0 {
		count = 1
	}
	ranges := make([][2]int, 0, count)
	base := xlsxGridColumns / count
	remainder := xlsxGridColumns % count
	column := 1
	for index := 0; index < count; index++ {
		width := base
		if index < remainder {
			width++
		}
		if width < 1 {
			width = 1
		}
		to := column + width - 1
		ranges = append(ranges, [2]int{column, to})
		column = to + 1
	}
	return ranges
}

func xlsxValueRowHeight(value string, span int) float64 {
	span = normalizedExportSpan(span)
	charactersPerLine := span * 4
	if charactersPerLine < 8 {
		charactersPerLine = 8
	}
	lineCount := 0
	for _, line := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
		length := len([]rune(line))
		lines := (length + charactersPerLine - 1) / charactersPerLine
		if lines < 1 {
			lines = 1
		}
		lineCount += lines
	}
	height := float64(lineCount*16 + 8)
	if height < 32 {
		return 32
	}
	if height > 120 {
		return 120
	}
	return height
}

func xlsxColumnName(index int) string {
	name := ""
	for index > 0 {
		index--
		name = string(rune('A'+index%26)) + name
		index /= 26
	}
	return name
}

func uniqueSheetNames(documents []exportDocument) []string {
	seen := make(map[string]int, len(documents))
	result := make([]string, 0, len(documents))
	for _, document := range documents {
		name := strings.Map(func(r rune) rune {
			if strings.ContainsRune(`[]:*?/\\`, r) || r < 32 {
				return '_'
			}
			return r
		}, firstText(document.FileStem, document.Title, "流程"))
		name = truncateRunes(strings.Trim(name, "' "), 31)
		if name == "" {
			name = "流程"
		}
		base := name
		seen[base]++
		if seen[base] > 1 {
			suffix := fmt.Sprintf("-%d", seen[base])
			name = truncateRunes(base, 31-utf8.RuneCountInString(suffix)) + suffix
		}
		result = append(result, name)
	}
	return result
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
