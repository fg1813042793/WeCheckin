package review

import "strings"

func sanitizeValueRubric(items []ValueRubric) []ValueRubric {
	result := make([]ValueRubric, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		if label == "" {
			continue
		}
		result = append(result, ValueRubric{
			Label:       label,
			Score:       clampNumber(item.Score, 0, 100),
			Description: strings.TrimSpace(item.Description),
		})
	}
	return result
}

func departmentText(parts ...string) string {
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return strings.Join(items, " / ")
}
