package workflowsummary

import (
	"fmt"
	"strings"
)

const (
	answerSheetPDFBodySize   = 10.5
	answerSheetPDFBodyHeight = 16.0
)

// Keep answer-sheet pagination separate from workflow tables and their repeated
// page headers. Every paragraph, including titles, can flow over multiple pages.
type answerSheetPDFRenderer struct {
	*pdfRenderer
}

func renderAnswerSheetPDF(document exportDocument) []byte {
	renderer := &answerSheetPDFRenderer{newPDFRenderer(document.Title)}
	renderer.paragraph(document.Title, 19, 26, pdfColorText, true, 26)
	renderer.y += 12
	renderer.metadata(document.Metadata)
	for _, section := range document.Sections {
		renderer.y += 12
		if section.Title != "" {
			following := 8 + answerSheetPDFBodyHeight
			if len(section.AnswerBlocks) > 0 {
				block := section.AnswerBlocks[0]
				statusHeight := answerSheetPDFStatusHeight(block)
				following = 14 + answerSheetPDFStartHeight(block.Title, 11, 17, statusHeight)
			}
			renderer.paragraph(section.Title, 12, 19, pdfColorPrimary, true, following)
			renderer.y += 8
		}
		if len(section.AnswerBlocks) == 0 {
			renderer.paragraph("暂无题目", answerSheetPDFBodySize, answerSheetPDFBodyHeight, pdfColorMuted, false, 0)
			continue
		}
		for _, block := range section.AnswerBlocks {
			renderer.y += 6
			statusHeight := answerSheetPDFStatusHeight(block)
			renderer.paragraph(block.Title, 11, 17, pdfColorText, true, statusHeight)
			bodyStart := answerSheetPDFStartHeight(block.Answer, answerSheetPDFBodySize, answerSheetPDFBodyHeight, 0)
			renderer.paragraph(answerSheetStatus(block), 9.5, 14, pdfColorMuted, false, 4+14+bodyStart)
			renderer.y += 4
			renderer.answer("用户答案", block.Answer)
			renderer.answer("参考答案", block.ReferenceAnswer)
			renderer.y += 6
		}
	}
	return buildPDFFile(renderer.finish())
}

func (renderer *answerSheetPDFRenderer) ensureSpace(height float64) {
	if renderer.y+height > pdfPageBottom {
		// No unbounded repeated title: the full document title has already been
		// printed (and wrapped) at the beginning of the answer sheet.
		renderer.newPage(false)
	}
}

func (renderer *answerSheetPDFRenderer) paragraph(value string, size, lineHeight float64, color string, bold bool, followingHeight float64) {
	lines := wrapAnswerSheetPDFText(value, pdfContentWidth, size)
	renderer.ensureSpace(answerSheetPDFStartHeight(value, size, lineHeight, followingHeight))
	for index, line := range lines {
		required := lineHeight
		if index == len(lines)-1 {
			required += followingHeight
		}
		renderer.ensureSpace(required)
		renderer.text(pdfPageMargin, renderer.y, size, color, line, bold)
		renderer.y += lineHeight
	}
}

// Reserve the beginning of a paragraph, never the whole question/answer. Short
// heading chains stay with the first answer line; long headings can still split.
func answerSheetPDFStartHeight(value string, size, lineHeight, followingHeight float64) float64 {
	count := len(wrapAnswerSheetPDFText(value, pdfContentWidth, size))
	if count <= 2 {
		return float64(count)*lineHeight + followingHeight
	}
	return 2 * lineHeight
}

func answerSheetPDFStatusHeight(block AnswerBlock) float64 {
	bodyStart := answerSheetPDFStartHeight(block.Answer, answerSheetPDFBodySize, answerSheetPDFBodyHeight, 0)
	return answerSheetPDFStartHeight(answerSheetStatus(block), 9.5, 14, 4+14+bodyStart)
}

func (renderer *answerSheetPDFRenderer) text(x, top, size float64, color, value string, bold bool) {
	if bold {
		// Reuse both existing PDF fonts, with a light fill-and-stroke emphasis
		// for headings so CJK and Latin glyphs receive the same treatment.
		fmt.Fprintf(renderer.current, "q\n%s RG\n0.25 w\n2 Tr\n", color)
	}
	for _, run := range splitPDFTextRuns(value) {
		renderer.drawText(x, top, size, color, run.value)
		x += answerSheetPDFTextWidth(run.value, size)
	}
	if bold {
		renderer.current.WriteString("Q\n")
	}
}

func (renderer *answerSheetPDFRenderer) answer(label, value string) {
	bodyStart := answerSheetPDFStartHeight(value, answerSheetPDFBodySize, answerSheetPDFBodyHeight, 0)
	renderer.paragraph(label, 9.5, 14, pdfColorMuted, true, bodyStart)
	renderer.paragraph(value, answerSheetPDFBodySize, answerSheetPDFBodyHeight, pdfColorText, false, 0)
	renderer.y += 6
}

func (renderer *answerSheetPDFRenderer) metadata(rows [][]string) {
	const gap = 18.0
	width := (pdfContentWidth - gap) / 2
	for index := 0; index < len(rows); index += 2 {
		var columns [2][]string
		maxLines := 0
		for pair := 0; pair < 2 && index+pair < len(rows); pair++ {
			label, value := exportRowPair(rows[index+pair])
			columns[pair] = wrapAnswerSheetPDFText(strings.Join([]string{label, value}, "："), width, 9.5)
			if len(columns[pair]) > maxLines {
				maxLines = len(columns[pair])
			}
		}
		for line := 0; line < maxLines; line++ {
			renderer.ensureSpace(14)
			for column, lines := range columns {
				if line < len(lines) {
					renderer.text(pdfPageMargin+float64(column)*(width+gap), renderer.y, 9.5, pdfColorMuted, lines[line], false)
				}
			}
			renderer.y += 14
		}
		renderer.y += 4
	}
}

// Helvetica's printable ASCII advances (1/1000 em), matching the existing F2
// font. The workflow renderer's half-em approximation can overflow on W/M and
// overlap adjacent CJK runs; use these metrics only for answer sheets.
var answerSheetHelveticaWidths = [...]float64{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556,
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556,
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556,
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584,
}

func answerSheetPDFRuneWidth(r rune, size float64) float64 {
	if r >= 32 && r <= 126 {
		return answerSheetHelveticaWidths[r-32] * size / 1000
	}
	return size
}

func answerSheetPDFTextWidth(value string, size float64) float64 {
	width := 0.0
	for _, r := range value {
		width += answerSheetPDFRuneWidth(r, size)
	}
	return width
}

func wrapAnswerSheetPDFText(value string, width, size float64) []string {
	var lines []string
	for _, paragraph := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
		var line strings.Builder
		lineWidth := 0.0
		for _, r := range paragraph {
			runeWidth := answerSheetPDFRuneWidth(r, size)
			if line.Len() > 0 && lineWidth+runeWidth > width {
				lines = append(lines, line.String())
				line.Reset()
				lineWidth = 0
			}
			line.WriteRune(r)
			lineWidth += runeWidth
		}
		lines = append(lines, line.String())
	}
	return lines
}
