package service

import (
	"strings"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ===================== DataScope =====================

// BuildDataScopeFilter returns a WHERE condition string and args for data scope filtering.
// Supports per-table field names for dept_id and create_by.
// For User/Mgr tables that use association tables, use a separate approach.
func BuildDataScopeFilter(admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	if admin.Type == 1 {
		return "", nil
	}
	var role model.Role
	if err := database.DB.First(&role, admin.RoleID).Error; err != nil {
		return "", nil
	}
	switch role.DataScope {
	case 1: // 全部
		return "", nil
	case 2: // 本部门
		deptIDs := getAdminDeptIDs(admin.ID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		args := make([]interface{}, len(deptIDs))
		for i, d := range deptIDs {
			args[i] = d
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{args}
	case 3: // 本人
		return createByField + " = ?", []interface{}{admin.ID}
	case 4: // 自定义
		deptIDs := GetRoleDeptIDs(admin.RoleID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		args := make([]interface{}, len(deptIDs))
		for i, d := range deptIDs {
			args[i] = d
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{args}
	}
	return "", nil
}

func parseSort(sortStr string, allowedFields map[string]string) string {
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
