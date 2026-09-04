package dingtalk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

var (
	ErrDatabaseUnavailable = errors.New("dingtalk admin database is not initialized")
	ErrPerfReviewNotFound  = errors.New("performance review not found")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type PerfReviewListItem struct {
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

type PerfReviewDetail struct {
	PerfReviewListItem
	ObjectivesJSON         string             `json:"objectivesJson"`
	NextObjectivesJSON     string             `json:"nextObjectivesJson"`
	ValuesJSON             string             `json:"valuesJson"`
	SelfSummary            string             `json:"selfSummary"`
	ManagerComment         string             `json:"managerComment"`
	HRBPComment            string             `json:"hrbpComment"`
	EmployeeConfirmComment string             `json:"employeeConfirmComment"`
	EmployeeConfirmedAt    int64              `json:"employeeConfirmedAt"`
	FinalNote              string             `json:"finalNote"`
	Histories              []PerfHistoryItem  `json:"histories"`
	HRBPReviewerAccount    string             `json:"hrbpReviewerAccount"`
	DepartmentLevel1       string             `json:"departmentLevel1"`
	DepartmentLevel2       string             `json:"departmentLevel2"`
	DepartmentLevel3       string             `json:"departmentLevel3"`
	Audit                  map[string]any     `json:"audit"`
	StatusOptions          []PerfStatusOption `json:"statusOptions"`
}

type PerfHistoryItem struct {
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

type PerfStatusOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type PerfReviewList struct {
	List          []PerfReviewListItem `json:"list"`
	Total         int64                `json:"total"`
	StatusOptions []PerfStatusOption   `json:"statusOptions"`
}

type PerfHistoryList struct {
	List  []PerfHistoryItem `json:"list"`
	Total int64             `json:"total"`
}

type PerfReviewQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Employee string
	Period   string
	Status   string
}

type PerfHistoryQuery struct {
	Page      int
	PageSize  int
	Keyword   string
	ReviewNo  string
	ByAccount string
	Action    string
}

type AdminIdentity struct {
	ID   uint
	Name string
}

func (service *Service) ListPerfReviews(ctx context.Context, query PerfReviewQuery) (PerfReviewList, error) {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return PerfReviewList{}, err
	}
	defer cancel()
	query.Page, query.PageSize = normalizePagination(query.Page, query.PageSize)
	statement := applyPerfReviewFilters(db.Model(&model.DingTalkH5PerfReview{}), query)
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return PerfReviewList{}, err
	}
	var rows []model.DingTalkH5PerfReview
	if err := statement.Order("`id` DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return PerfReviewList{}, err
	}
	list := make([]PerfReviewListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildPerfReviewListItem(row))
	}
	return PerfReviewList{List: list, Total: total, StatusOptions: perfStatusOptions()}, nil
}

func (service *Service) GetPerfReviewDetail(ctx context.Context, id uint) (PerfReviewDetail, error) {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return PerfReviewDetail{}, err
	}
	defer cancel()
	var row model.DingTalkH5PerfReview
	if err := db.Where("`id` = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PerfReviewDetail{}, ErrPerfReviewNotFound
		}
		return PerfReviewDetail{}, err
	}
	histories, err := perfReviewHistories(db, row.ID, row.ReviewNo)
	if err != nil {
		return PerfReviewDetail{}, err
	}
	return PerfReviewDetail{
		PerfReviewListItem:     buildPerfReviewListItem(row),
		ObjectivesJSON:         row.ObjectivesJSON,
		NextObjectivesJSON:     row.NextObjectivesJSON,
		ValuesJSON:             row.ValuesJSON,
		SelfSummary:            row.SelfSummary,
		ManagerComment:         row.ManagerComment,
		HRBPComment:            row.HRBPComment,
		EmployeeConfirmComment: row.EmployeeConfirmComment,
		EmployeeConfirmedAt:    row.EmployeeConfirmedAt,
		FinalNote:              row.FinalNote,
		Histories:              histories,
		HRBPReviewerAccount:    row.HRBPReviewerAccount,
		DepartmentLevel1:       row.DepartmentLevel1,
		DepartmentLevel2:       row.DepartmentLevel2,
		DepartmentLevel3:       row.DepartmentLevel3,
		StatusOptions:          perfStatusOptions(),
		Audit: map[string]any{
			"createBy": row.CreateBy, "updateBy": row.UpdateBy,
			"createDeptId": row.CreateDeptID, "updateDeptId": row.UpdateDeptID,
			"deleteBy": row.DeleteBy, "deleteDeptId": row.DeleteDeptID, "deletedAt": row.DeletedAt,
		},
	}, nil
}

func (service *Service) DeletePerfReview(ctx context.Context, id uint, actor AdminIdentity) error {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	now := database.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		var row model.DingTalkH5PerfReview
		if err := tx.Where("`id` = ?", id).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPerfReviewNotFound
			}
			return err
		}
		return deletePerfReviews(tx, []model.DingTalkH5PerfReview{row}, actor, now)
	})
}

func (service *Service) DeletePerfReviews(ctx context.Context, ids []uint, actor AdminIdentity) error {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	now := database.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		var rows []model.DingTalkH5PerfReview
		if err := tx.Where("`id` IN ?", ids).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return ErrPerfReviewNotFound
		}
		return deletePerfReviews(tx, rows, actor, now)
	})
}

func (service *Service) ListPerfHistories(ctx context.Context, query PerfHistoryQuery) (PerfHistoryList, error) {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return PerfHistoryList{}, err
	}
	defer cancel()
	query.Page, query.PageSize = normalizePagination(query.Page, query.PageSize)
	statement := db.Model(&model.DingTalkH5PerfHistory{}).Where("`deleted_at` = 0")
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		statement = statement.Where("`review_no` LIKE ? OR `by_account` LIKE ? OR `by_name` LIKE ? OR `action` LIKE ?", like, like, like, like)
	}
	if value := strings.TrimSpace(query.ReviewNo); value != "" {
		statement = statement.Where("`review_no` = ?", value)
	}
	if value := strings.TrimSpace(query.ByAccount); value != "" {
		statement = statement.Where("`by_account` = ?", value)
	}
	if value := strings.TrimSpace(query.Action); value != "" {
		statement = statement.Where("`action` LIKE ?", "%"+value+"%")
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return PerfHistoryList{}, err
	}
	var rows []model.DingTalkH5PerfHistory
	if err := statement.Order("`id` DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return PerfHistoryList{}, err
	}
	list := make([]PerfHistoryItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildPerfHistoryItem(row))
	}
	return PerfHistoryList{List: list, Total: total}, nil
}

func (service *Service) DeletePerfHistories(ctx context.Context, ids []uint) error {
	db, cancel, err := service.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Delete(&model.DingTalkH5PerfHistory{}, ids).Error
}

func (service *Service) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if service == nil || service.db == nil {
		return nil, func() {}, ErrDatabaseUnavailable
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return service.db.WithContext(queryCtx), cancel, nil
}

func deletePerfReviews(tx *gorm.DB, rows []model.DingTalkH5PerfReview, actor AdminIdentity, now int64) error {
	ids := make([]uint, 0, len(rows))
	histories := make([]model.DingTalkH5PerfHistory, 0, len(rows))
	for _, row := range rows {
		if row.ID == 0 {
			continue
		}
		ids = append(ids, row.ID)
		histories = append(histories, model.DingTalkH5PerfHistory{
			ReviewID: row.ID, ReviewNo: row.ReviewNo,
			ByAccount: fmt.Sprintf("admin:%d", actor.ID), ByName: actor.Name,
			Action: "后台物理删除考评单", AddTime: now, EditTime: now,
			DingTalkH5AuditFields: model.DingTalkH5AuditFields{CreateBy: actor.ID, UpdateBy: actor.ID},
		})
	}
	if len(ids) == 0 {
		return ErrPerfReviewNotFound
	}
	if len(histories) > 0 {
		if err := tx.Create(&histories).Error; err != nil {
			return err
		}
	}
	return tx.Delete(&model.DingTalkH5PerfReview{}, ids).Error
}

func applyPerfReviewFilters(statement *gorm.DB, query PerfReviewQuery) *gorm.DB {
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		statement = statement.Where("`review_no` LIKE ? OR `employee_account` LIKE ? OR `manager_account` LIKE ? OR `hrbp_account` LIKE ? OR `department` LIKE ?", like, like, like, like, like)
	}
	if value := strings.TrimSpace(query.Employee); value != "" {
		statement = statement.Where("`employee_account` = ?", value)
	}
	if value := strings.TrimSpace(query.Period); value != "" {
		statement = statement.Where("`period` = ?", value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		statement = statement.Where("`status` = ?", value)
	}
	return statement
}

func perfReviewHistories(db *gorm.DB, reviewID uint, reviewNo string) ([]PerfHistoryItem, error) {
	var rows []model.DingTalkH5PerfHistory
	statement := db.Model(&model.DingTalkH5PerfHistory{})
	if reviewID > 0 {
		statement = statement.Where("`review_id` = ?", reviewID)
	} else if reviewNo != "" {
		statement = statement.Where("`review_no` = ?", reviewNo)
	}
	if err := statement.Order("`id` DESC").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]PerfHistoryItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildPerfHistoryItem(row))
	}
	return list, nil
}

func buildPerfReviewListItem(row model.DingTalkH5PerfReview) PerfReviewListItem {
	return PerfReviewListItem{
		ID: row.ID, ReviewNo: row.ReviewNo, EmployeeAccount: row.EmployeeAccount,
		ManagerAccount: row.ManagerAccount, HRBPAccount: row.HRBPAccount, Department: row.Department,
		Period: row.Period, NextPeriod: row.NextPeriod, Status: row.Status, StatusLabel: perfStatusLabel(row.Status),
		TargetScore: targetScoreFromObjectives(row.ObjectivesJSON), ManagerGrade: row.ManagerGrade,
		HRBPGrade: row.HRBPGrade, FinalGrade: row.FinalGrade, EmployeeConfirm: row.EmployeeConfirmResult,
		AddTime: row.AddTime, EditTime: row.EditTime, CreateBy: row.CreateBy, UpdateBy: row.UpdateBy,
		CreateDeptID: row.CreateDeptID, UpdateDeptID: row.UpdateDeptID, DeleteBy: row.DeleteBy,
		DeleteDeptID: row.DeleteDeptID, DeletedAt: row.DeletedAt,
		ObjectiveSourceNo: row.ObjectiveSourceReviewNo, ObjectiveSourceTime: row.ObjectiveSourcePeriod,
	}
}

func buildPerfHistoryItem(row model.DingTalkH5PerfHistory) PerfHistoryItem {
	return PerfHistoryItem{
		ID: row.ID, ReviewID: row.ReviewID, ReviewNo: row.ReviewNo, ByAccount: row.ByAccount,
		ByName: row.ByName, Action: row.Action, AddTime: row.AddTime, EditTime: row.EditTime,
		CreateBy: row.CreateBy, UpdateBy: row.UpdateBy,
	}
}

func perfStatusOptions() []PerfStatusOption {
	return []PerfStatusOption{
		{Value: "employee_draft", Label: "员工草稿"}, {Value: "employee_fill", Label: "员工填写"},
		{Value: "manager_review", Label: "上级评价"}, {Value: "hrbp_review", Label: "HRBP评价"},
		{Value: "employee_confirm", Label: "员工确认"}, {Value: "hr_final", Label: "HRBP归档"},
		{Value: "completed", Label: "完成"},
	}
}

func perfStatusLabel(status string) string {
	for _, item := range perfStatusOptions() {
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
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	return 0
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
