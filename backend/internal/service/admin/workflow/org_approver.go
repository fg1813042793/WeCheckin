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
	ID           uint   `json:"id"`
	SubjectType  string `json:"subjectType"`
	SubjectID    uint   `json:"subjectId"`
	SubjectName  string `json:"subjectName"`
	DepartmentID uint   `json:"departmentId,omitempty"`
	IdentityCode string `json:"identityCode"`
	UserID       uint   `json:"userId"`
	UserName     string `json:"userName"`
	Sort         int    `json:"sort"`
	Status       int    `json:"status"`
}

type SaveOrgApproverAssignmentsRequest struct {
	DepartmentID uint   `json:"departmentId"`
	SubjectType  string `json:"subjectType"`
	SubjectID    uint   `json:"subjectId"`
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

func ListOrgApproverAssignmentsContext(ctx context.Context, subjectType string, subjectID uint, identityCode string) ([]OrgApproverAssignmentItem, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	normalizedSubjectType := normalizeOrgApproverSubjectType(subjectType)
	if strings.TrimSpace(subjectType) != "" && normalizedSubjectType == "" {
		return nil, errors.New("适用对象类型无效")
	}
	if subjectID > 0 && normalizedSubjectType == "" {
		return nil, errors.New("适用对象类型不能为空")
	}
	query := db.Table("workflow_org_approver_assignments AS a").
		Select("a.id, a.subject_type, a.subject_id, CASE WHEN a.subject_type = 'department' THEN d.dept_name ELSE subject_user.user_name END AS subject_name, a.department_id, a.identity_code, a.user_id, u.user_name AS user_name, a.assignment_sort AS sort, a.assignment_status AS status").
		Joins("LEFT JOIN departments AS d ON a.subject_type = 'department' AND d.id = a.subject_id").
		Joins("LEFT JOIN users AS subject_user ON a.subject_type = 'user' AND subject_user.id = a.subject_id").
		Joins("LEFT JOIN users AS u ON u.id = a.user_id")
	if normalizedSubjectType != "" {
		query = query.Where("a.subject_type = ?", normalizedSubjectType)
	}
	if subjectID > 0 {
		query = query.Where("a.subject_id = ?", subjectID)
	}
	if code := strings.TrimSpace(identityCode); code != "" {
		query = query.Where("a.identity_code = ?", code)
	}
	var rows []OrgApproverAssignmentItem
	if err := query.Order("a.subject_type ASC, a.subject_id ASC, a.identity_code ASC, a.assignment_sort ASC, a.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func SaveOrgApproverAssignmentsContext(ctx context.Context, request SaveOrgApproverAssignmentsRequest) error {
	request.IdentityCode = strings.TrimSpace(request.IdentityCode)
	request.SubjectType = normalizeOrgApproverSubjectType(request.SubjectType)
	if request.SubjectType == "" && request.DepartmentID > 0 {
		request.SubjectType = model.OrgApproverSubjectTypeDepartment
		request.SubjectID = request.DepartmentID
	}
	if request.SubjectID == 0 {
		return errors.New("适用对象 ID 无效")
	}
	if request.SubjectType == "" {
		return errors.New("适用对象类型无效")
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
		if err := ensureOrgApproverSubjectAvailableTx(tx, request.SubjectType, request.SubjectID); err != nil {
			return err
		}
		userIDs, err := ensureOrgApproverUsersAvailableTx(tx, request.UserIDs)
		if err != nil {
			return err
		}
		if err := tx.Where("subject_type = ? AND subject_id = ? AND identity_code = ?", request.SubjectType, request.SubjectID, request.IdentityCode).
			Delete(&model.WorkflowOrgApproverAssignment{}).Error; err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return nil
		}
		now := database.Now()
		rows := make([]model.WorkflowOrgApproverAssignment, 0, len(userIDs))
		for index, userID := range userIDs {
			departmentID := uint(0)
			if request.SubjectType == model.OrgApproverSubjectTypeDepartment {
				departmentID = request.SubjectID
			}
			rows = append(rows, model.WorkflowOrgApproverAssignment{
				SubjectType:  request.SubjectType,
				SubjectID:    request.SubjectID,
				DepartmentID: departmentID,
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

func normalizeOrgApproverSubjectType(value string) string {
	switch strings.TrimSpace(value) {
	case model.OrgApproverSubjectTypeDepartment:
		return model.OrgApproverSubjectTypeDepartment
	case model.OrgApproverSubjectTypeUser:
		return model.OrgApproverSubjectTypeUser
	default:
		return ""
	}
}

func ensureOrgApproverSubjectAvailableTx(tx *gorm.DB, subjectType string, subjectID uint) error {
	if subjectType == model.OrgApproverSubjectTypeDepartment {
		return ensureDepartmentAvailableTx(tx, subjectID)
	}
	if subjectType == model.OrgApproverSubjectTypeUser {
		var count int64
		if err := tx.Model(&model.User{}).Where("id = ? AND user_status = 1", subjectID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("适用人员不存在或已停用")
		}
		return nil
	}
	return errors.New("适用对象类型无效")
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
