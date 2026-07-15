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
	ID        uint      `json:"id" gorm:"primaryKey;comment:角色ID"`
	Name      string    `json:"name" gorm:"size:100;column:role_name;comment:角色名称"`
	Remark    string    `json:"remark" gorm:"size:200;column:role_remark;comment:角色备注"`
	Sort      int       `json:"sort" gorm:"default:0;column:role_sort;comment:排序"`
	Status    int       `json:"status" gorm:"default:1;column:role_status;comment:状态:1正常 0禁用"`
	DataScope int       `json:"dataScope" gorm:"default:1;column:role_data_scope;comment:数据范围:1全部 2本部门 3本人"`
	AddTime   int64     `json:"addTime" gorm:"column:role_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:role_edit_time;comment:修改时间"`
	AddIP     string    `json:"addIp" gorm:"size:50;column:role_add_ip;comment:创建IP"`
	EditIP    string    `json:"editIp" gorm:"size:50;column:role_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Menu struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:菜单ID"`
	Name      string    `json:"name" gorm:"size:100;column:menu_name;comment:菜单名称"`
	ParentID  uint      `json:"parentId" gorm:"default:0;column:menu_parent_id;comment:上级菜单ID"`
	Path      string    `json:"path" gorm:"size:200;column:menu_path;comment:路由路径"`
	Perms     string    `json:"perms" gorm:"size:200;column:menu_perms;comment:权限标识(多个逗号分隔)"`
	Icon      string    `json:"icon" gorm:"size:100;column:menu_icon;comment:图标"`
	Sort      int       `json:"sort" gorm:"default:0;column:menu_sort;comment:排序"`
	Status    int       `json:"status" gorm:"default:1;column:menu_status;comment:状态:1正常 0禁用"`
	Type      int       `json:"type" gorm:"default:1;column:menu_type;comment:类型:0目录 1菜单 2按钮"`
	AddTime   int64     `json:"addTime" gorm:"column:menu_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:menu_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Children  []*Menu   `json:"children" gorm:"-"`
}

type RoleMenu struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	RoleID    uint      `json:"roleId" gorm:"index;column:role_menu_role_id;comment:角色ID"`
	MenuID    uint      `json:"menuId" gorm:"index;column:role_menu_menu_id;comment:菜单ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type RoleDept struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	RoleID    uint      `json:"roleId" gorm:"index;column:role_dept_role_id;comment:角色ID"`
	DeptID    uint      `json:"deptId" gorm:"index;column:role_dept_dept_id;comment:部门ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
