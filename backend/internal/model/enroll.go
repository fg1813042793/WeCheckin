package model

import "time"

type Enroll struct {
	ID             uint                     `json:"id" gorm:"primaryKey;comment:打卡项目ID"`
	Title          string                   `json:"title" gorm:"size:200;column:enroll_title;comment:打卡标题"`
	Desc           string                   `json:"desc" gorm:"-"`
	Img            string                   `json:"img" gorm:"-"`
	Status         int                      `json:"status" gorm:"default:1;column:enroll_status;comment:状态:1正常 0停用"`
	DeptID         uint                     `json:"deptId" gorm:"default:0;column:create_dept_id;comment:创建人部门ID"`
	PublishDeptIds string                   `json:"publishDeptIds" gorm:"size:500;column:enroll_publish_dept_ids;comment:发布部门ID列表,逗号分隔"`
	CreateBy       uint                     `json:"createBy" gorm:"default:0;column:create_by;comment:创建人ID"`
	UpdateBy       uint                     `json:"updateBy" gorm:"default:0;column:update_by;comment:更新人ID"`
	UpdateDeptID   uint                     `json:"updateDeptId" gorm:"default:0;column:update_dept_id;comment:更新人部门ID"`
	CateID         string                   `json:"cateId" gorm:"size:50;column:enroll_cate_id;comment:分类ID"`
	CateName       string                   `json:"cateName" gorm:"size:50;column:enroll_cate_name;comment:分类名称"`
	Start          int64                    `json:"timeStart" gorm:"column:enroll_start;comment:开始时间"`
	End            int64                    `json:"timeEnd" gorm:"column:enroll_end;comment:结束时间"`
	DayCnt         int                      `json:"dayCnt" gorm:"column:enroll_day_cnt;comment:打卡天数"`
	Order          int                      `json:"order" gorm:"default:9999;column:enroll_order;comment:排序值"`
	Vouch          int                      `json:"vouch" gorm:"default:0;column:enroll_vouch;comment:推荐标记:1推荐"`
	AllowRepeat    bool                     `json:"allowRepeat" gorm:"default:0;column:enroll_repeat;comment:允许重复打卡"`
	DailyLimit     int                      `json:"dailyLimit" gorm:"default:1;column:enroll_limit;comment:每日打卡次数限制"`
	Forms          string                   `json:"forms" gorm:"type:text;column:enroll_forms;comment:打卡表单字段定义JSON"`
	Obj            string                   `json:"obj" gorm:"type:text;column:enroll_obj;comment:扩展对象数据JSON(封面/描述)"`
	JoinForms      string                   `json:"joinForms" gorm:"type:text;column:enroll_join_forms;comment:打卡表单字段定义JSON(兼容)"`
	QR             string                   `json:"qr" gorm:"size:500;column:enroll_qr;comment:二维码URL"`
	ViewCnt        int                      `json:"viewCnt" gorm:"default:0;column:enroll_view_cnt;comment:浏览次数"`
	JoinCnt        int                      `json:"joinCount" gorm:"default:0;column:enroll_join_cnt;comment:打卡总次数"`
	UserCnt        int                      `json:"userCnt" gorm:"default:0;column:enroll_user_cnt;comment:参与人数"`
	UserList       string                   `json:"-" gorm:"type:text;column:enroll_user_list;comment:用户列表JSON"`
	UserListArr    []map[string]interface{} `json:"userList" gorm:"-"`
	AddTime        int64                    `json:"_createTime" gorm:"column:add_time;comment:创建时间"`
	EditTime       int64                    `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	AddIP          string                   `json:"ENROLL_ADD_IP" gorm:"size:50;column:enroll_add_ip;comment:创建IP"`
	EditIP         string                   `json:"ENROLL_EDIT_IP" gorm:"size:50;column:enroll_edit_ip;comment:修改IP"`
	CreatedAt      time.Time                `json:"-"`
	UpdatedAt      time.Time                `json:"-"`

	IsJoin         bool                     `json:"isJoin"`
	StartStr       string                   `json:"start"`
	EndStr         string                   `json:"end"`
	Content        []map[string]string      `json:"content" gorm:"-"`
	DayList        []map[string]string      `json:"dayList" gorm:"-"`
	RankList       []map[string]interface{} `json:"rankList" gorm:"-"`
	StatusDesc     string                   `json:"statusDesc" gorm:"-"`
	MyEnrollJoinID string                   `json:"myEnrollJoinId" gorm:"-"`
}

func (e Enroll) GetCreateTime() string {
	if e.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(e.AddTime).Format("2006-01-02 15:04:05")
}

type EnrollJoin struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:打卡记录ID"`
	EnrollID  string    `json:"enrollId" gorm:"size:50;column:enroll_join_enroll_id;comment:打卡项目ID"`
	UserID    string    `json:"userId" gorm:"size:200;column:enroll_join_user_id;comment:用户openid"`
	Day       string    `json:"day" gorm:"size:20;column:enroll_join_day;comment:打卡日期(YYYY-MM-DD)"`
	Forms     string    `json:"forms" gorm:"type:text;column:enroll_join_forms;comment:打卡表单数据JSON"`
	Status    int       `json:"status" gorm:"default:1;column:enroll_join_status;comment:状态:1已通过 0待审核 2未通过"`
	AddTime   int64     `json:"_createTime" gorm:"column:enroll_join_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:enroll_join_edit_time;comment:修改时间"`
	AddIP     string    `json:"ENROLL_JOIN_ADD_IP" gorm:"size:50;column:enroll_join_add_ip;comment:创建IP"`
	EditIP    string    `json:"ENROLL_JOIN_EDIT_IP" gorm:"size:50;column:enroll_join_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	EnrollTitle string   `json:"enrollTitle" gorm:"-"`
	UserName    string   `json:"userName" gorm:"-"`
	DeptName    string   `json:"deptName" gorm:"-"`
	TopDeptName string   `json:"topDeptName" gorm:"-"`
	Content     string   `json:"content" gorm:"-"`
	Images      []string `json:"images" gorm:"-"`
	Location    string   `json:"location" gorm:"-"`
}

func (ej EnrollJoin) GetCreateTime() string {
	if ej.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(ej.AddTime).Format("2006-01-02 15:04:05")
}

type EnrollUser struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:参与用户ID"`
	EnrollID       string    `json:"enrollId" gorm:"size:50;column:enroll_user_enroll_id;comment:打卡项目ID"`
	MiniOpenID     string    `json:"miniOpenId" gorm:"size:200;column:enroll_user_mini_openid;comment:用户openid"`
	Forms          string    `json:"forms" gorm:"type:text;column:enroll_user_forms;comment:报名表单数据JSON"`
	JoinCnt        int       `json:"joinCnt" gorm:"default:0;column:enroll_user_join_cnt;comment:打卡次数"`
	DayCnt         int       `json:"dayCnt" gorm:"default:0;column:enroll_user_day_cnt;comment:打卡天数"`
	LastDay        string    `json:"lastDay" gorm:"column:enroll_user_last_day;comment:最后打卡日期"`
	CheckedInToday bool      `json:"checkedInToday" gorm:"-"`
	TodayJoinCnt   int       `json:"todayJoinCnt" gorm:"-"`
	EnrollTitle    string    `json:"title" gorm:"-"`
	UserName       string    `json:"userName" gorm:"-"`
	DeptName       string    `json:"deptName" gorm:"-"`
	TopDeptName    string    `json:"topDeptName" gorm:"-"`
	DailyLimit     int       `json:"dailyLimit" gorm:"-"`
	AddTime        int64     `json:"_createTime" gorm:"column:enroll_user_add_time;comment:创建时间"`
	EditTime       int64     `json:"editTime" gorm:"column:enroll_user_edit_time;comment:修改时间"`
	AddIP          string    `json:"ENROLL_USER_ADD_IP" gorm:"size:50;column:enroll_user_add_ip;comment:创建IP"`
	EditIP         string    `json:"ENROLL_USER_EDIT_IP" gorm:"size:50;column:enroll_user_edit_ip;comment:修改IP"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}
