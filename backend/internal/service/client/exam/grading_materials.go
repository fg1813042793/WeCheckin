package exam

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	htmlparser "golang.org/x/net/html"
)

func gradingMaterials(question reviewQuestionSnapshot) ([]GradingMaterial, []GradingImage, string, bool) {
	var definition map[string]interface{}
	_ = json.Unmarshal(question.Definition, &definition)
	props, _ := definition["props"].(map[string]interface{})
	rows, _ := props["rows"].([]interface{})
	columns, _ := props["columns"].([]interface{})
	available := true
	if strings.HasPrefix(question.Type, "matrix") {
		available = len(columns) > 0 && (question.Type == "matrixAuto" || len(rows) > 0)
	}
	var materials []GradingMaterial
	var images []GradingImage
	plain := func(value, label string) string {
		text, found := gradingHTMLMaterial(value, label)
		images = append(images, found...)
		return text
	}
	_ = plain(schemaStringValue(definition["title"]), "题目配图")
	if show, _ := definition["showDescription"].(bool); definition["showDescription"] == nil || show {
		if description := plain(schemaStringValue(definition["description"]), "题目说明配图"); description != "" {
			materials = append(materials, GradingMaterial{Label: "题目说明", Content: description})
		}
	}
	for _, section := range []struct{ key, label string }{{"options", "选项"}, {"rows", "矩阵行"}, {"columns", "矩阵列"}, {"fields", "填空项"}} {
		items, _ := props[section.key].([]interface{})
		if len(items) == 0 {
			items, _ = definition[section.key].([]interface{})
		}
		lines := make([]string, 0, len(items))
		for index, item := range items {
			label := gradingMaterialLabel(item, fmt.Sprintf("第%d项", index+1))
			if section.key == "options" {
				label = plain(label, section.label+"配图")
			}
			if section.key == "options" {
				if entry, ok := item.(map[string]interface{}); ok {
					value := schemaStringValue(entry["value"])
					if value != "" && value != label {
						label = value + "：" + label
					}
				}
			}
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, label))
		}
		if len(lines) > 0 {
			materials = append(materials, GradingMaterial{Label: section.label, Content: strings.Join(lines, "\n")})
		}
	}
	for _, source := range []map[string]interface{}{definition, props} {
		for _, key := range []string{"image", "imageUrl", "picture", "pictureUrl"} {
			value := source[key]
			if entry, ok := value.(map[string]interface{}); ok {
				value = entry["url"]
				if value == nil {
					value = entry["src"]
				}
			}
			if imageURL := gradingImageURL(schemaStringValue(value)); imageURL != "" {
				images = append(images, GradingImage{URL: imageURL, Alt: "题目配图"})
			}
		}
	}
	uniqueImages := make([]GradingImage, 0, len(images))
	seen := make(map[string]bool)
	for _, image := range images {
		if !seen[image.URL] {
			uniqueImages = append(uniqueImages, image)
			seen[image.URL] = true
		}
	}
	reference := question.ReferenceAnswer
	if strings.TrimSpace(reference) == "" {
		fields, _ := props["fields"].([]interface{})
		var references []string
		for index, raw := range fields {
			field, _ := raw.(map[string]interface{})
			if answer, exists := field["examCorrectAnswer"]; exists {
				label := gradingMaterialLabel(field, fmt.Sprintf("第%d空", index+1))
				references = append(references, label+"："+gradingDisplayValue(answer))
			}
		}
		reference = strings.Join(references, "\n")
	}
	return materials, uniqueImages, reference, available
}

func gradingMaterialLabel(value interface{}, fallback string) string {
	if item, ok := value.(map[string]interface{}); ok {
		for _, key := range []string{"title", "label", "name", "placeholder", "value"} {
			if text := schemaStringValue(item[key]); text != "" {
				return text
			}
		}
		return fallback
	}
	if text := schemaStringValue(value); text != "" {
		return text
	}
	return fallback
}

func gradingHTMLMaterial(raw, label string) (string, []GradingImage) {
	document, err := htmlparser.Parse(strings.NewReader(raw))
	if err != nil {
		return stripExamExportHTML(raw), nil
	}
	var text strings.Builder
	var images []GradingImage
	var walk func(*htmlparser.Node)
	walk = func(node *htmlparser.Node) {
		if node.Type == htmlparser.ElementNode && (node.Data == "script" || node.Data == "style") {
			return
		}
		if node.Type == htmlparser.TextNode {
			text.WriteString(node.Data)
		}
		if node.Type == htmlparser.ElementNode && node.Data == "img" {
			var src, alt string
			for _, attr := range node.Attr {
				if attr.Key == "src" {
					src = attr.Val
				}
				if attr.Key == "alt" {
					alt = attr.Val
				}
			}
			if src = gradingImageURL(src); src != "" {
				images = append(images, GradingImage{URL: src, Alt: firstExamExportText(alt, label)})
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if node.Type == htmlparser.ElementNode && (node.Data == "p" || node.Data == "div" || node.Data == "br" || node.Data == "li") {
			text.WriteByte('\n')
		}
	}
	walk(document)
	return strings.TrimSpace(text.String()), images
}

func gradingImageURL(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.Contains(value, "\\") {
		return value
	}
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Hostname() != "" && parsed.User == nil {
		return parsed.String()
	}
	return ""
}

func gradingQuestionAnswer(value interface{}, question reviewQuestionSnapshot) string {
	if !strings.HasPrefix(question.Type, "matrix") {
		return gradingDisplayValue(value)
	}
	if encoded, ok := value.(string); ok {
		var decoded interface{}
		if json.Unmarshal([]byte(encoded), &decoded) != nil {
			return encoded
		}
		switch decoded.(type) {
		case []interface{}, map[string]interface{}:
			value = decoded
		default:
			return encoded
		}
	}
	switch value.(type) {
	case []interface{}, map[string]interface{}:
	default:
		return gradingDisplayValue(value)
	}
	var definition map[string]interface{}
	_ = json.Unmarshal(question.Definition, &definition)
	props, _ := definition["props"].(map[string]interface{})
	rows, _ := props["rows"].([]interface{})
	columns, _ := props["columns"].([]interface{})
	if question.Type == "matrixAuto" {
		if values, ok := value.([]interface{}); ok {
			rows = make([]interface{}, len(values))
		}
	}
	if len(rows) == 0 {
		return gradingDisplayValue(value)
	}
	var lines []string
	for index, row := range rows {
		label := gradingMaterialLabel(row, fmt.Sprintf("第%d行", index+1))
		answer := gradingIndexedValue(value, index, row)
		if cells, ok := answer.(map[string]interface{}); ok && len(columns) > 0 {
			var parts []string
			for colIndex, column := range columns {
				parts = append(parts, gradingMaterialLabel(column, fmt.Sprintf("第%d列", colIndex+1))+"："+gradingDisplayValue(gradingIndexedValue(cells, colIndex, column)))
			}
			lines = append(lines, label+"："+strings.Join(parts, "；"))
		} else {
			for _, raw := range columns {
				if column, ok := raw.(map[string]interface{}); ok && schemaStringValue(column["value"]) != "" && schemaStringValue(column["value"]) == schemaStringValue(answer) {
					answer = gradingMaterialLabel(column, "")
					break
				}
			}
			lines = append(lines, label+"："+gradingDisplayValue(answer))
		}
	}
	return strings.Join(lines, "\n")
}

func gradingIndexedValue(value interface{}, index int, definition interface{}) interface{} {
	switch values := value.(type) {
	case []interface{}:
		if index < len(values) {
			return values[index]
		}
	case map[string]interface{}:
		if entry, exists := values[fmt.Sprint(index)]; exists {
			return entry
		}
		if item, ok := definition.(map[string]interface{}); ok {
			for _, name := range []string{"id", "value", "title", "label"} {
				if key := schemaStringValue(item[name]); key != "" {
					if entry, exists := values[key]; exists {
						return entry
					}
				}
			}
		}
	}
	return nil
}
