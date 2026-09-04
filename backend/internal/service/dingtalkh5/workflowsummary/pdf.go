package workflowsummary

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"unicode/utf16"
)

const (
	pdfPageWidth    = 595.0
	pdfPageHeight   = 842.0
	pdfPageMargin   = 42.0
	pdfContentWidth = pdfPageWidth - pdfPageMargin*2
	pdfPageBottom   = 800.0
)

const (
	pdfColorText        = "0.12 0.14 0.16"
	pdfColorMuted       = "0.31 0.35 0.41"
	pdfColorPrimary     = "0.06 0.46 0.43"
	pdfColorSectionFill = "0.91 0.96 0.95"
	pdfColorLabelFill   = "0.96 0.97 0.98"
	pdfColorWhite       = "1 1 1"
	pdfColorBorder      = "0.85 0.88 0.91"
)

type pdfRenderer struct {
	title   string
	pages   []*strings.Builder
	current *strings.Builder
	y       float64
}

type pdfCell struct {
	x         float64
	width     float64
	lines     []string
	fill      string
	textColor string
	fontSize  float64
}

type pdfTextRun struct {
	value string
	ascii bool
}

func renderPDF(document exportDocument) ([]byte, error) {
	renderer := newPDFRenderer(document.Title)
	renderer.drawText(pdfPageMargin, 43, 17, pdfColorText, document.Title)
	renderer.y = 76
	renderer.drawSectionBand("流程信息")
	for _, metadata := range document.Metadata {
		label, value := exportRowPair(metadata)
		renderer.drawRow([]pdfCell{
			newPDFCell(pdfPageMargin, 100, label, pdfColorLabelFill, pdfColorMuted, 9),
			newPDFCell(pdfPageMargin+100, pdfContentWidth-100, value, pdfColorWhite, pdfColorText, 10),
		}, 24, 13)
	}

	for _, section := range document.Sections {
		renderer.y += 12
		renderer.ensureSpace(54)
		renderer.drawSectionBand(section.Title)
		if len(section.FieldRows) > 0 {
			renderer.drawFieldRows(section.FieldRows)
			continue
		}
		renderer.drawTable(section.Headers, section.Rows)
	}
	return buildPDFFile(renderer.finish()), nil
}

func newPDFRenderer(title string) *pdfRenderer {
	renderer := &pdfRenderer{title: firstText(title, "流程记录")}
	renderer.newPage(false)
	return renderer
}

func (renderer *pdfRenderer) newPage(continued bool) {
	if renderer.current != nil {
		renderer.current.WriteString("Q\n")
	}
	page := &strings.Builder{}
	page.WriteString("q\n")
	renderer.pages = append(renderer.pages, page)
	renderer.current = page
	renderer.y = pdfPageMargin
	if continued {
		renderer.drawText(pdfPageMargin, renderer.y, 11, pdfColorMuted, renderer.title+"（续）")
		renderer.y += 22
		renderer.drawHorizontalLine(renderer.y, pdfColorBorder)
		renderer.y += 10
	}
}

func (renderer *pdfRenderer) finish() [][]byte {
	if renderer.current != nil {
		renderer.current.WriteString("Q\n")
	}
	pages := make([][]byte, 0, len(renderer.pages))
	for _, page := range renderer.pages {
		pages = append(pages, []byte(page.String()))
	}
	return pages
}

func (renderer *pdfRenderer) ensureSpace(height float64) {
	if renderer.y+height > pdfPageBottom {
		renderer.newPage(true)
	}
}

func (renderer *pdfRenderer) drawSectionBand(title string) {
	renderer.ensureSpace(26)
	renderer.drawRectangle(pdfPageMargin, renderer.y, pdfContentWidth, 26, pdfColorSectionFill, true)
	renderer.drawText(pdfPageMargin+9, renderer.y+6, 11, pdfColorPrimary, title)
	renderer.y += 26
}

func (renderer *pdfRenderer) drawFieldRows(rows [][]exportFieldCell) {
	for _, fields := range rows {
		if len(fields) == 0 {
			continue
		}
		x := pdfPageMargin
		labelCells := make([]pdfCell, 0, len(fields))
		valueCells := make([]pdfCell, 0, len(fields))
		for _, field := range fields {
			span := normalizedExportSpan(field.Span)
			width := pdfContentWidth * float64(span) / 24
			labelCells = append(labelCells, newPDFCell(x, width, field.Label, pdfColorLabelFill, pdfColorMuted, 9))
			valueCells = append(valueCells, newPDFCell(x, width, field.Value, pdfColorWhite, pdfColorText, 10))
			x += width
		}
		if x < pdfPageMargin+pdfContentWidth-0.1 {
			remaining := pdfPageMargin + pdfContentWidth - x
			labelCells = append(labelCells, newPDFCell(x, remaining, "", pdfColorLabelFill, pdfColorMuted, 9))
			valueCells = append(valueCells, newPDFCell(x, remaining, "", pdfColorWhite, pdfColorText, 10))
		}
		renderer.ensureSpace(54)
		renderer.drawRow(labelCells, 22, 12)
		renderer.drawRow(valueCells, 32, 14)
	}
}

func (renderer *pdfRenderer) drawTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		headers = []string{"内容"}
	}
	widths := pdfTableWidths(headers)
	headerCells := pdfTableCells(headers, widths, pdfColorLabelFill, pdfColorMuted, 9)
	renderer.ensureSpace(54)
	renderer.drawRow(headerCells, 25, 12)
	for _, row := range rows {
		values := make([]string, len(headers))
		copy(values, row)
		renderer.drawRow(pdfTableCells(values, widths, pdfColorWhite, pdfColorText, 9), 28, 13)
	}
}

func pdfTableCells(values []string, widths []float64, fill, textColor string, fontSize float64) []pdfCell {
	cells := make([]pdfCell, 0, len(widths))
	x := pdfPageMargin
	for index, width := range widths {
		value := ""
		if index < len(values) {
			value = values[index]
		}
		cells = append(cells, newPDFCell(x, width, value, fill, textColor, fontSize))
		x += width
	}
	return cells
}

func pdfTableWidths(headers []string) []float64 {
	weights := make([]float64, len(headers))
	for index := range weights {
		weights[index] = 1
	}
	if len(headers) == 2 && headers[0] == "字段" {
		weights = []float64{0.32, 0.68}
	}
	if len(headers) == 5 && headers[0] == "节点" && headers[3] == "意见" {
		weights = []float64{1.05, 0.85, 0.65, 1.45, 1.15}
	}
	total := 0.0
	for _, weight := range weights {
		total += weight
	}
	widths := make([]float64, len(weights))
	used := 0.0
	for index, weight := range weights {
		if index == len(weights)-1 {
			widths[index] = pdfContentWidth - used
			continue
		}
		widths[index] = pdfContentWidth * weight / total
		used += widths[index]
	}
	return widths
}

func newPDFCell(x, width float64, value, fill, textColor string, fontSize float64) pdfCell {
	return pdfCell{
		x:         x,
		width:     width,
		lines:     wrapPDFCellText(value, width-12, fontSize),
		fill:      fill,
		textColor: textColor,
		fontSize:  fontSize,
	}
}

func (renderer *pdfRenderer) drawRow(cells []pdfCell, minHeight, lineHeight float64) {
	maxLines := 1
	for _, cell := range cells {
		if len(cell.lines) > maxLines {
			maxLines = len(cell.lines)
		}
	}
	offset := 0
	for offset < maxLines {
		renderer.ensureSpace(minHeight)
		availableLines := int(math.Floor((pdfPageBottom - renderer.y - 10) / lineHeight))
		if availableLines < 1 {
			renderer.newPage(true)
			availableLines = int(math.Floor((pdfPageBottom - renderer.y - 10) / lineHeight))
		}
		lineCount := maxLines - offset
		if lineCount > availableLines {
			lineCount = availableLines
		}
		height := math.Max(minHeight, float64(lineCount)*lineHeight+10)
		for _, cell := range cells {
			renderer.drawRectangle(cell.x, renderer.y, cell.width, height, cell.fill, true)
			end := offset + lineCount
			if end > len(cell.lines) {
				end = len(cell.lines)
			}
			if offset < len(cell.lines) {
				for index, line := range cell.lines[offset:end] {
					renderer.drawText(cell.x+6, renderer.y+6+float64(index)*lineHeight, cell.fontSize, cell.textColor, line)
				}
			}
		}
		renderer.y += height
		offset += lineCount
		if offset < maxLines {
			renderer.newPage(true)
		}
	}
}

func (renderer *pdfRenderer) drawRectangle(x, top, width, height float64, fill string, border bool) {
	y := pdfPageHeight - top - height
	if fill != "" {
		fmt.Fprintf(renderer.current, "q\n%s rg\n%.2f %.2f %.2f %.2f re f\nQ\n", fill, x, y, width, height)
	}
	if border {
		fmt.Fprintf(renderer.current, "q\n%s RG\n0.5 w\n%.2f %.2f %.2f %.2f re S\nQ\n", pdfColorBorder, x, y, width, height)
	}
}

func (renderer *pdfRenderer) drawHorizontalLine(top float64, color string) {
	y := pdfPageHeight - top
	fmt.Fprintf(renderer.current, "q\n%s RG\n0.5 w\n%.2f %.2f m %.2f %.2f l S\nQ\n", color, pdfPageMargin, y, pdfPageMargin+pdfContentWidth, y)
}

func (renderer *pdfRenderer) drawText(x, top, size float64, color, value string) {
	if value == "" {
		return
	}
	baseline := pdfPageHeight - top - size
	cursor := x
	for _, run := range splitPDFTextRuns(value) {
		font := "F1"
		encoded := pdfTextHex(run.value)
		if run.ascii {
			font = "F2"
			encoded = strings.ToUpper(hex.EncodeToString([]byte(run.value)))
		}
		fmt.Fprintf(renderer.current, "BT\n/%s %.1f Tf\n0 Tc\n%s rg\n1 0 0 1 %.2f %.2f Tm\n<%s> Tj\nET\n", font, size, color, cursor, baseline, encoded)
		cursor += pdfTextWidth(run.value, size)
	}
}

func splitPDFTextRuns(value string) []pdfTextRun {
	runs := make([]pdfTextRun, 0)
	var current strings.Builder
	currentASCII := false
	for _, r := range value {
		ascii := r >= 32 && r <= 126
		if current.Len() > 0 && ascii != currentASCII {
			runs = append(runs, pdfTextRun{value: current.String(), ascii: currentASCII})
			current.Reset()
		}
		if current.Len() == 0 {
			currentASCII = ascii
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		runs = append(runs, pdfTextRun{value: current.String(), ascii: currentASCII})
	}
	return runs
}

func pdfTextWidth(value string, fontSize float64) float64 {
	width := 0.0
	for _, r := range value {
		width += pdfRuneWidth(r, fontSize)
	}
	return width
}

func wrapPDFCellText(value string, width, fontSize float64) []string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	result := make([]string, 0)
	for _, paragraph := range strings.Split(value, "\n") {
		if paragraph == "" {
			result = append(result, "")
			continue
		}
		var line strings.Builder
		lineWidth := 0.0
		for _, r := range paragraph {
			runeWidth := pdfRuneWidth(r, fontSize)
			if line.Len() > 0 && lineWidth+runeWidth > width {
				result = append(result, line.String())
				line.Reset()
				lineWidth = 0
			}
			line.WriteRune(r)
			lineWidth += runeWidth
		}
		result = append(result, line.String())
	}
	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func pdfRuneWidth(r rune, fontSize float64) float64 {
	if r == ' ' {
		return fontSize * 0.3
	}
	if r >= 32 && r <= 126 {
		return fontSize * 0.5
	}
	return fontSize
}

func buildPDFFile(pages [][]byte) []byte {
	if len(pages) == 0 {
		pages = [][]byte{nil}
	}
	fontObjectID := 3 + len(pages)*2
	cidFontObjectID := fontObjectID + 1
	latinFontObjectID := cidFontObjectID + 1
	objects := make([][]byte, latinFontObjectID+1)
	objects[1] = []byte("<< /Type /Catalog /Pages 2 0 R >>")
	pageIDs := make([]string, 0, len(pages))
	for index, content := range pages {
		pageObjectID := 3 + index*2
		contentObjectID := pageObjectID + 1
		pageIDs = append(pageIDs, fmt.Sprintf("%d 0 R", pageObjectID))
		objects[pageObjectID] = []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 %d 0 R /F2 %d 0 R >> >> /Contents %d 0 R >>", fontObjectID, latinFontObjectID, contentObjectID))
		objects[contentObjectID] = []byte(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	}
	objects[2] = []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(pageIDs, " "), len(pages)))
	objects[fontObjectID] = []byte(fmt.Sprintf("<< /Type /Font /Subtype /Type0 /BaseFont /STSong-Light /Encoding /UniGB-UCS2-H /DescendantFonts [%d 0 R] >>", cidFontObjectID))
	objects[cidFontObjectID] = []byte("<< /Type /Font /Subtype /CIDFontType0 /BaseFont /STSong-Light /CIDSystemInfo << /Registry (Adobe) /Ordering (GB1) /Supplement 5 >> /DW 1000 /W [32 126 500] >>")
	objects[latinFontObjectID] = []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")

	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objects))
	for objectID := 1; objectID < len(objects); objectID++ {
		offsets[objectID] = buffer.Len()
		fmt.Fprintf(&buffer, "%d 0 obj\n", objectID)
		buffer.Write(objects[objectID])
		buffer.WriteString("\nendobj\n")
	}
	xrefOffset := buffer.Len()
	fmt.Fprintf(&buffer, "xref\n0 %d\n", len(objects))
	buffer.WriteString("0000000000 65535 f \n")
	for objectID := 1; objectID < len(objects); objectID++ {
		fmt.Fprintf(&buffer, "%010d 00000 n \n", offsets[objectID])
	}
	fmt.Fprintf(&buffer, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects), xrefOffset)
	return buffer.Bytes()
}

func documentPDFLines(document exportDocument) []string {
	lines := []string{document.Title, ""}
	for _, row := range document.Metadata {
		lines = append(lines, joinExportRow(row))
	}
	for _, section := range document.Sections {
		lines = append(lines, "", "【"+section.Title+"】")
		for _, fields := range section.FieldRows {
			values := make([]string, 0, len(fields))
			for _, field := range fields {
				values = append(values, firstText(field.Label+"："+field.Value, field.Label))
			}
			lines = append(lines, joinExportRow(values))
		}
		if len(section.Headers) > 0 {
			lines = append(lines, joinExportRow(section.Headers))
		}
		for _, row := range section.Rows {
			lines = append(lines, joinExportRow(row))
		}
	}
	return lines
}

func joinExportRow(values []string) string {
	return strings.Join(values, "  |  ")
}

func wrapPDFText(value string, width int) []string {
	return wrapPDFCellText(value, float64(width)*10, 10)
}

func pdfPageContent(lines []string) string {
	var content strings.Builder
	content.WriteString("BT\n/F1 10 Tf\n50 800 Td\n")
	for index, line := range lines {
		if index > 0 {
			content.WriteString("0 -16 Td\n")
		}
		content.WriteString("<" + pdfTextHex(line) + "> Tj\n")
	}
	content.WriteString("ET")
	return content.String()
}

func pdfTextHex(value string) string {
	codes := utf16.Encode([]rune(value))
	data := make([]byte, 0, len(codes)*2)
	for _, code := range codes {
		data = append(data, byte(code>>8), byte(code))
	}
	return strings.ToUpper(hex.EncodeToString(data))
}
