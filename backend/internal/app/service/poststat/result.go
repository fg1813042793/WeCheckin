package poststat

import (
	"fmt"
	"sort"
	"strings"

	"wecheckin-backend/backend/internal/app/formkit/report"
)

func buildResultText(stats []report.FieldStat, statMode string) string {
	modeLabel := "选项标签"
	if statMode == "value" {
		modeLabel = "选项值"
	}
	type kv struct {
		k string
		v int
	}
	agg := map[string]int{}
	for _, s := range stats {
		if s.Type == "divider" || s.Type == "description" {
			continue
		}
		for k, v := range s.Dist {
			agg[k] += v
		}
	}
	if len(agg) == 0 {
		return ""
	}
	var sorted []kv
	total := 0
	for k, v := range agg {
		sorted = append(sorted, kv{k, v})
		total += v
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	lines := []string{fmt.Sprintf("统计类型（%s）", modeLabel)}
	for _, p := range sorted {
		pct := float64(p.v) / float64(total) * 100
		lines = append(lines, fmt.Sprintf("%s 个数 %d 比重 %.1f%%", p.k, p.v, pct))
	}
	return strings.Join(lines, "\n")
}
