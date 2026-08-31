package workflowmodel

import "time"

const (
	OrgApproverIdentityStatusDisabled = 0
	OrgApproverIdentityStatusEnabled  = 1
	OrgApproverAssignmentStatusOff    = 0
	OrgApproverAssignmentStatusOn     = 1
)

type OrgApproverIdentity struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:组织审批身份ID"`
	Code      string    `json:"code" gorm:"size:80;uniqueIndex;column:identity_code;comment:身份编码"`
	Name      string    `json:"name" gorm:"size:100;column:identity_name;comment:身份名称"`
	Sort      int       `json:"sort" gorm:"default:0;column:identity_sort;comment:排序"`
	Status    int       `json:"status" gorm:"default:1;column:identity_status;comment:状态:1启用 0停用"`
	AddTime   int64     `json:"addTime" gorm:"column:identity_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:identity_edit_time;comment:更新时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (OrgApproverIdentity) TableName() string { return "workflow_org_approver_identities" }

type OrgApproverAssignment struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:组织审批身份人员ID"`
	DepartmentID uint      `json:"departmentId" gorm:"index;default:0;column:department_id;comment:部门ID"`
	IdentityCode string    `json:"identityCode" gorm:"size:80;index;column:identity_code;comment:身份编码"`
	UserID       uint      `json:"userId" gorm:"index;default:0;column:user_id;comment:用户ID"`
	Sort         int       `json:"sort" gorm:"default:0;column:assignment_sort;comment:审批顺序"`
	Status       int       `json:"status" gorm:"default:1;column:assignment_status;comment:状态:1启用 0停用"`
	AddTime      int64     `json:"addTime" gorm:"column:assignment_add_time;comment:创建时间"`
	EditTime     int64     `json:"editTime" gorm:"column:assignment_edit_time;comment:更新时间"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (OrgApproverAssignment) TableName() string { return "workflow_org_approver_assignments" }
