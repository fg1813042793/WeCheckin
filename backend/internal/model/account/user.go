package account

import "time"

const (
	ReportingRelationTypeDirect = "direct"
	ReportingRelationTypeDotted = "dotted"

	ReportingRelationStatusOff = 0
	ReportingRelationStatusOn  = 1
)

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:用户ID"`
	MiniOpenID     string    `json:"miniOpenID" gorm:"uniqueIndex;size:200;column:user_mini_openid;comment:微信小程序openid"`
	Status         int       `json:"status" gorm:"default:1;column:user_status;comment:状态:1正常 0禁用"`
	CheckReason    string    `json:"checkReason" gorm:"size:500;column:user_check_reason;comment:审核原因"`
	Account        string    `json:"account" gorm:"size:100;index;column:user_account;comment:后台登录账号"`
	Name           string    `json:"name" gorm:"size:100;column:user_name;comment:用户昵称"`
	Mobile         string    `json:"mobile" gorm:"size:20;column:user_mobile;comment:手机号"`
	PositionID     uint      `json:"positionId" gorm:"index;default:0;column:user_position_id;comment:岗位ID"`
	Pic            string    `json:"avatar" gorm:"size:500;column:user_pic;comment:头像URL"`
	Forms          string    `json:"forms" gorm:"type:text;column:user_forms;comment:扩展表单数据JSON"`
	Obj            string    `json:"obj" gorm:"type:text;column:user_obj;comment:扩展对象数据JSON"`
	Password       string    `json:"-" gorm:"size:100;column:user_password;comment:密码(md5 hex)"`
	AdminEnabled   int       `json:"adminEnabled" gorm:"default:0;column:user_admin_enabled;comment:是否后台账号"`
	AdminType      int       `json:"adminType" gorm:"default:0;column:user_admin_type;comment:后台账号类型:1超级管理员"`
	RoleID         uint      `json:"roleId" gorm:"default:0;column:user_role_id;comment:后台角色ID"`
	AdminDesc      string    `json:"adminDesc" gorm:"size:200;column:user_admin_desc;comment:管理员描述"`
	AdminToken     string    `json:"-" gorm:"size:100;column:user_admin_token;comment:后台登录token"`
	AdminTokenTime int64     `json:"-" gorm:"column:user_admin_token_time;comment:后台token生成时间"`
	LoginCnt       int       `json:"loginCnt" gorm:"default:0;column:user_login_cnt;comment:登录次数"`
	LoginTime      int64     `json:"loginTime" gorm:"column:user_login_time;comment:最后登录时间"`
	AddTime        int64     `json:"addTime" gorm:"column:user_add_time;comment:创建时间"`
	AddIP          string    `json:"addIP" gorm:"column:user_add_ip;comment:创建IP"`
	EditTime       int64     `json:"editTime" gorm:"column:user_edit_time;comment:修改时间"`
	EditIP         string    `json:"editIP" gorm:"column:user_edit_ip;comment:修改IP"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`

	Role            string   `json:"role" gorm:"-"`
	RoleIDs         []uint   `json:"roleIds" gorm:"-"`
	RoleNames       []string `json:"roleNames" gorm:"-"`
	DeptName        string   `json:"deptName" gorm:"-"`
	TopDeptName     string   `json:"topDeptName" gorm:"-"`
	PositionName    string   `json:"positionName" gorm:"-"`
	ManagerUserID   uint     `json:"managerUserId" gorm:"-"`
	ManagerUserName string   `json:"managerUserName" gorm:"-"`
}

func (u *User) GetRole() string {
	if u.Status == 9 || u.Status == 1 {
		return "admin"
	}
	return "user"
}

func (u User) GetCreateTime() string {
	if u.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(u.AddTime).Format("2006-01-02 15:04:05")
}

type UserDept struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	UserID    uint      `json:"userId" gorm:"index;column:user_dept_user_id;comment:用户ID"`
	DeptID    uint      `json:"deptId" gorm:"index;column:user_dept_dept_id;comment:部门ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type UserReportingRelation struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:汇报关系ID"`
	EmployeeUserID uint      `json:"employeeUserId" gorm:"index;column:employee_user_id;comment:员工用户ID"`
	ManagerUserID  uint      `json:"managerUserId" gorm:"index;column:manager_user_id;comment:上级用户ID"`
	RelationType   string    `json:"relationType" gorm:"size:40;index;column:relation_type;comment:关系类型:direct直属 dotted虚线"`
	IsPrimary      int       `json:"isPrimary" gorm:"default:1;column:is_primary;comment:是否主关系"`
	Sort           int       `json:"sort" gorm:"default:0;column:relation_sort;comment:排序"`
	Status         int       `json:"status" gorm:"default:1;index;column:relation_status;comment:状态:1启用 0停用"`
	EffectiveFrom  int64     `json:"effectiveFrom" gorm:"default:0;index;column:effective_from;comment:生效时间"`
	EffectiveTo    int64     `json:"effectiveTo" gorm:"default:0;index;column:effective_to;comment:失效时间,0长期有效"`
	AddTime        int64     `json:"addTime" gorm:"column:relation_add_time;comment:创建时间"`
	EditTime       int64     `json:"editTime" gorm:"column:relation_edit_time;comment:更新时间"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (UserReportingRelation) TableName() string { return "user_reporting_relations" }

type UserRole struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	UserID    uint      `json:"userId" gorm:"uniqueIndex:uk_user_roles_user_role,priority:1;index;column:user_role_user_id;comment:用户ID"`
	RoleID    uint      `json:"roleId" gorm:"uniqueIndex:uk_user_roles_user_role,priority:2;index;column:user_role_role_id;comment:角色ID"`
	IsPrimary int       `json:"isPrimary" gorm:"default:0;index;column:user_role_is_primary;comment:是否主角色"`
	Status    int       `json:"status" gorm:"default:1;index;column:user_role_status;comment:状态:1启用 0停用"`
	Source    string    `json:"source" gorm:"size:40;column:user_role_source;comment:来源"`
	AddTime   int64     `json:"addTime" gorm:"column:user_role_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:user_role_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (UserRole) TableName() string { return "user_roles" }

type UserFormField struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:字段ID"`
	Label     string    `json:"label" gorm:"size:100;column:field_label;comment:字段名称"`
	Type      string    `json:"type" gorm:"size:20;column:field_type;comment:字段类型(文本/数字/多行文本/选择/图片/定位)"`
	Required  int       `json:"required" gorm:"default:0;column:field_required;comment:是否必填"`
	Options   string    `json:"options" gorm:"size:500;column:field_options;comment:选项(逗号分隔)"`
	Sort      int       `json:"sort" gorm:"default:0;column:field_sort;comment:排序"`
	AddTime   int64     `json:"addTime" gorm:"column:field_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:field_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
