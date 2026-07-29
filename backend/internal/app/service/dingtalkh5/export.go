package dingtalkh5

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/model"
)

type ExportResult struct {
	Filename    string
	ContentType string
	Body        string
}

func ExportReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, filters ReviewFilters) (*ExportResult, error) {
	reviewList, err := listReviewsContext(ctx, user, filters, false)
	if err != nil {
		return nil, err
	}
	rows := []string{tableRow([]string{"考评月份", "目标月份", "员工账号", "部门", "直属上级", "HRBP", "状态", "目标得分", "价值观自评总分", "价值观上级总分", "价值观HRBP总分", "上级分档", "HRBP分档", "最终分档", "员工总结", "上级评价", "HRBP评价", "员工确认", "确认意见/异议原因", "确认时间", "HRBP备注"})}
	for _, review := range reviewList.List {
		rows = append(rows, tableRow([]string{
			review.Period,
			review.NextPeriod,
			review.EmployeeID,
			review.Department,
			review.ManagerID,
			review.HRBPID,
			review.Status,
			formatFloat(objectiveTotal(review.Objectives)),
			toString(valueTotal(review.Values, "self")),
			toString(valueTotal(review.Values, "manager")),
			toString(valueTotal(review.Values, "hrbp")),
			review.ManagerGrade,
			review.HRBPGrade,
			effectiveGrade(review),
			review.SelfSummary,
			review.ManagerComment,
			review.HRBPComment,
			employeeConfirmText(review.EmployeeConfirmResult),
			review.EmployeeConfirmComment,
			formatTime(review.EmployeeConfirmedAt),
			review.FinalNote,
		}))
	}
	body := "\ufeff<!doctype html><html><head><meta charset=\"utf-8\"></head><body><table>" + strings.Join(rows, "") + "</table></body></html>"
	return &ExportResult{
		Filename:    fmt.Sprintf("dingtalk-h5-performance-%s.xls", time.Now().Format("2006-01-02")),
		ContentType: "application/vnd.ms-excel; charset=utf-8",
		Body:        body,
	}, nil
}

func tableRow(cells []string) string {
	var b strings.Builder
	b.WriteString("<tr>")
	for _, cell := range cells {
		b.WriteString("<td>")
		b.WriteString(html.EscapeString(cell))
		b.WriteString("</td>")
	}
	b.WriteString("</tr>")
	return b.String()
}

func employeeConfirmText(value string) string {
	switch value {
	case "confirmed":
		return "已确认"
	case "disputed":
		return "有异议"
	default:
		return ""
	}
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", value), "0"), ".")
}

func formatTime(value int64) string {
	if value == 0 {
		return ""
	}
	return time.UnixMilli(value).Format("2006-01-02 15:04:05")
}
