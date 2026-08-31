package workflowservice

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type OrgApproverIdentityItem struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Sort   int    `json:"sort"`
	Status int    `json:"status"`
}

type OrgApproverAssignmentItem struct {
	ID             uint   `json:"id"`
	DepartmentID   uint   `json:"departmentId"`
	DepartmentName string `json:"departmentName"`
	IdentityCode   string `json:"identityCode"`
	UserID         uint   `json:"userId"`
	UserName       string `json:"userName"`
	Sort           int    `json:"sort"`
	Status         int    `json:"status"`
}

type SaveOrgApproverAssignmentsRequest struct {
	DepartmentID uint   `json:"departmentId"`
	IdentityCode string `json:"identityCode"`
	UserIDs      []uint `json:"userIds"`
}

func ListOrgApproverIdentitiesContext(ctx context.Context) ([]OrgApproverIdentityItem, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var rows []model.WorkflowOrgApproverIdentity
	if err := db.Model(&model.WorkflowOrgApproverIdentity{}).
		Where("identity_status = ?", model.OrgApproverIdentityStatusEnabled).
		Order("identity_sort ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]OrgApproverIdentityItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, OrgApproverIdentityItem{
			ID:     row.ID,
			Code:   row.Code,
			Name:   row.Name,
			Sort:   row.Sort,
			Status: row.Status,
		})
	}
	return result, nil
}

func ListOrgApproverAssignmentsContext(ctx context.Context, departmentID uint, identityCode string) ([]OrgApproverAssignmentItem, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Table("workflow_org_approver_assignments AS a").
		Select("a.id, a.department_id, d.dept_name AS department_name, a.identity_code, a.user_id, u.user_name AS user_name, a.assignment_sort AS sort, a.assignment_status AS status").
		Joins("LEFT JOIN departments AS d ON d.id = a.department_id").
		Joins("LEFT JOIN users AS u ON u.id = a.user_id")
	if departmentID > 0 {
		query = query.Where("a.department_id = ?", departmentID)
	}
	if code := strings.TrimSpace(identityCode); code != "" {
		query = query.Where("a.identity_code = ?", code)
	}
	var rows []OrgApproverAssignmentItem
	if err := query.Order("a.department_id ASC, a.identity_code ASC, a.assignment_sort ASC, a.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func SaveOrgApproverAssignmentsContext(ctx context.Context, request SaveOrgApproverAssignmentsRequest) error {
	request.IdentityCode = strings.TrimSpace(request.IdentityCode)
	if request.DepartmentID == 0 {
		return errors.New("部门 ID 无效")
	}
	if request.IdentityCode == "" {
		return errors.New("组织审批身份不能为空")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := ensureOrgApproverIdentityAvailableTx(tx, request.IdentityCode); err != nil {
			return err
		}
		if err := ensureDepartmentAvailableTx(tx, request.DepartmentID); err != nil {
			return err
		}
		userIDs, err := ensureOrgApproverUsersAvailableTx(tx, request.UserIDs)
		if err != nil {
			return err
		}
		if err := tx.Where("department_id = ? AND identity_code = ?", request.DepartmentID, request.IdentityCode).
			Delete(&model.WorkflowOrgApproverAssignment{}).Error; err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return nil
		}
		now := database.Now()
		rows := make([]model.WorkflowOrgApproverAssignment, 0, len(userIDs))
		for index, userID := range userIDs {
			rows = append(rows, model.WorkflowOrgApproverAssignment{
				DepartmentID: request.DepartmentID,
				IdentityCode: request.IdentityCode,
				UserID:       userID,
				Sort:         index + 1,
				Status:       model.OrgApproverAssignmentStatusOn,
				AddTime:      now,
				EditTime:     now,
			})
		}
		return tx.CreateInBatches(rows, 100).Error
	})
}

func ensureOrgApproverIdentityAvailableTx(tx *gorm.DB, identityCode string) error {
	var count int64
	if err := tx.Model(&model.WorkflowOrgApproverIdentity{}).
		Where("identity_code = ? AND identity_status = ?", identityCode, model.OrgApproverIdentityStatusEnabled).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("组织审批身份不存在或已停用")
	}
	return nil
}

func ensureDepartmentAvailableTx(tx *gorm.DB, departmentID uint) error {
	var count int64
	if err := tx.Model(&model.Department{}).
		Where("id = ? AND dept_status = 1", departmentID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("部门不存在或已停用")
	}
	return nil
}

func ensureOrgApproverUsersAvailableTx(tx *gorm.DB, rawUserIDs []uint) ([]uint, error) {
	seen := map[uint]struct{}{}
	userIDs := make([]uint, 0, len(rawUserIDs))
	for _, userID := range rawUserIDs {
		if userID == 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}
	if len(userIDs) == 0 {
		return userIDs, nil
	}
	var activeUserIDs []uint
	if err := tx.Model(&model.User{}).
		Where("id IN ? AND user_status = 1", userIDs).
		Order("id ASC").
		Pluck("id", &activeUserIDs).Error; err != nil {
		return nil, err
	}
	active := map[uint]struct{}{}
	for _, userID := range activeUserIDs {
		active[userID] = struct{}{}
	}
	for _, userID := range userIDs {
		if _, ok := active[userID]; !ok {
			return nil, errors.New("组织审批身份关联用户不存在或已停用")
		}
	}
	return userIDs, nil
}
