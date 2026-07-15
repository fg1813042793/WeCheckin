package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func checkPublishDeptAccess(publishDeptIds string, userDeptIDs []uint) bool {
	if publishDeptIds == "" {
		return true
	}
	ids := strings.Split(publishDeptIds, ",")
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

func getJoinStatusDesc(status int) string {
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

func getTimeShow(t int64) string {
	return time.UnixMilli(t).Format("2006-01-02 15:04:05")
}

func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}

func getUserDeptIDsByMiniOpenID(miniOpenID string) []uint {
	var user model.User
	if err := database.DB.Where("`user_mini_openid` = ?", miniOpenID).First(&user).Error; err != nil {
		return nil
	}
	ids := getUserDeptIDs(user.ID)
	// Include all ancestor departments so users see items published to any parent department
	seen := map[uint]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	for _, id := range ids {
		for _, aid := range getAncestorDeptIDs(id) {
			if !seen[aid] {
				ids = append(ids, aid)
				seen[aid] = true
			}
		}
	}
	return ids
}

func getAncestorDeptIDs(deptID uint) []uint {
	var result []uint
	visited := map[uint]bool{}
	for deptID > 0 {
		if visited[deptID] {
			break
		}
		visited[deptID] = true
		var dept model.Department
		if err := database.DB.First(&dept, deptID).Error; err != nil {
			break
		}
		if dept.ParentID > 0 {
			result = append(result, dept.ParentID)
		}
		deptID = dept.ParentID
	}
	return result
}

func buildDeptOverlap(column string, deptIDs []uint) string {
	if len(deptIDs) == 0 {
		return "1 = 0"
	}
	parts := make([]string, len(deptIDs))
	for i, id := range deptIDs {
		parts[i] = fmt.Sprintf("FIND_IN_SET('%d', `%s`)", id, column)
	}
	return strings.Join(parts, " OR ")
}
