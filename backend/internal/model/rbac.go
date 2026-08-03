package model

import "time"

type Department struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:部门ID"`
	Name      string    `json:"name" gorm:"size:100;column:dept_name;comment:部门名称"`
	ParentID  uint      `json:"parentId" gorm:"default:0;column:dept_parent_id;comment:上级部门ID"`
	Sort      int       `json:"sort" gorm:"default:0;column:dept_sort;comment:排序"`
	Status    int       `json:"status" gorm:"default:1;column:dept_status;comment:状态:1正常 0禁用"`
	AddTime   int64     `json:"addTime" gorm:"column:dept_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:dept_edit_time;comment:修改时间"`
	AddIP     string    `json:"deptAddIp" gorm:"size:50;column:dept_add_ip;comment:创建IP"`
	EditIP    string    `json:"deptEditIp" gorm:"size:50;column:dept_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Children []*Department `json:"children" gorm:"-"`
}

type Role struct {
	ID              uint      `json:"id" gorm:"primaryKey;comment:角色ID"`
	Name            string    `json:"name" gorm:"size:100;column:role_name;comment:角色名称"`
	Remark          string    `json:"remark" gorm:"size:200;column:role_remark;comment:角色备注"`
	Sort            int       `json:"sort" gorm:"default:0;column:role_sort;comment:排序"`
	Status          int       `json:"status" gorm:"default:1;column:role_status;comment:状态:1正常 0禁用"`
	AllowAdminLogin int       `json:"allowAdminLogin" gorm:"default:1;column:role_allow_admin_login;comment:是否允许后台登录"`
	DataScope       int       `json:"dataScope" gorm:"default:1;column:role_data_scope;comment:数据范围:1全部 2本部门及子部门 3本人 4自定义部门"`
	AddTime         int64     `json:"addTime" gorm:"column:role_add_time;comment:创建时间"`
	EditTime        int64     `json:"editTime" gorm:"column:role_edit_time;comment:修改时间"`
	AddIP           string    `json:"addIp" gorm:"size:50;column:role_add_ip;comment:创建IP"`
	EditIP          string    `json:"editIp" gorm:"size:50;column:role_edit_ip;comment:修改IP"`
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}
