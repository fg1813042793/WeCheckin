package organization

import "time"

type Position struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:岗位ID"`
	Name      string    `json:"name" gorm:"size:100;column:position_name;comment:岗位名称"`
	Sort      int       `json:"sort" gorm:"default:0;column:position_sort;comment:排序"`
	Status    int       `json:"status" gorm:"default:1;column:position_status;comment:状态:1正常 0禁用"`
	AddTime   int64     `json:"addTime" gorm:"column:position_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:position_edit_time;comment:修改时间"`
	AddIP     string    `json:"addIp" gorm:"size:50;column:position_add_ip;comment:创建IP"`
	EditIP    string    `json:"editIp" gorm:"size:50;column:position_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
