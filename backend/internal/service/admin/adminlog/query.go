package adminlog

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"wecheckin/backend/internal/model"
)

const (
	CategoryOperation       = "operation"
	CategoryLogin           = "login"
	MaxPageSize             = 100
	MaxExportSize           = 1000
	MaxPage                 = 1000000
	maxTimestamp      int64 = 253402300799999
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string  { return e.Field + ": " + e.Message }
func invalid(field, message string) error { return &ValidationError{Field: field, Message: message} }

// Query uses inclusive Unix millisecond boundaries, like the stored add_time.
type Query struct {
	Search    string
	Source    string
	Module    string
	Action    string
	Outcome   string
	ActorID   *uint
	IP        string
	StartTime int64
	EndTime   int64
	Page      int
	PageSize  int
}

type Page struct {
	List     []model.Log `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func ParseQuery(values url.Values, export bool) (Query, error) {
	q := Query{Page: 1, PageSize: 20}
	if export {
		q.PageSize = MaxExportSize
	}
	allowed := map[string]bool{"search": true, "source": true, "module": true, "action": true, "outcome": true, "actorId": true, "ip": true, "startTime": true, "endTime": true, "page": true, "pageSize": true}
	for key, vals := range values {
		if !allowed[key] || len(vals) != 1 {
			return q, invalid(key, "不支持的参数或重复参数")
		}
	}
	q.Search, q.Source, q.Module = strings.TrimSpace(values.Get("search")), strings.TrimSpace(values.Get("source")), strings.TrimSpace(values.Get("module"))
	q.Action, q.Outcome, q.IP = strings.TrimSpace(values.Get("action")), strings.TrimSpace(values.Get("outcome")), strings.TrimSpace(values.Get("ip"))
	for _, item := range []struct {
		name string
		dst  *int
	}{{"page", &q.Page}, {"pageSize", &q.PageSize}} {
		if raw, ok := values[item.name]; ok {
			n, err := strconv.Atoi(raw[0])
			if err != nil || n < 1 {
				return q, invalid(item.name, "必须为正整数")
			}
			*item.dst = n
		}
	}
	if raw, ok := values["actorId"]; ok {
		n, err := strconv.ParseUint(raw[0], 10, strconv.IntSize)
		if err != nil {
			return q, invalid("actorId", "必须为非负整数")
		}
		id := uint(n)
		q.ActorID = &id
	}
	for _, item := range []struct {
		name string
		dst  *int64
		end  bool
	}{{"startTime", &q.StartTime, false}, {"endTime", &q.EndTime, true}} {
		if raw, ok := values[item.name]; ok {
			ms, err := parseTime(raw[0], item.end)
			if err != nil {
				return q, invalid(item.name, "必须为毫秒时间戳、RFC3339时间或YYYY-MM-DD日期")
			}
			*item.dst = ms
		}
	}
	return q, q.Validate(export)
}

func (q Query) Validate(export bool) error {
	maxSize := MaxPageSize
	if export {
		maxSize = MaxExportSize
	}
	if q.Page < 1 || q.Page > MaxPage {
		return invalid("page", fmt.Sprintf("范围为1-%d", MaxPage))
	}
	if q.PageSize < 1 || q.PageSize > maxSize {
		return invalid("pageSize", fmt.Sprintf("范围为1-%d", maxSize))
	}
	if q.Source != "" && !oneOf(q.Source, "admin", "client", "h5app", "system") {
		return invalid("source", "不支持的来源")
	}
	if q.Outcome != "" && !validOutcome(q.Outcome) {
		return invalid("outcome", "不支持的结果")
	}
	for _, item := range []struct {
		name, value string
		max         int
	}{{"search", q.Search, 200}, {"module", q.Module, 64}, {"action", q.Action, 64}} {
		if err := validateText(item.name, item.value, item.max); err != nil {
			return err
		}
	}
	if q.IP != "" && net.ParseIP(q.IP) == nil {
		return invalid("ip", "无效的IP地址")
	}
	if q.StartTime < 0 || q.EndTime < 0 || q.StartTime > maxTimestamp || q.EndTime > maxTimestamp {
		return invalid("startTime/endTime", "时间超出范围")
	}
	if q.StartTime != 0 && q.EndTime != 0 && q.StartTime > q.EndTime {
		return invalid("startTime/endTime", "开始时间不能晚于结束时间")
	}
	return nil
}

func parseTime(raw string, end bool) (int64, error) {
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if n < 0 || n > maxTimestamp {
			return 0, fmt.Errorf("time out of range")
		}
		return n, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UnixMilli(), nil
	}
	t, err := time.ParseInLocation("2006-01-02", raw, time.FixedZone("Asia/Shanghai", 8*3600))
	if err != nil {
		return 0, err
	}
	if end {
		t = t.AddDate(0, 0, 1).Add(-time.Millisecond)
	}
	return t.UnixMilli(), nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}
func validOutcome(value string) bool {
	return oneOf(value, "success", "failure", "blocked", "partial", "accepted", "unknown")
}
func validateText(field, value string, max int) error {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > max {
		return invalid(field, fmt.Sprintf("最多%d个字符", max))
	}
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return invalid(field, "不能包含控制字符")
	}
	return nil
}
