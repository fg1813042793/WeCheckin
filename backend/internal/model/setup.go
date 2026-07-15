package model

import "time"

type Setup struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:配置ID"`
	Key       string    `json:"key" gorm:"uniqueIndex;size:100;column:setup_key;comment:配置键名"`
	Value     string    `json:"value" gorm:"type:text;column:setup_value;comment:配置值"`
	Type      string    `json:"setup_type" gorm:"size:20;column:setup_type;comment:配置类型"`
	AddTime   int64     `json:"setup_add_time" gorm:"column:setup_add_time;comment:创建时间"`
	EditTime  int64     `json:"edit_time" gorm:"column:setup_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
