package workflowsummary

import (
	"fmt"
	"strings"
)

// These styles belong only to answer sheets; workflow forms retain docxStyles.
// keepNext joins headings/labels to the start of their content. No question is
// enclosed in a table row, and body paragraphs may split across pages.
const answerSheetDOCXStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Microsoft YaHei" w:hAnsi="Microsoft YaHei" w:eastAsia="Microsoft YaHei" w:cs="Microsoft YaHei"/><w:sz w:val="21"/><w:szCs w:val="21"/><w:color w:val="1F2329"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="100" w:line="300" w:lineRule="auto"/><w:widowControl/></w:pPr></w:pPrDefault></w:docDefaults>
  <w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerBody"><w:name w:val="Answer Body"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext w:val="0"/><w:keepLines w:val="0"/></w:pPr></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerTitle"><w:name w:val="Answer Title"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext/><w:jc w:val="center"/><w:spacing w:after="220"/></w:pPr><w:rPr><w:b/><w:sz w:val="38"/><w:szCs w:val="38"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerSection"><w:name w:val="Answer Section"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:before="220" w:after="140"/></w:pPr><w:rPr><w:b/><w:sz w:val="24"/><w:szCs w:val="24"/><w:color w:val="126E68"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerQuestion"><w:name w:val="Answer Question"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext/><w:keepLines w:val="0"/><w:spacing w:before="160" w:after="60"/></w:pPr><w:rPr><w:b/><w:sz w:val="22"/><w:szCs w:val="22"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerStatus"><w:name w:val="Answer Status"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:after="120"/></w:pPr><w:rPr><w:sz w:val="19"/><w:szCs w:val="19"/><w:color w:val="505A69"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerLabel"><w:name w:val="Answer Label"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:after="40"/></w:pPr><w:rPr><w:b/><w:sz w:val="19"/><w:szCs w:val="19"/><w:color w:val="505A69"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="AnswerMetadata"><w:name w:val="Answer Metadata"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:after="0" w:line="260" w:lineRule="auto"/></w:pPr><w:rPr><w:sz w:val="19"/><w:szCs w:val="19"/></w:rPr></w:style>
</w:styles>`

func answerSheetDOCXDocument(document exportDocument) string {
	var body strings.Builder
	body.WriteString(docxParagraph(document.Title, "AnswerTitle"))
	if len(document.Metadata) > 0 {
		body.WriteString(answerSheetDOCXMetadata(document.Metadata))
	}
	for _, section := range document.Sections {
		if section.Title != "" {
			body.WriteString(docxParagraph(section.Title, "AnswerSection"))
		}
		if len(section.AnswerBlocks) == 0 {
			body.WriteString(docxParagraph("暂无题目", "AnswerBody"))
			continue
		}
		for _, block := range section.AnswerBlocks {
			body.WriteString(docxParagraph(block.Title, "AnswerQuestion"))
			body.WriteString(docxParagraph(answerSheetStatus(block), "AnswerStatus"))
			body.WriteString(docxParagraph("用户答案", "AnswerLabel"))
			body.WriteString(docxParagraph(block.Answer, "AnswerBody"))
			body.WriteString(docxParagraph("参考答案", "AnswerLabel"))
			body.WriteString(docxParagraph(block.ReferenceAnswer, "AnswerBody"))
		}
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + body.String() + `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="850" w:right="850" w:bottom="850" w:left="850"/><w:cols w:space="425"/></w:sectPr></w:body></w:document>`
}

func answerSheetStatus(block AnswerBlock) string {
	return "得分：" + block.Score + "    判题结果：" + block.Result
}

func answerSheetDOCXMetadata(rows [][]string) string {
	widths := []int{1100, 4003, 1100, 4003}
	var body strings.Builder
	body.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="10206" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="50" w:type="dxa"/><w:bottom w:w="50" w:type="dxa"/><w:right w:w="100" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`)
	for _, width := range widths {
		fmt.Fprintf(&body, `<w:gridCol w:w="%d"/>`, width)
	}
	body.WriteString(`</w:tblGrid>`)
	for index := 0; index < len(rows); index += 2 {
		body.WriteString(`<w:tr>`)
		for pair := 0; pair < 2; pair++ {
			label, value := "", ""
			if index+pair < len(rows) {
				label, value = exportRowPair(rows[index+pair])
			}
			for column, text := range []string{label, value} {
				fmt.Fprintf(&body, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/><w:vAlign w:val="top"/></w:tcPr><w:p><w:pPr><w:pStyle w:val="AnswerMetadata"/></w:pPr>`, widths[pair*2+column])
				body.WriteString(docxRun(text, column == 0))
				body.WriteString(`</w:p></w:tc>`)
			}
		}
		body.WriteString(`</w:tr>`)
	}
	body.WriteString(`</w:tbl>`)
	return body.String()
}
