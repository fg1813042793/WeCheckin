package model

import "time"

type DingTalkH5AuditFields struct {
	CreateBy     uint  `json:"createBy" gorm:"default:0;index;column:create_by;comment:创建人ID"`
	UpdateBy     uint  `json:"updateBy" gorm:"default:0;index;column:update_by;comment:更新人ID"`
	CreateDeptID uint  `json:"createDeptId" gorm:"default:0;index;column:create_dept_id;comment:创建人部门ID"`
	UpdateDeptID uint  `json:"updateDeptId" gorm:"default:0;index;column:update_dept_id;comment:更新人部门ID"`
	DeleteBy     uint  `json:"deleteBy" gorm:"default:0;index;column:delete_by;comment:删除人ID"`
	DeleteDeptID uint  `json:"deleteDeptId" gorm:"default:0;index;column:delete_dept_id;comment:删除人部门ID"`
	DeletedAt    int64 `json:"deletedAt" gorm:"default:0;index;column:deleted_at;comment:软删除时间"`
}

type DingTalkH5PerfUser struct {
	ID                     uint      `json:"id" gorm:"primaryKey;comment:用户ID"`
	Account                string    `json:"account" gorm:"uniqueIndex;size:200;column:user_mini_openid;comment:用户账号"`
	Name                   string    `json:"name" gorm:"size:100;index;column:user_name;comment:姓名"`
	Password               string    `json:"-" gorm:"size:100;column:user_password;comment:密码哈希"`
	Pic                    string    `json:"avatar" gorm:"size:500;column:user_pic;comment:头像URL"`
	Status                 int       `json:"status" gorm:"default:1;index;column:user_status;comment:状态:1正常 0禁用"`
	RoleID                 uint      `json:"roleId" gorm:"column:user_role_id;comment:角色ID"`
	RoleIDs                []uint    `json:"roleIds" gorm:"-"`
	PositionID             uint      `json:"positionId" gorm:"column:user_position_id;comment:岗位ID"`
	Obj                    string    `json:"-" gorm:"type:text;column:user_obj;comment:扩展对象数据JSON"`
	AddTime                int64     `json:"addTime" gorm:"column:user_add_time;comment:创建时间"`
	EditTime               int64     `json:"editTime" gorm:"column:user_edit_time;comment:修改时间"`
	Role                   string    `json:"role" gorm:"-"`
	Position               string    `json:"position" gorm:"-"`
	Department             string    `json:"department" gorm:"-"`
	DepartmentLevel1       string    `json:"departmentLevel1" gorm:"-"`
	DepartmentLevel2       string    `json:"departmentLevel2" gorm:"-"`
	DepartmentLevel3       string    `json:"departmentLevel3" gorm:"-"`
	ManagerAccount         string    `json:"managerId" gorm:"-"`
	HRBPAccount            string    `json:"hrbpId" gorm:"-"`
	ResponsibleDepartments string    `json:"responsibleDepartments" gorm:"-"`
	CreatedAt              time.Time `json:"-"`
	UpdatedAt              time.Time `json:"-"`
}

func (DingTalkH5PerfUser) TableName() string { return "users" }

type DingTalkH5CorpConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:配置ID"`
	CorpID    string    `json:"corpId" gorm:"uniqueIndex:uk_dt_h5_corp_id;size:120;column:corp_id;comment:钉钉企业CorpId"`
	CorpName  string    `json:"corpName" gorm:"size:120;column:corp_name;comment:企业名称"`
	AppKey    string    `json:"appKey" gorm:"size:160;column:app_key;comment:钉钉内部应用AppKey"`
	AppSecret string    `json:"-" gorm:"type:text;column:app_secret;comment:钉钉内部应用AppSecret"`
	AgentID   string    `json:"agentId" gorm:"size:80;column:agent_id;comment:钉钉内部应用AgentId"`
	Enabled   int       `json:"enabled" gorm:"default:1;index;column:enabled;comment:是否启用"`
	AddTime   int64     `json:"addTime" gorm:"column:add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (DingTalkH5CorpConfig) TableName() string { return "dingtalk_h5_corp_configs" }

type DingTalkH5UserBinding struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:绑定ID"`
	CorpID         string    `json:"corpId" gorm:"uniqueIndex:uk_dt_h5_binding_corp_user,priority:1;size:120;column:corp_id;comment:钉钉企业CorpId"`
	DingTalkUserID string    `json:"dingTalkUserId" gorm:"uniqueIndex:uk_dt_h5_binding_corp_user,priority:2;size:160;column:dingtalk_user_id;comment:钉钉用户UserId"`
	UnionID        string    `json:"unionId" gorm:"size:160;index;column:union_id;comment:钉钉UnionId"`
	UserID         uint      `json:"userId" gorm:"index;column:user_id;comment:本地用户ID"`
	Enabled        int       `json:"enabled" gorm:"default:1;index;column:enabled;comment:是否启用"`
	AddTime        int64     `json:"addTime" gorm:"column:add_time;comment:创建时间"`
	EditTime       int64     `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (DingTalkH5UserBinding) TableName() string { return "dingtalk_h5_user_bindings" }

type DingTalkH5PerfReview struct {
	ID                      uint   `json:"id" gorm:"primaryKey;comment:考评单ID"`
	ReviewNo                string `json:"reviewNo" gorm:"size:120;index;column:review_no;comment:考评单编号"`
	EmployeeAccount         string `json:"employeeId" gorm:"size:80;index:idx_dt_h5_review_employee_period,priority:1;column:employee_account;comment:员工账号"`
	ManagerAccount          string `json:"managerId" gorm:"size:80;index;column:manager_account;comment:直属上级账号"`
	HRBPAccount             string `json:"hrbpId" gorm:"size:80;index;column:hrbp_account;comment:HRBP账号"`
	HRBPReviewerAccount     string `json:"hrbpReviewerId" gorm:"size:80;index;column:hrbp_reviewer_account;comment:实际HRBP处理人账号"`
	Department              string `json:"department" gorm:"size:200;index;column:department;comment:部门全称"`
	DepartmentLevel1        string `json:"departmentLevel1" gorm:"size:100;index;column:department_level1;comment:一级部门"`
	DepartmentLevel2        string `json:"departmentLevel2" gorm:"size:100;index;column:department_level2;comment:二级部门"`
	DepartmentLevel3        string `json:"departmentLevel3" gorm:"size:100;column:department_level3;comment:三级部门"`
	Period                  string `json:"period" gorm:"size:20;index:idx_dt_h5_review_employee_period,priority:2;index;column:period;comment:考评月份"`
	NextPeriod              string `json:"nextPeriod" gorm:"size:20;index;column:next_period;comment:目标月份"`
	Status                  string `json:"status" gorm:"size:30;index;column:status;comment:流程状态"`
	ObjectiveSourceReviewNo string `json:"objectiveSourceReviewId" gorm:"size:120;column:objective_source_review_no;comment:目标来源考评单"`
	ObjectiveSourcePeriod   string `json:"objectiveSourcePeriod" gorm:"size:20;column:objective_source_period;comment:目标来源月份"`
	ObjectivesJSON          string `json:"objectives" gorm:"type:mediumtext;column:objectives_json;comment:本月目标JSON"`
	NextObjectivesJSON      string `json:"nextObjectives" gorm:"type:mediumtext;column:next_objectives_json;comment:下月目标JSON"`
	ValuesJSON              string `json:"values" gorm:"type:mediumtext;column:values_json;comment:价值观评分JSON"`
	SelfSummary             string `json:"selfSummary" gorm:"type:text;column:self_summary;comment:员工总结"`
	ManagerComment          string `json:"managerComment" gorm:"type:text;column:manager_comment;comment:上级评价"`
	ManagerGrade            string `json:"managerGrade" gorm:"size:20;index;column:manager_grade;comment:上级分档"`
	HRBPComment             string `json:"hrbpComment" gorm:"type:text;column:hrbp_comment;comment:HRBP评价"`
	HRBPGrade               string `json:"hrbpGrade" gorm:"size:20;index;column:hrbp_grade;comment:HRBP分档"`
	EmployeeConfirmResult   string `json:"employeeConfirmResult" gorm:"size:30;column:employee_confirm_result;comment:员工确认结果"`
	EmployeeConfirmComment  string `json:"employeeConfirmComment" gorm:"type:text;column:employee_confirm_comment;comment:确认意见或异议原因"`
	EmployeeConfirmedAt     int64  `json:"employeeConfirmedAt" gorm:"column:employee_confirmed_at;comment:确认时间"`
	FinalGrade              string `json:"finalGrade" gorm:"size:20;index;column:final_grade;comment:最终分档"`
	FinalNote               string `json:"finalNote" gorm:"type:text;column:final_note;comment:HRBP备注"`
	AddTime                 int64  `json:"addTime" gorm:"column:add_time;comment:创建时间"`
	EditTime                int64  `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	DingTalkH5AuditFields   `gorm:"embedded"`
	CreatedAt               time.Time `json:"-"`
	UpdatedAt               time.Time `json:"-"`
}

func (DingTalkH5PerfReview) TableName() string { return "dingtalk_h5_perf_reviews" }

type DingTalkH5PerfHistory struct {
	ID                    uint   `json:"id" gorm:"primaryKey;comment:流转记录ID"`
	ReviewID              uint   `json:"reviewId" gorm:"index;column:review_id;comment:考评单ID"`
	ReviewNo              string `json:"reviewNo" gorm:"size:120;index;column:review_no;comment:考评单编号"`
	ByAccount             string `json:"byAccount" gorm:"size:80;column:by_account;comment:操作人账号"`
	ByName                string `json:"byName" gorm:"size:100;column:by_name;comment:操作人姓名"`
	Action                string `json:"action" gorm:"size:500;column:action;comment:操作内容"`
	AddTime               int64  `json:"addTime" gorm:"index;column:add_time;comment:创建时间"`
	EditTime              int64  `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	DingTalkH5AuditFields `gorm:"embedded"`
	CreatedAt             time.Time `json:"-"`
	UpdatedAt             time.Time `json:"-"`
}

func (DingTalkH5PerfHistory) TableName() string { return "dingtalk_h5_perf_histories" }

type DingTalkH5PerfTemplate struct {
	ID                    uint   `json:"id" gorm:"primaryKey;comment:模板ID"`
	Key                   string `json:"key" gorm:"uniqueIndex;size:80;column:template_key;comment:模板键"`
	Payload               string `json:"payload" gorm:"type:mediumtext;column:payload;comment:模板JSON"`
	AddTime               int64  `json:"addTime" gorm:"column:add_time;comment:创建时间"`
	EditTime              int64  `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	DingTalkH5AuditFields `gorm:"embedded"`
	CreatedAt             time.Time `json:"-"`
	UpdatedAt             time.Time `json:"-"`
}

func (DingTalkH5PerfTemplate) TableName() string { return "dingtalk_h5_perf_templates" }
