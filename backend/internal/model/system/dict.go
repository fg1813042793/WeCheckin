package system

import "time"

type SysDictType struct {
	TypeCode  string    `json:"typeCode" gorm:"primaryKey;size:50;column:dict_type_code;comment:字典类型编码"`
	TypeName  string    `json:"typeName" gorm:"size:100;not null;column:dict_type_name;comment:字典类型名称"`
	Status    int       `json:"status" gorm:"not null;default:1;column:dict_type_status;index;comment:状态(1正常 0停用)"`
	Remark    string    `json:"remark" gorm:"size:500;not null;default:'';column:dict_type_remark;comment:备注"`
	AddTime   int64     `json:"addTime" gorm:"not null;default:0;column:dict_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"not null;default:0;column:dict_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (SysDictType) TableName() string { return "sys_dict_types" }

type SysDict struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:字典ID"`
	TypeCode  string    `json:"typeCode" gorm:"size:50;column:dict_type_code;index;comment:字典类型编码"`
	TypeName  string    `json:"typeName" gorm:"size:100;column:dict_type_name;comment:字典类型名称"`
	Label     string    `json:"label" gorm:"size:100;column:dict_label;comment:字典标签"`
	Value     string    `json:"value" gorm:"size:200;column:dict_value;comment:字典值"`
	Sort      int       `json:"sort" gorm:"default:0;column:dict_sort;comment:排序"`
	Status    int       `json:"status" gorm:"default:1;column:dict_status;comment:状态(1正常 0停用)"`
	Remark    string    `json:"remark" gorm:"size:500;column:dict_remark;comment:备注"`
	AddTime   int64     `json:"addTime" gorm:"column:dict_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:dict_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (SysDict) TableName() string { return "sys_dicts" }
