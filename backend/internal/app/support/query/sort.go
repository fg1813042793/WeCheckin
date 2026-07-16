package query

import (
	"strings"
)

// ParseSort converts API sort expressions into a safe SQL order clause.
func ParseSort(sortStr string, allowedFields map[string]string) string {
	if sortStr == "" {
		return ""
	}
	parts := strings.Split(sortStr, ",")
	var orders []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		field := strings.TrimSpace(kv[0])
		order := "ASC"
		if len(kv) > 1 && strings.ToUpper(strings.TrimSpace(kv[1])) == "DESC" {
			order = "DESC"
		}
		dbField, ok := allowedFields[field]
		if !ok {
			continue
		}
		orders = append(orders, "`"+dbField+"` "+order)
	}
	if len(orders) == 0 {
		return ""
	}
	return strings.Join(orders, ", ")
}
