package model

import "time"

type Permission struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:权限ID"`
	Key          string    `json:"key" gorm:"uniqueIndex;size:160;column:permission_key;comment:权限编码"`
	Name         string    `json:"name" gorm:"size:120;column:permission_name;comment:权限名称"`
	Platform     string    `json:"platform" gorm:"index;size:40;column:permission_platform;comment:平台"`
	Type         string    `json:"type" gorm:"index;size:40;column:permission_type;comment:权限类型"`
	ParentKey    string    `json:"parentKey" gorm:"size:160;column:permission_parent_key;comment:父权限编码"`
	ResourceID   uint      `json:"resourceId" gorm:"index;column:permission_resource_id;comment:旧资源ID"`
	ResourcePath string    `json:"resourcePath" gorm:"size:240;column:permission_resource_path;comment:资源路径"`
	Icon         string    `json:"icon" gorm:"size:100;column:permission_icon;comment:图标"`
	Perms        string    `json:"perms" gorm:"size:240;column:permission_perms;comment:兼容权限标识"`
	Sort         int       `json:"sort" gorm:"default:0;column:permission_sort;comment:排序"`
	Status       int       `json:"status" gorm:"default:1;column:permission_status;comment:状态:1启用 0停用"`
	AddTime      int64     `json:"addTime" gorm:"column:permission_add_time;comment:创建时间"`
	EditTime     int64     `json:"editTime" gorm:"column:permission_edit_time;comment:修改时间"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (Permission) TableName() string { return "permissions" }

type PermissionGrant struct {
	ID            uint      `json:"id" gorm:"primaryKey;comment:授权ID"`
	SubjectType   string    `json:"subjectType" gorm:"uniqueIndex:idx_permission_grants_subject_permission,priority:1;index;size:20;column:grant_subject_type;comment:授权主体类型"`
	SubjectID     uint      `json:"subjectId" gorm:"uniqueIndex:idx_permission_grants_subject_permission,priority:2;index;column:grant_subject_id;comment:授权主体ID"`
	PermissionKey string    `json:"permissionKey" gorm:"uniqueIndex:idx_permission_grants_subject_permission,priority:3;index;size:160;column:grant_permission_key;comment:权限编码"`
	PermissionID  uint      `json:"permissionId" gorm:"index;column:grant_permission_id;comment:权限ID"`
	Effect        string    `json:"effect" gorm:"size:20;default:allow;column:grant_effect;comment:授权效果"`
	ScopeValue    string    `json:"scopeValue" gorm:"type:text;column:grant_scope_value;comment:范围JSON"`
	Source        string    `json:"source" gorm:"size:40;column:grant_source;comment:授权来源"`
	Status        int       `json:"status" gorm:"default:1;column:grant_status;comment:状态:1启用 0停用"`
	AddTime       int64     `json:"addTime" gorm:"column:grant_add_time;comment:创建时间"`
	EditTime      int64     `json:"editTime" gorm:"column:grant_edit_time;comment:修改时间"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

func (PermissionGrant) TableName() string { return "permission_grants" }
