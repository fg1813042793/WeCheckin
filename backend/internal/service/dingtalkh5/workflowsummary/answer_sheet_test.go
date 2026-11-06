package workflowsummary

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

func answerSheetFixture(t *testing.T, title, answer string) exportDocument {
	t.Helper()
	return exportDocument{
		Layout: DocumentLayoutAnswerSheet, Title: title, FileStem: "answer-sheet",
		Metadata: [][]string{{"考生", "张三"}, {"总分", "待批改"}, {"考试", "测试"}},
		Sections: []exportSection{{
			Title: "一、简答题", Headers: []string{"题目", "用户答案", "参考答案", "得分", "判题结果"},
			Rows: [][]string{{"LEGACY_ROW_ONLY", "原答案", "原参考", "0", "原结果"}},
			AnswerBlocks: []AnswerBlock{{Title: "1. 解释 <规则> & 边界", Answer: answer,
				ReferenceAnswer: "参考 <答案> & 原文\n第二行", Score: "待批改 / 10", Result: "待人工批改"}},
		}, {Title: "二、空章节"}},
	}
}

func TestAnswerSheetDOCXUsesReadingLayoutAndDedicatedStyles(t *testing.T) {
	document := answerSheetFixture(t, "考试答卷", "第一行 < & >\n\n  第二行  ")
	body, err := renderDOCX(document)
	if err != nil {
		t.Fatal(err)
	}
	entries := readZipEntries(t, body)
	content, styles := entries["word/document.xml"], entries["word/styles.xml"]
	for name, content := range entries {
		decoder := xml.NewDecoder(strings.NewReader(content))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("invalid XML in %s: %v", name, err)
			}
		}
	}
	for _, want := range []string{"1. 解释 &lt;规则&gt; &amp; 边界", "第一行 &lt; &amp; &gt;", "  第二行  ", "参考 &lt;答案&gt; &amp; 原文", "待批改 / 10", "待人工批改", "二、空章节", "暂无题目", `<w:br/>`, `w:w="11906" w:h="16838"`} {
		if !strings.Contains(content, want) {
			t.Errorf("DOCX missing %q", want)
		}
	}
	if strings.Contains(content, "LEGACY_ROW_ONLY") || strings.Contains(content, "w:cantSplit") {
		t.Error("answer sheet must use flowing paragraphs, without duplicate table rows or unsplittable questions")
	}
	if strings.Count(content, "<w:tbl>") != 1 || strings.Count(content, "<w:tr>") != 2 {
		t.Error("only metadata should be a table, with two key/value pairs per row")
	}
	for _, want := range []string{`w:styleId="AnswerBody"`, `w:styleId="AnswerQuestion"`, `w:sz w:val="21"`, `w:sz w:val="38"`, `w:eastAsia="Microsoft YaHei"`, `<w:keepNext/>`} {
		if !strings.Contains(styles, want) {
			t.Errorf("answer sheet styles missing %q", want)
		}
	}
	if styles == docxStyles {
		t.Error("answer sheet must select its own styles.xml")
	}
	if strings.Contains(styles, "w:pBdr") {
		t.Error("answer-sheet sections should use whitespace, not horizontal rules")
	}
}

func TestAnswerSheetLongAnswersRemainComplete(t *testing.T) {
	var answer strings.Builder
	for i := 0; i < 240; i++ {
		fmt.Fprintf(&answer, "答案第%03d行：完整保留 <内容> & 换行。\n", i)
	}
	document := answerSheetFixture(t, "长答案答卷", answer.String()+"ANSWER_END_MARKER")
	for _, format := range []ExportFormat{ExportFormatDOCX, ExportFormatPDF} {
		t.Run(string(format), func(t *testing.T) {
			result, err := RenderDocuments([]Document{document}, format)
			if err != nil {
				t.Fatal(err)
			}
			var text string
			if format == ExportFormatDOCX {
				text = readZipEntries(t, result.Body)["word/document.xml"]
			} else {
				pages := answerSheetPDFText(t, result.Body)
				if len(pages) < 4 {
					t.Errorf("long answer should flow over multiple pages, got %d", len(pages))
				}
				text = strings.Join(pages, "")
			}
			for i := 0; i < 240; i++ {
				want := fmt.Sprintf("答案第%03d行", i)
				if strings.Count(text, want) != 1 {
					t.Fatalf("answer line %q must occur exactly once", want)
				}
			}
			if !strings.Contains(text, "ANSWER_END_MARKER") || !strings.Contains(text, "第二行") {
				t.Error("answer or reference tail was truncated")
			}
		})
	}
}

func TestAnswerSheetPDFWrapsLongTitleAndPreservesReadingOrder(t *testing.T) {
	title := strings.Repeat("很长的考试名称", 35) + "标题末尾"
	document := answerSheetFixture(t, title, "我的完整回答")
	body, err := renderPDF(document)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(answerSheetPDFText(t, body), "")
	last := -1
	for _, want := range []string{title, "考生", "张三", "一、简答题", "1. 解释 <规则> & 边界", "待批改 / 10", "待人工批改", "用户答案", "我的完整回答", "参考答案", "参考 <答案> & 原文", "二、空章节", "暂无题目"} {
		index := strings.Index(text, want)
		if index < 0 || index <= last {
			t.Fatalf("missing or reordered text %q", want)
		}
		last = index
	}
	if strings.Contains(text, "LEGACY_ROW_ONLY") || strings.Contains(text, "流程信息") {
		t.Error("PDF answer sheet must not include workflow or legacy table content")
	}
	if bytes.Contains(body, []byte(" l S")) {
		t.Error("answer-sheet sections should use whitespace, not horizontal rules")
	}
}

func TestAnswerSheetOptInPreservesDefaultWorkflowAndXLSX(t *testing.T) {
	document := answerSheetFixture(t, "考试答卷", "BLOCK_ONLY")
	defaultDocument := answerSheetFixture(t, "考试答卷", "BLOCK_ONLY")
	defaultDocument.Layout = ""
	for index := range defaultDocument.Sections {
		defaultDocument.Sections[index].AnswerBlocks = nil
	}
	for _, layout := range []string{"", "unknown"} {
		// Adding answer blocks without the exact opt-in layout must do nothing.
		withBlocks := answerSheetFixture(t, "考试答卷", "BLOCK_ONLY")
		withBlocks.Layout = layout
		for _, format := range []ExportFormat{ExportFormatDOCX, ExportFormatPDF} {
			want, err := RenderDocuments([]Document{defaultDocument}, format)
			if err != nil {
				t.Fatal(err)
			}
			got, err := RenderDocuments([]Document{withBlocks}, format)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got.Body, want.Body) {
				t.Errorf("%s default layout changed for %q", format, layout)
			}
		}
	}
	want, err := renderXLSX([]exportDocument{defaultDocument})
	if err != nil {
		t.Fatal(err)
	}
	got, err := renderXLSX([]exportDocument{document})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Error("XLSX must keep consuming the original table regardless of Layout/AnswerBlocks")
	}
}

func TestAnswerSheetPDFKeepsSectionWithFirstQuestion(t *testing.T) {
	document := answerSheetFixture(t, "答卷", strings.TrimSuffix(strings.Repeat("回答\n", 27), "\n"))
	document.Sections[1].AnswerBlocks = []AnswerBlock{{Title: "2. 下一题", Answer: "回答", ReferenceAnswer: "参考", Score: "1", Result: "正确"}}
	body, err := renderPDF(document)
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range answerSheetPDFText(t, body) {
		if strings.Contains(page, "二、空章节") && !strings.Contains(page, "2. 下一题") {
			t.Fatal("section heading is orphaned at the foot of the previous page")
		}
	}
}

func TestAnswerSheetPDFWrapsWideLatinGlyphs(t *testing.T) {
	document := answerSheetFixture(t, strings.Repeat("W", 200), "回答")
	body, err := renderPDF(document)
	if err != nil {
		t.Fatal(err)
	}
	// Helvetica W is 944/1000 em, not the generic renderer's half-em estimate.
	for _, match := range regexp.MustCompile(`<((?:57)+)> Tj`).FindAllSubmatch(body, -1) {
		if float64(len(match[1])/2)*19*0.944 > pdfContentWidth {
			t.Fatal("wide Latin title glyphs overflow the printable width")
		}
	}
}

func TestAnswerSheetPDFKeepsAnswerLabelWithMultilineBody(t *testing.T) {
	renderer := &answerSheetPDFRenderer{newPDFRenderer("答卷")}
	renderer.y = pdfPageBottom - 35
	renderer.answer("参考答案", "第一行\n第二行")
	pages := answerSheetPDFText(t, buildPDFFile(renderer.finish()))
	for _, page := range pages {
		if strings.Contains(page, "参考答案") && !strings.Contains(page, "第一行") {
			t.Fatal("answer label must remain with the start of a multiline answer")
		}
	}
}

func TestAnswerSheetPDFLongQuestionAndReferenceFlowAcrossPages(t *testing.T) {
	document := answerSheetFixture(t, "答卷", "首行\n\n末行")
	block := &document.Sections[0].AnswerBlocks[0]
	block.Title = strings.Repeat("极长题目 WWWiii123\n", 100) + "QUESTION_END"
	block.ReferenceAnswer = strings.Repeat("参考内容 < & >\n", 100) + "REFERENCE_END"
	body, err := renderPDF(document)
	if err != nil {
		t.Fatal(err)
	}
	pages := answerSheetPDFText(t, body)
	text := strings.Join(pages, "")
	if len(pages) < 4 || strings.Count(text, "极长题目 WWWiii123") != 100 || strings.Count(text, "参考内容 < & >") != 100 {
		t.Fatal("long question/reference must paginate without losing or duplicating text")
	}
	for _, want := range []string{"QUESTION_END", "REFERENCE_END", "首行", "末行"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func answerSheetPDFText(t *testing.T, body []byte) []string {
	t.Helper()
	streams := regexp.MustCompile(`(?s)stream\n(.*?)\nendstream`).FindAllSubmatch(body, -1)
	runs := regexp.MustCompile(`(?s)BT\n/(F[12]) ([0-9.]+) Tf\n.*?1 0 0 1 ([0-9.]+) ([0-9.]+) Tm\n<([A-F0-9]+)> Tj\nET`).FindAllSubmatch
	pages := make([]string, 0, len(streams))
	for page, stream := range streams {
		var text strings.Builder
		for _, run := range runs(stream[1], -1) {
			data, err := hex.DecodeString(string(run[5]))
			if err != nil {
				t.Fatal(err)
			}
			value := string(data)
			if string(run[1]) == "F1" {
				codes := make([]uint16, 0, len(data)/2)
				for i := 0; i+1 < len(data); i += 2 {
					codes = append(codes, uint16(data[i])<<8|uint16(data[i+1]))
				}
				value = string(utf16.Decode(codes))
			}
			size, _ := strconv.ParseFloat(string(run[2]), 64)
			x, _ := strconv.ParseFloat(string(run[3]), 64)
			y, _ := strconv.ParseFloat(string(run[4]), 64)
			if x < pdfPageMargin-0.1 || x+answerSheetPDFTextWidth(value, size) > pdfPageWidth-pdfPageMargin+0.1 || y < pdfPageHeight-pdfPageBottom-0.1 || y+size > pdfPageHeight-pdfPageMargin+0.1 {
				t.Errorf("page %d text outside printable area: x=%.2f y=%.2f size=%.1f text=%q", page+1, x, y, size, value)
			}
			text.WriteString(value)
		}
		pages = append(pages, text.String())
	}
	return pages
}
