package model

import "time"

type Admin struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:管理员ID"`
	Name      string    `json:"name" gorm:"uniqueIndex;size:100;column:admin_name;comment:管理员用户名"`
	Password  string    `json:"-" gorm:"size:100;column:admin_password;comment:密码(sha256 hex)"`
	Desc      string    `json:"desc" gorm:"size:200;column:admin_desc;comment:管理员描述"`
	Pic       string    `json:"pic" gorm:"size:500;column:admin_pic;comment:头像URL"`
	Phone     string    `json:"phone" gorm:"size:20;column:admin_phone;comment:手机号"`
	Status    int       `json:"status" gorm:"default:1;column:admin_status;comment:状态:1正常 0禁用"`
	Type      int       `json:"type" gorm:"default:0;column:admin_type;comment:类型:1超级管理员"`
	RoleID    uint      `json:"roleId" gorm:"default:0;column:admin_role_id;comment:角色ID"`
	Token     string    `json:"token" gorm:"size:100;column:admin_token;comment:登录token"`
	TokenTime int64     `json:"tokenTime" gorm:"column:admin_token_time;comment:token生成时间"`
	LoginCnt  int       `json:"loginCnt" gorm:"default:0;column:admin_login_cnt;comment:登录次数"`
	LoginTime int64     `json:"loginTime" gorm:"column:admin_login_time;comment:最后登录时间"`
	AddTime   int64     `json:"_createTime" gorm:"column:admin_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:admin_edit_time;comment:修改时间"`
	AddIP     string    `json:"ADMIN_ADD_IP" gorm:"size:50;column:admin_add_ip;comment:创建IP"`
	EditIP    string    `json:"ADMIN_EDIT_IP" gorm:"size:50;column:admin_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Log struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:日志ID"`
	Type      int       `json:"type" gorm:"column:log_type;comment:日志类型"`
	Content   string    `json:"content" gorm:"type:text;column:log_content;comment:日志内容"`
	AdminID   string    `json:"adminId" gorm:"size:50;column:log_admin_id;comment:管理员ID"`
	AdminName string    `json:"adminName" gorm:"size:100;column:log_admin_name;comment:管理员用户名"`
	AdminDesc string    `json:"adminDesc" gorm:"size:200;column:log_admin_desc;comment:管理员描述"`
	AddTime   int64     `json:"_createTime" gorm:"column:log_add_time;comment:创建时间"`
	AddIP     string    `json:"LOG_ADD_IP" gorm:"size:50;column:log_add_ip;comment:创建IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type AdminDept struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	AdminID   uint      `json:"adminId" gorm:"index;column:admin_dept_admin_id;comment:管理员ID"`
	DeptID    uint      `json:"deptId" gorm:"index;column:admin_dept_dept_id;comment:部门ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
