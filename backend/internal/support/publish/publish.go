package publish

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// HasDeptAccess reports whether a user's department list can access a published item.
func HasDeptAccess(publishDeptIDs string, userDeptIDs []uint) bool {
	if publishDeptIDs == "" {
		return true
	}
	ids := strings.Split(publishDeptIDs, ",")
	for _, pid := range ids {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		for _, uid := range userDeptIDs {
			if strconv.FormatUint(uint64(uid), 10) == pid {
				return true
			}
		}
	}
	return false
}

// JoinStatusDesc returns the display text for enroll join status.
func JoinStatusDesc(status int) string {
	switch status {
	case 0:
		return "待审核"
	case 1:
		return "已通过"
	case 2:
		return "未通过"
	default:
		return "未知"
	}
}

// FormatUnixMilli formats a millisecond timestamp for admin/client display.
func FormatUnixMilli(t int64) string {
	return time.UnixMilli(t).Format("2006-01-02 15:04:05")
}

// UnixMilli converts a millisecond timestamp to time.Time.
func UnixMilli(ms int64) time.Time {
	return time.UnixMilli(ms)
}

// DeptOverlap builds a FIND_IN_SET SQL expression for comma-separated department IDs.
func DeptOverlap(column string, deptIDs []uint) string {
	if len(deptIDs) == 0 {
		return "1 = 0"
	}
	parts := make([]string, len(deptIDs))
	for i, id := range deptIDs {
		parts[i] = fmt.Sprintf("FIND_IN_SET('%d', `%s`)", id, column)
	}
	return strings.Join(parts, " OR ")
}
