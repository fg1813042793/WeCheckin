package access

import "strings"

// ResourceAuditFields centralizes audit and data-scope column names.
// Resource tables use the standard audit columns, while a few business columns
// such as publish department ids keep their original names.
type ResourceAuditFields struct {
	Name              string
	Table             string
	IDColumn          string
	DeptColumn        string
	PublishDeptColumn string
	CreateByColumn    string
	UpdateByColumn    string
	CreateDeptColumn  string
	UpdateDeptColumn  string
	CreateTimeColumn  string
	UpdateTimeColumn  string
}

func (f ResourceAuditFields) IDField() string {
	return quoteColumn(f.IDColumn)
}

func (f ResourceAuditFields) DeptField() string {
	return quoteColumn(f.DeptColumn)
}

func (f ResourceAuditFields) PublishDeptField() string {
	return quoteColumn(f.PublishDeptColumn)
}

func (f ResourceAuditFields) CreateByField() string {
	return quoteColumn(f.CreateByColumn)
}

func (f ResourceAuditFields) UpdateByField() string {
	return quoteColumn(f.UpdateByColumn)
}

func (f ResourceAuditFields) CreateDeptField() string {
	return quoteColumn(f.CreateDeptColumn)
}

func (f ResourceAuditFields) UpdateDeptField() string {
	return quoteColumn(f.UpdateDeptColumn)
}

func (f ResourceAuditFields) CreateTimeField() string {
	return quoteColumn(f.CreateTimeColumn)
}

func (f ResourceAuditFields) UpdateTimeField() string {
	return quoteColumn(f.UpdateTimeColumn)
}

func (f ResourceAuditFields) DataScopeDeptField() string {
	if f.DeptColumn != "" {
		return f.DeptField()
	}
	return f.CreateDeptField()
}

func (f ResourceAuditFields) DataScopeCreateByField() string {
	return f.CreateByField()
}

func (f ResourceAuditFields) HasDataScopeFields() bool {
	return f.DataScopeDeptField() != "" && f.DataScopeCreateByField() != ""
}

func quoteColumn(column string) string {
	column = strings.TrimSpace(strings.Trim(column, "`"))
	if column == "" {
		return ""
	}
	return "`" + column + "`"
}

var (
	UserAuditFields = ResourceAuditFields{
		Name:             "user",
		Table:            "users",
		IDColumn:         "id",
		CreateTimeColumn: "user_add_time",
		UpdateTimeColumn: "user_edit_time",
	}

	AdminLogAuditFields = ResourceAuditFields{
		Name:             "admin_log",
		Table:            "logs",
		IDColumn:         "id",
		CreateByColumn:   "create_by",
		UpdateByColumn:   "update_by",
		CreateDeptColumn: "create_dept_id",
		UpdateDeptColumn: "update_dept_id",
		CreateTimeColumn: "add_time",
		UpdateTimeColumn: "edit_time",
	}

	NewsAuditFields = ResourceAuditFields{
		Name:              "news",
		Table:             "news",
		IDColumn:          "id",
		PublishDeptColumn: "news_publish_dept_ids",
		CreateByColumn:    "create_by",
		UpdateByColumn:    "update_by",
		CreateDeptColumn:  "create_dept_id",
		UpdateDeptColumn:  "update_dept_id",
		CreateTimeColumn:  "add_time",
		UpdateTimeColumn:  "edit_time",
	}

	EnrollAuditFields = ResourceAuditFields{
		Name:              "enroll",
		Table:             "enrolls",
		IDColumn:          "id",
		PublishDeptColumn: "enroll_publish_dept_ids",
		CreateByColumn:    "create_by",
		UpdateByColumn:    "update_by",
		CreateDeptColumn:  "create_dept_id",
		UpdateDeptColumn:  "update_dept_id",
		CreateTimeColumn:  "add_time",
		UpdateTimeColumn:  "edit_time",
	}

	EventAuditFields = ResourceAuditFields{
		Name:              "event",
		Table:             "events",
		IDColumn:          "id",
		PublishDeptColumn: "event_publish_dept_ids",
		CreateByColumn:    "create_by",
		UpdateByColumn:    "update_by",
		CreateDeptColumn:  "create_dept_id",
		UpdateDeptColumn:  "update_dept_id",
		CreateTimeColumn:  "add_time",
		UpdateTimeColumn:  "edit_time",
	}

	SurveyAuditFields = ResourceAuditFields{
		Name:             "survey",
		Table:            "survey",
		IDColumn:         "survey_id",
		CreateByColumn:   "create_by",
		UpdateByColumn:   "update_by",
		CreateDeptColumn: "create_dept_id",
		UpdateDeptColumn: "update_dept_id",
		CreateTimeColumn: "add_time",
		UpdateTimeColumn: "edit_time",
	}

	ExamAuditFields = ResourceAuditFields{
		Name:              "exam",
		Table:             "exam",
		IDColumn:          "exam_id",
		PublishDeptColumn: "exam_publish_dept_ids",
		CreateByColumn:    "create_by",
		UpdateByColumn:    "update_by",
		CreateDeptColumn:  "create_dept_id",
		UpdateDeptColumn:  "update_dept_id",
		CreateTimeColumn:  "add_time",
		UpdateTimeColumn:  "edit_time",
	}

	SurveyQuestionAuditFields = ResourceAuditFields{
		Name:             "survey_question",
		Table:            "survey_question",
		IDColumn:         "survey_q_id",
		CreateByColumn:   "create_by",
		UpdateByColumn:   "update_by",
		CreateDeptColumn: "create_dept_id",
		UpdateDeptColumn: "update_dept_id",
		CreateTimeColumn: "add_time",
		UpdateTimeColumn: "edit_time",
	}

	ExamQuestionAuditFields = ResourceAuditFields{
		Name:             "exam_question",
		Table:            "exam_question",
		IDColumn:         "exam_q_id",
		CreateByColumn:   "create_by",
		UpdateByColumn:   "update_by",
		CreateDeptColumn: "create_dept_id",
		UpdateDeptColumn: "update_dept_id",
		CreateTimeColumn: "add_time",
		UpdateTimeColumn: "edit_time",
	}

	ExamPaperAuditFields = ResourceAuditFields{
		Name:             "exam_paper",
		Table:            "exam_paper",
		IDColumn:         "exam_p_id",
		CreateByColumn:   "create_by",
		UpdateByColumn:   "update_by",
		CreateDeptColumn: "create_dept_id",
		UpdateDeptColumn: "update_dept_id",
		CreateTimeColumn: "add_time",
		UpdateTimeColumn: "edit_time",
	}

	DingTalkH5PerfUserAuditFields = ResourceAuditFields{
		Name:             "dingtalk_h5_perf_user",
		Table:            "users",
		IDColumn:         "id",
		CreateTimeColumn: "user_add_time",
		UpdateTimeColumn: "user_edit_time",
	}

	DingTalkH5AuditFields         = dingtalkH5EmbeddedAuditFields("dingtalk_h5_embedded_audit", "")
	DingTalkH5ReviewAuditFields   = dingtalkH5EmbeddedAuditFields("dingtalk_h5_perf_review", "dingtalk_h5_perf_reviews")
	DingTalkH5HistoryAuditFields  = dingtalkH5EmbeddedAuditFields("dingtalk_h5_perf_history", "dingtalk_h5_perf_histories")
	DingTalkH5TemplateAuditFields = dingtalkH5EmbeddedAuditFields("dingtalk_h5_perf_template", "dingtalk_h5_perf_templates")
)

func dingtalkH5EmbeddedAuditFields(name, table string) ResourceAuditFields {
	return ResourceAuditFields{
		Name:             name,
		Table:            table,
		IDColumn:         "id",
		CreateByColumn:   "create_by",
		UpdateByColumn:   "update_by",
		CreateDeptColumn: "create_dept_id",
		UpdateDeptColumn: "update_dept_id",
		CreateTimeColumn: "add_time",
		UpdateTimeColumn: "edit_time",
	}
}
