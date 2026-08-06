package interaction

import "time"

type Event struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:赛事活动ID"`
	Title          string    `json:"title" gorm:"size:200;column:event_title;comment:标题"`
	Desc           string    `json:"desc" gorm:"-"`
	Img            string    `json:"img" gorm:"-"`
	Type           int       `json:"type" gorm:"default:1;column:event_type;comment:类型:1活动 2赛事"`
	Status         int       `json:"status" gorm:"default:1;column:event_status;comment:状态:1正常 0停用"`
	DeptID         uint      `json:"deptId" gorm:"default:0;column:create_dept_id;comment:创建人部门ID"`
	PublishDeptIds string    `json:"publishDeptIds" gorm:"size:500;column:event_publish_dept_ids;comment:发布部门ID列表,逗号分隔"`
	CreateBy       uint      `json:"createBy" gorm:"default:0;column:create_by;comment:创建人ID"`
	UpdateBy       uint      `json:"updateBy" gorm:"default:0;column:update_by;comment:更新人ID"`
	UpdateDeptID   uint      `json:"updateDeptId" gorm:"default:0;column:update_dept_id;comment:更新人部门ID"`
	CateID         string    `json:"cateId" gorm:"size:50;column:event_cate_id;comment:分类ID"`
	CateName       string    `json:"cateName" gorm:"size:50;column:event_cate_name;comment:分类名称"`
	RegStart       int64     `json:"regStart" gorm:"column:event_reg_start;comment:报名开始时间"`
	RegEnd         int64     `json:"regEnd" gorm:"column:event_reg_end;comment:报名结束时间"`
	EventStart     int64     `json:"eventStart" gorm:"column:event_event_start;comment:活动开始时间"`
	EventEnd       int64     `json:"eventEnd" gorm:"column:event_event_end;comment:活动结束时间"`
	Order          int       `json:"order" gorm:"default:9999;column:event_order;comment:排序值"`
	Vouch          int       `json:"vouch" gorm:"default:0;column:event_vouch;comment:推荐标记:1推荐"`
	IsTop          int       `json:"isTop" gorm:"default:0;column:event_is_top;comment:置顶标记:1置顶"`
	Forms          string    `json:"forms" gorm:"type:text;column:event_forms;comment:报名表单字段定义JSON"`
	Obj            string    `json:"obj" gorm:"type:text;column:event_obj;comment:扩展对象数据JSON(封面/描述)"`
	QR             string    `json:"qr" gorm:"size:500;column:event_qr;comment:二维码URL"`
	ViewCnt        int       `json:"viewCnt" gorm:"default:0;column:event_view_cnt;comment:浏览次数"`
	JoinCnt        int       `json:"joinCount" gorm:"default:0;column:event_join_cnt;comment:参与人次"`
	UserCnt        int       `json:"userCnt" gorm:"default:0;column:event_user_cnt;comment:参与人数"`
	AddTime        int64     `json:"_createTime" gorm:"column:add_time;comment:创建时间"`
	EditTime       int64     `json:"editTime" gorm:"column:edit_time;comment:修改时间"`
	AddIP          string    `json:"EVENT_ADD_IP" gorm:"size:50;column:event_add_ip;comment:创建IP"`
	EditIP         string    `json:"EVENT_EDIT_IP" gorm:"size:50;column:event_edit_ip;comment:修改IP"`
	ScoreFields    string    `json:"scoreFields" gorm:"type:text;column:event_score_fields;comment:评分项定义JSON"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`

	IsJoin        bool   `json:"isJoin"`
	RegStartStr   string `json:"regStartStr"`
	RegEndStr     string `json:"regEndStr"`
	EventStartStr string `json:"eventStartStr"`
	EventEndStr   string `json:"eventEndStr"`
	StatusDesc    string `json:"statusDesc" gorm:"-"`

	RoleName   string              `json:"roleName" gorm:"-"`
	Rules      string              `json:"rules" gorm:"-"`
	Organizers []map[string]string `json:"organizers" gorm:"-"`
	Assistants []map[string]string `json:"assistants" gorm:"-"`
	Referees   []map[string]string `json:"referees" gorm:"-"`
}

func (e Event) GetCreateTime() string {
	if e.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(e.AddTime).Format("2006-01-02 15:04:05")
}

type EventRole struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:角色ID"`
	EventID   uint      `json:"eventId" gorm:"index;column:event_role_event_id;comment:赛事活动ID"`
	UserID    string    `json:"userId" gorm:"size:200;index;column:event_role_user_id;comment:用户openid"`
	Role      string    `json:"role" gorm:"size:20;column:event_role_role;comment:角色:organizer/assistant/referee"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type EventParticipant struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:参与ID"`
	EventID    uint      `json:"eventId" gorm:"index;column:event_part_event_id;comment:赛事活动ID"`
	MiniOpenID string    `json:"miniOpenId" gorm:"size:200;index;column:event_part_mini_openid;comment:用户openid"`
	Forms      string    `json:"forms" gorm:"type:text;column:event_part_forms;comment:报名表单数据JSON"`
	Status     int       `json:"status" gorm:"default:1;column:event_part_status;comment:状态:1已参与"`
	AddTime    int64     `json:"_createTime" gorm:"column:event_part_add_time;comment:创建时间"`
	EditTime   int64     `json:"editTime" gorm:"column:event_part_edit_time;comment:修改时间"`
	AddIP      string    `json:"EVENT_PART_ADD_IP" gorm:"size:50;column:event_part_add_ip;comment:创建IP"`
	EditIP     string    `json:"EVENT_PART_EDIT_IP" gorm:"size:50;column:event_part_edit_ip;comment:修改IP"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`

	UserName    string `json:"userName" gorm:"-"`
	UserAvatar  string `json:"userAvatar" gorm:"-"`
	Mobile      string `json:"mobile" gorm:"-"`
	DeptName    string `json:"deptName" gorm:"-"`
	TopDeptName string `json:"topDeptName" gorm:"-"`
}

func (ep EventParticipant) GetCreateTime() string {
	if ep.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(ep.AddTime).Format("2006-01-02 15:04:05")
}

type EventDynamic struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:动态ID"`
	EventID   uint      `json:"eventId" gorm:"index;column:event_dynamic_event_id;comment:赛事活动ID"`
	UserID    string    `json:"userId" gorm:"size:200;column:event_dynamic_user_id;comment:发布者openid"`
	Title     string    `json:"title" gorm:"size:200;column:event_dynamic_title;comment:动态标题"`
	Content   string    `json:"content" gorm:"type:text;column:event_dynamic_content;comment:动态内容"`
	Images    string    `json:"images" gorm:"type:text;column:event_dynamic_images;comment:图片JSON数组"`
	Videos    string    `json:"videos" gorm:"type:text;column:event_dynamic_videos;comment:视频JSON数组"`
	AddTime   int64     `json:"_createTime" gorm:"column:event_dynamic_add_time;comment:创建时间"`
	EditTime  int64     `json:"editTime" gorm:"column:event_dynamic_edit_time;comment:修改时间"`
	AddIP     string    `json:"EVENT_DYNAMIC_ADD_IP" gorm:"size:50;column:event_dynamic_add_ip;comment:创建IP"`
	EditIP    string    `json:"EVENT_DYNAMIC_EDIT_IP" gorm:"size:50;column:event_dynamic_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	UserName   string   `json:"userName" gorm:"-"`
	UserAvatar string   `json:"userAvatar" gorm:"-"`
	ImageList  []string `json:"imageList" gorm:"-"`
	VideoList  []string `json:"videoList" gorm:"-"`
}

func (ed EventDynamic) GetCreateTime() string {
	if ed.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(ed.AddTime).Format("2006-01-02 15:04:05")
}

type EventScore struct {
	ID                 uint      `json:"id" gorm:"primaryKey;comment:成绩ID"`
	EventID            uint      `json:"eventId" gorm:"index;column:event_score_event_id;comment:赛事活动ID"`
	ParticipantID      string    `json:"participantId" gorm:"size:200;index;column:event_score_participant_id;comment:参赛者openid"`
	Score              string    `json:"score" gorm:"type:text;column:event_score_score;comment:成绩"`
	JudgeID            string    `json:"judgeId" gorm:"size:200;column:event_score_judge_id;comment:裁判openid"`
	AddTime            int64     `json:"_createTime" gorm:"column:event_score_add_time;comment:创建时间"`
	EditTime           int64     `json:"editTime" gorm:"column:event_score_edit_time;comment:修改时间"`
	AddIP              string    `json:"EVENT_SCORE_ADD_IP" gorm:"size:50;column:event_score_add_ip;comment:创建IP"`
	EditIP             string    `json:"EVENT_SCORE_EDIT_IP" gorm:"size:50;column:event_score_edit_ip;comment:修改IP"`
	CreatedAt          time.Time `json:"-"`
	UpdatedAt          time.Time `json:"-"`
	ParticipantName    string    `json:"participantName" gorm:"-"`
	ParticipantAvatar  string    `json:"participantAvatar" gorm:"-"`
	ParticipantDept    string    `json:"participantDept" gorm:"-"`
	ParticipantTopDept string    `json:"participantTopDept" gorm:"-"`
}

func (es EventScore) GetCreateTime() string {
	if es.AddTime == 0 {
		return ""
	}
	return time.UnixMilli(es.AddTime).Format("2006-01-02 15:04:05")
}

func ParseJSON(s string) []string {
	if s == "" || s == "[]" || s == "{}" {
		return nil
	}
	var result []string
	if len(s) > 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	return append(result, s)
}
