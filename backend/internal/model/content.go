package model

import "time"

type News struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:新闻ID"`
	Title          string    `json:"title" gorm:"size:200;column:news_title;comment:新闻标题"`
	Desc           string    `json:"desc" gorm:"size:500;column:news_desc;comment:新闻简介"`
	Status         int       `json:"status" gorm:"default:1;column:news_status;comment:状态:1正常 0禁用"`
	DeptID         uint      `json:"deptId" gorm:"default:0;column:news_dept_id;comment:所属部门ID"`
	PublishDeptIds string    `json:"publishDeptIds" gorm:"size:500;column:news_publish_dept_ids;comment:发布部门ID列表,逗号分隔"`
	CreateBy       uint      `json:"createBy" gorm:"default:0;column:news_create_by;comment:创建管理员ID"`
	CateID         string    `json:"cateId" gorm:"size:50;column:news_cate_id;comment:分类ID"`
	CateName       string    `json:"cateName" gorm:"size:50;column:news_cate_name;comment:分类名称"`
	Order          int       `json:"order" gorm:"default:9999;column:news_order;comment:排序值"`
	Vouch          int       `json:"vouch" gorm:"default:0;column:news_vouch;comment:推荐标记:1推荐"`
	Content        string    `json:"content" gorm:"type:text;column:news_content;comment:新闻内容"`
	QR             string    `json:"qr" gorm:"size:500;column:news_qr;comment:二维码URL"`
	ViewCnt        int       `json:"viewCnt" gorm:"default:0;column:news_view_cnt;comment:浏览次数"`
	Pic            string    `json:"-" gorm:"type:text;column:news_pic;comment:图片列表JSON"`
	Img            string    `json:"img" gorm:"-"`
	Forms          string    `json:"forms" gorm:"type:text;column:news_forms;comment:扩展表单数据JSON"`
	Obj            string    `json:"obj" gorm:"type:text;column:news_obj;comment:扩展对象数据JSON"`
	AddTime        int64     `json:"_createTime" gorm:"column:news_add_time;comment:创建时间"`
	EditTime       int64     `json:"editTime" gorm:"column:news_edit_time;comment:修改时间"`
	AddIP          string    `json:"NEWS_ADD_IP" gorm:"size:50;column:news_add_ip;comment:创建IP"`
	EditIP         string    `json:"NEWS_EDIT_IP" gorm:"size:50;column:news_edit_ip;comment:修改IP"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (n News) GetCreateTime() string {
	if n.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(n.AddTime).Format("2006-01-02 15:04:05")
}

type Favorite struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:收藏ID"`
	UserID    string    `json:"userId" gorm:"size:200;column:fav_user_id;comment:用户openid"`
	Title     string    `json:"title" gorm:"size:200;column:fav_title;comment:收藏标题"`
	Type      string    `json:"type" gorm:"size:20;column:fav_type;comment:收藏类型"`
	OID       string    `json:"oid" gorm:"size:50;column:fav_oid;comment:关联对象ID"`
	Path      string    `json:"path" gorm:"size:500;column:fav_path;comment:收藏路径"`
	AddTime   int64     `json:"_createTime" gorm:"column:fav_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:fav_edit_time;comment:修改时间"`
	AddIP     string    `json:"FAV_ADD_IP" gorm:"size:50;column:fav_add_ip;comment:创建IP"`
	EditIP    string    `json:"FAV_EDIT_IP" gorm:"size:50;column:fav_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Notify struct {
	ID         uint   `gorm:"primaryKey;column:notify_id" json:"id"`
	Title      string `gorm:"column:notify_title;size:255" json:"title"`
	Content    string `gorm:"column:notify_content;type:text" json:"content"`
	Type       string `gorm:"column:notify_type;size:32;index" json:"type"`
	SourceID   string `gorm:"column:notify_source_id;size:64;index" json:"sourceId"`
	SourceType string `gorm:"column:notify_source_type;size:32;index" json:"sourceType"`
	UserID     string `gorm:"column:notify_user_id;size:128;index" json:"userId"`
	IsRead     int    `gorm:"column:notify_is_read;default:0" json:"isRead"`
	AddTime    int64  `gorm:"column:notify_add_time" json:"addTime"`
}

func (Notify) TableName() string { return "notify" }
