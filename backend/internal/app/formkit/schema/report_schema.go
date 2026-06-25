package schema

// ReportQuestion 报表专用：统一老/新 schema 为一个简单的 question 列表
type ReportQuestion struct {
	ID       string
	Title    string
	Type     string
	Options  []map[string]interface{} // select/radio/checkbox/picker/matrix 等
	OldIndex int                      // 老格式数组里的索引（-1 表示新格式）
}

// NormalizeSchemaForReport 把 schemaJSON 统一为 ReportQuestion 列表。
// 老格式：每项有 label/type，无 id（用 q1/q2/...），OldIndex 为数组下标。
// 新格式：每项有 id/type/title，OldIndex = -1。
func NormalizeSchemaForReport(schemaJSON string) []ReportQuestion {
	if schemaJSON == "" {
		return nil
	}
	var raw interface{}
	if err := jsonUnmarshal(schemaJSON, &raw); err != nil {
		return nil
	}
	out := []ReportQuestion{}
	if IsOldFormat(schemaJSON) {
		// 数组
		arr, ok := raw.([]interface{})
		if !ok {
			if mArr, ok := raw.([]map[string]interface{}); ok {
				arr = make([]interface{}, len(mArr))
				for i, m := range mArr {
					arr[i] = m
				}
			} else {
				return out
			}
		}
		for i, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			label, _ := m["label"].(string)
			typ, _ := m["type"].(string)
			if typ == "" {
				typ = "input"
			}
			out = append(out, ReportQuestion{
				ID:       fmtSprintfForReport(i + 1),
				Title:    label,
				Type:     typ,
				Options:  extractOptions(m),
				OldIndex: i,
			})
		}
		return out
	}
	// 新格式
	m, ok := raw.(map[string]interface{})
	if !ok {
		return out
	}
	qs, _ := m["questions"].([]interface{})
	if qs == nil {
		if qsAlt, ok := m["questions"].([]map[string]interface{}); ok {
			qs = make([]interface{}, len(qsAlt))
			for i, q := range qsAlt {
				qs[i] = q
			}
		}
	}
	for _, item := range qs {
		qm, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := qm["id"].(string)
		typ, _ := qm["type"].(string)
		title, _ := qm["title"].(string)
		if id == "" || typ == "" {
			continue
		}
		out = append(out, ReportQuestion{ID: id, Title: title, Type: typ, OldIndex: -1, Options: extractOptions(qm)})
	}
	return out
}

func extractOptions(m map[string]interface{}) []map[string]interface{} {
	if m == nil {
		return nil
	}
	// 优先从 props.options（新格式）
	if props, ok := m["props"].(map[string]interface{}); ok {
		if opts, ok := props["options"].([]interface{}); ok {
			out := make([]map[string]interface{}, 0, len(opts))
			for _, o := range opts {
				if om, ok := o.(map[string]interface{}); ok {
					out = append(out, om)
				}
			}
			return out
		}
	}
	// 老格式：options 直接挂在 item
	if opts, ok := m["options"].([]interface{}); ok {
		out := make([]map[string]interface{}, 0, len(opts))
		for _, o := range opts {
			if om, ok := o.(map[string]interface{}); ok {
				out = append(out, om)
			}
		}
		return out
	}
	return nil
}
