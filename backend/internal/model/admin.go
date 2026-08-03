package model

import "time"

type Admin struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:用户ID"`
	MiniOpenID string    `json:"-" gorm:"size:200;column:user_mini_openid;comment:账号唯一标识"`
	Name       string    `json:"name" gorm:"size:100;index;column:user_name;comment:用户姓名"`
	Password   string    `json:"-" gorm:"size:100;column:user_password;comment:密码(sha256 hex)"`
	Desc       string    `json:"desc" gorm:"size:200;column:user_admin_desc;comment:管理员描述"`
	Pic        string    `json:"pic" gorm:"size:500;column:user_pic;comment:头像URL"`
	Phone      string    `json:"phone" gorm:"size:20;column:user_mobile;comment:手机号"`
	Status     int       `json:"status" gorm:"default:1;column:user_status;comment:状态:1正常 0禁用"`
	Type       int       `json:"type" gorm:"default:0;column:user_admin_type;comment:类型:1超级管理员"`
	RoleID     uint      `json:"roleId" gorm:"default:0;column:user_role_id;comment:角色ID"`
	RoleIDs    []uint    `json:"roleIds" gorm:"-"`
	RoleNames  []string  `json:"roleNames" gorm:"-"`
	Token      string    `json:"token" gorm:"size:100;column:user_admin_token;comment:登录token"`
	TokenTime  int64     `json:"tokenTime" gorm:"column:user_admin_token_time;comment:token生成时间"`
	LoginCnt   int       `json:"loginCnt" gorm:"default:0;column:user_login_cnt;comment:登录次数"`
	LoginTime  int64     `json:"loginTime" gorm:"column:user_login_time;comment:最后登录时间"`
	AddTime    int64     `json:"_createTime" gorm:"column:user_add_time;comment:创建时间"`
	EditTime   int64     `json:"editTime" gorm:"column:user_edit_time;comment:修改时间"`
	AddIP      string    `json:"ADMIN_ADD_IP" gorm:"size:50;column:user_add_ip;comment:创建IP"`
	EditIP     string    `json:"ADMIN_EDIT_IP" gorm:"size:50;column:user_edit_ip;comment:修改IP"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (Admin) TableName() string { return "users" }

type Log struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:日志ID"`
	Type         int       `json:"type" gorm:"column:log_type;comment:日志类型"`
	Content      string    `json:"content" gorm:"type:text;column:log_content;comment:日志内容"`
	AdminID      uint      `json:"adminId" gorm:"default:0;column:create_by;comment:创建人ID"`
	UpdateBy     uint      `json:"updateBy" gorm:"default:0;column:update_by;comment:更新人ID"`
	DeptID       uint      `json:"deptId" gorm:"default:0;column:create_dept_id;comment:创建人部门ID"`
	UpdateDeptID uint      `json:"updateDeptId" gorm:"default:0;column:update_dept_id;comment:更新人部门ID"`
	AdminName    string    `json:"adminName" gorm:"size:100;column:log_admin_name;comment:管理员用户名"`
	AdminDesc    string    `json:"adminDesc" gorm:"size:200;column:log_admin_desc;comment:管理员描述"`
	AddTime      int64     `json:"_createTime" gorm:"column:add_time;comment:创建时间"`
	EditTime     int64     `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	AddIP        string    `json:"LOG_ADD_IP" gorm:"size:50;column:log_add_ip;comment:创建IP"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}
