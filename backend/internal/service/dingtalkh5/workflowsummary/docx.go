package workflowsummary

import (
	"archive/zip"
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func renderDOCX(document exportDocument) ([]byte, error) {
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`,
		"word/_rels/document.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`,
		"word/styles.xml":   docxStyles,
		"word/document.xml": docxDocument(document),
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
			return nil, exportError(ExportFormatDOCX, err)
		}
	}
	if err := archive.Close(); err != nil {
		return nil, exportError(ExportFormatDOCX, err)
	}
	return buffer.Bytes(), nil
}

const docxStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:pPr><w:spacing w:after="0" w:line="300" w:lineRule="auto"/></w:pPr><w:rPr><w:rFonts w:ascii="Microsoft YaHei" w:hAnsi="Microsoft YaHei" w:eastAsia="Microsoft YaHei"/><w:sz w:val="21"/><w:color w:val="1F2329"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="240"/></w:pPr><w:rPr><w:b/><w:sz w:val="34"/></w:rPr></w:style>
</w:styles>`

func docxDocument(document exportDocument) string {
	var body strings.Builder
	body.WriteString(docxParagraph(document.Title, "Title"))
	body.WriteString(docxMetadataTable(document.Metadata))
	for _, section := range document.Sections {
		body.WriteString(docxSpacer(100))
		body.WriteString(docxSectionHeading(section.Title))
		if len(section.FieldRows) > 0 {
			body.WriteString(docxFieldTable(section.FieldRows))
			continue
		}
		body.WriteString(docxTable(section.Headers, section.Rows))
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + body.String() + `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="850" w:right="850" w:bottom="850" w:left="850"/><w:cols w:space="425"/></w:sectPr></w:body></w:document>`
}

func docxParagraph(value, style string) string {
	properties := ""
	if style != "" {
		properties = `<w:pPr><w:pStyle w:val="` + xmlText(style) + `"/></w:pPr>`
	}
	return `<w:p>` + properties + docxRun(value, false) + `</w:p>`
}

func docxSpacer(height int) string {
	return `<w:p><w:pPr><w:spacing w:before="0" w:after="0" w:line="` + fmt.Sprint(height) + `" w:lineRule="exact"/></w:pPr></w:p>`
}

func docxMetadataTable(rows [][]string) string {
	const labelWidth = 1900
	const valueWidth = 8306
	var body strings.Builder
	body.WriteString(docxTableStart([]int{labelWidth, valueWidth}))
	for _, row := range rows {
		label, value := exportRowPair(row)
		body.WriteString(`<w:tr><w:trPr><w:trHeight w:val="420" w:hRule="atLeast"/></w:trPr>`)
		body.WriteString(docxCell(label, true, "F5F7FA", 1, labelWidth))
		body.WriteString(docxCell(value, false, "FFFFFF", 1, valueWidth))
		body.WriteString(`</w:tr>`)
	}
	body.WriteString(`</w:tbl>`)
	return body.String()
}

func docxSectionHeading(title string) string {
	const width = 10206
	return docxTableStart([]int{width}) + `<w:tr><w:trPr><w:trHeight w:val="460" w:hRule="atLeast"/></w:trPr>` + docxCell(title, true, "E8F3F1", 1, width) + `</w:tr></w:tbl>`
}

func docxFieldTable(rows [][]exportFieldCell) string {
	const gridWidth = 425
	widths := make([]int, 24)
	for index := range widths {
		widths[index] = gridWidth
	}
	var body strings.Builder
	body.WriteString(docxTableStart(widths))
	for _, fields := range rows {
		body.WriteString(`<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="360" w:hRule="atLeast"/></w:trPr>`)
		used := 0
		for _, field := range fields {
			span := normalizedExportSpan(field.Span)
			body.WriteString(docxCell(field.Label, false, "F5F7FA", span, span*gridWidth))
			used += span
		}
		if used < 24 {
			body.WriteString(docxCell("", false, "F5F7FA", 24-used, (24-used)*gridWidth))
		}
		body.WriteString(`</w:tr>`)

		body.WriteString(`<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="620" w:hRule="atLeast"/></w:trPr>`)
		used = 0
		for _, field := range fields {
			span := normalizedExportSpan(field.Span)
			body.WriteString(docxCell(field.Value, false, "FFFFFF", span, span*gridWidth))
			used += span
		}
		if used < 24 {
			body.WriteString(docxCell("", false, "FFFFFF", 24-used, (24-used)*gridWidth))
		}
		body.WriteString(`</w:tr>`)
	}
	body.WriteString(`</w:tbl>`)
	return body.String()
}

func docxTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		headers = []string{"内容"}
	}
	widths := evenDOCXColumnWidths(len(headers), 10206)
	var body strings.Builder
	body.WriteString(docxTableStart(widths))
	body.WriteString(docxTableRow(headers, widths, true, "F5F7FA"))
	for _, row := range rows {
		body.WriteString(docxTableRow(row, widths, false, "FFFFFF"))
	}
	body.WriteString(`</w:tbl>`)
	return body.String()
}

func docxTableStart(widths []int) string {
	var grid strings.Builder
	for _, width := range widths {
		grid.WriteString(`<w:gridCol w:w="` + fmt.Sprint(width) + `"/>`)
	}
	return `<w:tbl><w:tblPr><w:tblW w:w="10206" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblBorders><w:top w:val="single" w:sz="4" w:color="D9E1E8"/><w:left w:val="single" w:sz="4" w:color="D9E1E8"/><w:bottom w:val="single" w:sz="4" w:color="D9E1E8"/><w:right w:val="single" w:sz="4" w:color="D9E1E8"/><w:insideH w:val="single" w:sz="4" w:color="D9E1E8"/><w:insideV w:val="single" w:sz="4" w:color="D9E1E8"/></w:tblBorders><w:tblCellMar><w:top w:w="90" w:type="dxa"/><w:left w:w="120" w:type="dxa"/><w:bottom w:w="90" w:type="dxa"/><w:right w:w="120" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>` + grid.String() + `</w:tblGrid>`
}

func docxTableRow(values []string, widths []int, bold bool, fill string) string {
	var row strings.Builder
	row.WriteString(`<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="440" w:hRule="atLeast"/></w:trPr>`)
	for index, width := range widths {
		value := ""
		if index < len(values) {
			value = values[index]
		}
		row.WriteString(docxCell(value, bold, fill, 1, width))
	}
	row.WriteString(`</w:tr>`)
	return row.String()
}

func docxCell(value string, bold bool, fill string, span, width int) string {
	properties := `<w:tcW w:w="` + fmt.Sprint(width) + `" w:type="dxa"/>`
	if span > 1 {
		properties += `<w:gridSpan w:val="` + fmt.Sprint(span) + `"/>`
	}
	if fill != "" {
		properties += `<w:shd w:fill="` + fill + `"/>`
	}
	properties += `<w:vAlign w:val="center"/>`
	return `<w:tc><w:tcPr>` + properties + `</w:tcPr><w:p><w:pPr><w:spacing w:before="0" w:after="0"/></w:pPr>` + docxRun(value, bold) + `</w:p></w:tc>`
}

func evenDOCXColumnWidths(count, total int) []int {
	if count <= 0 {
		count = 1
	}
	widths := make([]int, count)
	base := total / count
	remainder := total % count
	for index := range widths {
		widths[index] = base
		if index < remainder {
			widths[index]++
		}
	}
	return widths
}

func docxRun(value string, bold bool) string {
	properties := `<w:rPr><w:rFonts w:ascii="Microsoft YaHei" w:hAnsi="Microsoft YaHei" w:eastAsia="Microsoft YaHei"/>`
	if bold {
		properties += `<w:b/>`
	}
	properties += `</w:rPr>`
	parts := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	var text strings.Builder
	for index, part := range parts {
		if index > 0 {
			text.WriteString(`<w:br/>`)
		}
		text.WriteString(`<w:t xml:space="preserve">` + xmlText(part) + `</w:t>`)
	}
	return `<w:r>` + properties + text.String() + `</w:r>`
}
