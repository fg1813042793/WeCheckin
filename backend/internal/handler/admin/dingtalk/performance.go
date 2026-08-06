package dingtalk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/response"
)

type adminPerfReviewListItem struct {
	ID                  uint   `json:"id"`
	ReviewNo            string `json:"reviewNo"`
	EmployeeAccount     string `json:"employeeAccount"`
	ManagerAccount      string `json:"managerAccount"`
	HRBPAccount         string `json:"hrbpAccount"`
	Department          string `json:"department"`
	Period              string `json:"period"`
	NextPeriod          string `json:"nextPeriod"`
	Status              string `json:"status"`
	StatusLabel         string `json:"statusLabel"`
	TargetScore         int    `json:"targetScore"`
	ManagerGrade        string `json:"managerGrade"`
	HRBPGrade           string `json:"hrbpGrade"`
	FinalGrade          string `json:"finalGrade"`
	EmployeeConfirm     string `json:"employeeConfirm"`
	AddTime             int64  `json:"addTime"`
	EditTime            int64  `json:"editTime"`
	CreateBy            uint   `json:"createBy"`
	UpdateBy            uint   `json:"updateBy"`
	CreateDeptID        uint   `json:"createDeptId"`
	UpdateDeptID        uint   `json:"updateDeptId"`
	DeleteBy            uint   `json:"deleteBy"`
	DeleteDeptID        uint   `json:"deleteDeptId"`
	DeletedAt           int64  `json:"deletedAt"`
	ObjectiveSourceNo   string `json:"objectiveSourceReviewNo"`
	ObjectiveSourceTime string `json:"objectiveSourcePeriod"`
}

type adminPerfReviewDetail struct {
	adminPerfReviewListItem
	ObjectivesJSON         string                  `json:"objectivesJson"`
	NextObjectivesJSON     string                  `json:"nextObjectivesJson"`
	ValuesJSON             string                  `json:"valuesJson"`
	SelfSummary            string                  `json:"selfSummary"`
	ManagerComment         string                  `json:"managerComment"`
	HRBPComment            string                  `json:"hrbpComment"`
	EmployeeConfirmComment string                  `json:"employeeConfirmComment"`
	EmployeeConfirmedAt    int64                   `json:"employeeConfirmedAt"`
	FinalNote              string                  `json:"finalNote"`
	Histories              []adminPerfHistoryItem  `json:"histories"`
	HRBPReviewerAccount    string                  `json:"hrbpReviewerAccount"`
	DepartmentLevel1       string                  `json:"departmentLevel1"`
	DepartmentLevel2       string                  `json:"departmentLevel2"`
	DepartmentLevel3       string                  `json:"departmentLevel3"`
	RawAudit               map[string]interface{}  `json:"audit"`
	StatusOptions          []adminPerfStatusOption `json:"statusOptions"`
}

type adminPerfHistoryItem struct {
	ID        uint   `json:"id"`
	ReviewID  uint   `json:"reviewId"`
	ReviewNo  string `json:"reviewNo"`
	ByAccount string `json:"byAccount"`
	ByName    string `json:"byName"`
	Action    string `json:"action"`
	AddTime   int64  `json:"addTime"`
	EditTime  int64  `json:"editTime"`
	CreateBy  uint   `json:"createBy"`
	UpdateBy  uint   `json:"updateBy"`
}

type adminPerfStatusOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

func (h *AdminDingTalkHandler) GetPerfReviews(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}

	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	query := db.Model(&model.DingTalkH5PerfReview{})
	query = applyAdminPerfReviewFilters(query, c)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, "获取失败")
		return
	}
	var rows []model.DingTalkH5PerfReview
	if err := query.Order("`id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		response.Fail(c, "获取失败")
		return
	}

	list := make([]adminPerfReviewListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildAdminPerfReviewListItem(row))
	}
	response.JSON(c, map[string]interface{}{
		"list":          list,
		"total":         total,
		"statusOptions": adminPerfStatusOptions(),
	})
}

func (h *AdminDingTalkHandler) GetPerfReviewDetail(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	id := parseUint(c.Query("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}

	var row model.DingTalkH5PerfReview
	err := db.Where("`id` = ?", id).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, "考评单不存在")
			return
		}
		response.Fail(c, "获取失败")
		return
	}
	histories, err := adminPerfReviewHistories(db, row.ID, row.ReviewNo)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}

	item := adminPerfReviewDetail{
		adminPerfReviewListItem: buildAdminPerfReviewListItem(row),
		ObjectivesJSON:          row.ObjectivesJSON,
		NextObjectivesJSON:      row.NextObjectivesJSON,
		ValuesJSON:              row.ValuesJSON,
		SelfSummary:             row.SelfSummary,
		ManagerComment:          row.ManagerComment,
		HRBPComment:             row.HRBPComment,
		EmployeeConfirmComment:  row.EmployeeConfirmComment,
		EmployeeConfirmedAt:     row.EmployeeConfirmedAt,
		FinalNote:               row.FinalNote,
		Histories:               histories,
		HRBPReviewerAccount:     row.HRBPReviewerAccount,
		DepartmentLevel1:        row.DepartmentLevel1,
		DepartmentLevel2:        row.DepartmentLevel2,
		DepartmentLevel3:        row.DepartmentLevel3,
		StatusOptions:           adminPerfStatusOptions(),
		RawAudit: map[string]interface{}{
			"createBy":     row.CreateBy,
			"updateBy":     row.UpdateBy,
			"createDeptId": row.CreateDeptID,
			"updateDeptId": row.UpdateDeptID,
			"deleteBy":     row.DeleteBy,
			"deleteDeptId": row.DeleteDeptID,
			"deletedAt":    row.DeletedAt,
		},
	}
	response.JSON(c, item)
}

func (h *AdminDingTalkHandler) DeletePerfReview(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	adminID, adminName := currentAdminIdentity(c)
	now := database.Now()

	err := db.Transaction(func(tx *gorm.DB) error {
		var row model.DingTalkH5PerfReview
		if err := tx.Where("`id` = ?", id).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("考评单不存在")
			}
			return err
		}
		return deleteAdminPerfReviews(tx, []model.DingTalkH5PerfReview{row}, adminID, adminName, now)
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) DeletePerfReviews(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	ids := parseUintList(c.PostForm("ids"))
	if len(ids) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	adminID, adminName := currentAdminIdentity(c)
	now := database.Now()

	err := db.Transaction(func(tx *gorm.DB) error {
		var rows []model.DingTalkH5PerfReview
		if err := tx.Where("`id` IN ?", ids).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return errors.New("考评单不存在")
		}
		return deleteAdminPerfReviews(tx, rows, adminID, adminName, now)
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func deleteAdminPerfReviews(tx *gorm.DB, rows []model.DingTalkH5PerfReview, adminID uint, adminName string, now int64) error {
	ids := make([]uint, 0, len(rows))
	histories := make([]model.DingTalkH5PerfHistory, 0, len(rows))
	for _, row := range rows {
		if row.ID == 0 {
			continue
		}
		ids = append(ids, row.ID)
		histories = append(histories, model.DingTalkH5PerfHistory{
			ReviewID:  row.ID,
			ReviewNo:  row.ReviewNo,
			ByAccount: fmt.Sprintf("admin:%d", adminID),
			ByName:    adminName,
			Action:    "后台物理删除考评单",
			AddTime:   now,
			EditTime:  now,
			DingTalkH5AuditFields: model.DingTalkH5AuditFields{
				CreateBy: adminID,
				UpdateBy: adminID,
			},
		})
	}
	if len(ids) == 0 {
		return errors.New("考评单不存在")
	}
	if len(histories) > 0 {
		if err := tx.Create(&histories).Error; err != nil {
			return err
		}
	}
	return tx.Delete(&model.DingTalkH5PerfReview{}, ids).Error
}

func parseUintList(raw string) []uint {
	parts := strings.Split(raw, ",")
	ids := make([]uint, 0, len(parts))
	seen := make(map[uint]struct{}, len(parts))
	for _, part := range parts {
		id := parseUint(part)
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func (h *AdminDingTalkHandler) GetPerfHistories(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}

	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	query := db.Model(&model.DingTalkH5PerfHistory{}).Where("`deleted_at` = 0")
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("`review_no` LIKE ? OR `by_account` LIKE ? OR `by_name` LIKE ? OR `action` LIKE ?", likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	}
	reviewNo := strings.TrimSpace(c.Query("reviewNo"))
	if reviewNo != "" {
		query = query.Where("`review_no` = ?", reviewNo)
	}
	byAccount := strings.TrimSpace(c.Query("byAccount"))
	if byAccount != "" {
		query = query.Where("`by_account` = ?", byAccount)
	}
	action := strings.TrimSpace(c.Query("action"))
	if action != "" {
		query = query.Where("`action` LIKE ?", "%"+action+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, "获取失败")
		return
	}
	var rows []model.DingTalkH5PerfHistory
	if err := query.Order("`id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		response.Fail(c, "获取失败")
		return
	}
	list := make([]adminPerfHistoryItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildAdminPerfHistoryItem(row))
	}
	response.JSON(c, map[string]interface{}{
		"list":  list,
		"total": total,
	})
}

func (h *AdminDingTalkHandler) DeletePerfHistory(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := db.Delete(&model.DingTalkH5PerfHistory{}, id).Error; err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) DeletePerfHistories(ctx context.Context, c *app.RequestContext) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		response.Fail(c, "数据库未初始化")
		return
	}
	ids := parseUintList(c.PostForm("ids"))
	if len(ids) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := db.Delete(&model.DingTalkH5PerfHistory{}, ids).Error; err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func applyAdminPerfReviewFilters(query *gorm.DB, c *app.RequestContext) *gorm.DB {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("`review_no` LIKE ? OR `employee_account` LIKE ? OR `manager_account` LIKE ? OR `hrbp_account` LIKE ? OR `department` LIKE ?", likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	}
	employee := strings.TrimSpace(c.Query("employee"))
	if employee != "" {
		query = query.Where("`employee_account` = ?", employee)
	}
	period := strings.TrimSpace(c.Query("period"))
	if period != "" {
		query = query.Where("`period` = ?", period)
	}
	status := strings.TrimSpace(c.Query("status"))
	if status != "" {
		query = query.Where("`status` = ?", status)
	}
	return query
}

func adminPerfReviewHistories(db *gorm.DB, reviewID uint, reviewNo string) ([]adminPerfHistoryItem, error) {
	var rows []model.DingTalkH5PerfHistory
	query := db.Model(&model.DingTalkH5PerfHistory{})
	if reviewID > 0 {
		query = query.Where("`review_id` = ?", reviewID)
	} else if reviewNo != "" {
		query = query.Where("`review_no` = ?", reviewNo)
	}
	if err := query.Order("`id` DESC").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]adminPerfHistoryItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildAdminPerfHistoryItem(row))
	}
	return list, nil
}

func buildAdminPerfReviewListItem(row model.DingTalkH5PerfReview) adminPerfReviewListItem {
	return adminPerfReviewListItem{
		ID:                  row.ID,
		ReviewNo:            row.ReviewNo,
		EmployeeAccount:     row.EmployeeAccount,
		ManagerAccount:      row.ManagerAccount,
		HRBPAccount:         row.HRBPAccount,
		Department:          row.Department,
		Period:              row.Period,
		NextPeriod:          row.NextPeriod,
		Status:              row.Status,
		StatusLabel:         adminPerfStatusLabel(row.Status),
		TargetScore:         targetScoreFromObjectives(row.ObjectivesJSON),
		ManagerGrade:        row.ManagerGrade,
		HRBPGrade:           row.HRBPGrade,
		FinalGrade:          row.FinalGrade,
		EmployeeConfirm:     row.EmployeeConfirmResult,
		AddTime:             row.AddTime,
		EditTime:            row.EditTime,
		CreateBy:            row.CreateBy,
		UpdateBy:            row.UpdateBy,
		CreateDeptID:        row.CreateDeptID,
		UpdateDeptID:        row.UpdateDeptID,
		DeleteBy:            row.DeleteBy,
		DeleteDeptID:        row.DeleteDeptID,
		DeletedAt:           row.DeletedAt,
		ObjectiveSourceNo:   row.ObjectiveSourceReviewNo,
		ObjectiveSourceTime: row.ObjectiveSourcePeriod,
	}
}

func buildAdminPerfHistoryItem(row model.DingTalkH5PerfHistory) adminPerfHistoryItem {
	return adminPerfHistoryItem{
		ID:        row.ID,
		ReviewID:  row.ReviewID,
		ReviewNo:  row.ReviewNo,
		ByAccount: row.ByAccount,
		ByName:    row.ByName,
		Action:    row.Action,
		AddTime:   row.AddTime,
		EditTime:  row.EditTime,
		CreateBy:  row.CreateBy,
		UpdateBy:  row.UpdateBy,
	}
}

func adminPerfStatusOptions() []adminPerfStatusOption {
	return []adminPerfStatusOption{
		{Value: "employee_draft", Label: "员工草稿"},
		{Value: "employee_fill", Label: "员工填写"},
		{Value: "manager_review", Label: "上级评价"},
		{Value: "hrbp_review", Label: "HRBP评价"},
		{Value: "employee_confirm", Label: "员工确认"},
		{Value: "hr_final", Label: "HRBP归档"},
		{Value: "completed", Label: "完成"},
	}
}

func adminPerfStatusLabel(status string) string {
	for _, item := range adminPerfStatusOptions() {
		if item.Value == status {
			return item.Label
		}
	}
	if strings.TrimSpace(status) == "" {
		return "未设置"
	}
	return status
}

func targetScoreFromObjectives(raw string) int {
	// 后台列表只需要一个轻量汇总，避免把大 JSON 展开成复杂结构。
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	return 0
}

func currentAdminIdentity(c *app.RequestContext) (uint, string) {
	adminVal, ok := c.Get("admin")
	if !ok || adminVal == nil {
		return 0, "管理员"
	}
	admin, ok := adminVal.(*model.Admin)
	if !ok || admin == nil {
		return 0, "管理员"
	}
	name := strings.TrimSpace(admin.Name)
	if name == "" {
		name = fmt.Sprintf("管理员%d", admin.ID)
	}
	return admin.ID, name
}
