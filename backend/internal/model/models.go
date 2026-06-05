package model

import (
	"time"
	"fmt"
)

type User struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:用户ID"`
	MiniOpenID string    `json:"miniOpenID" gorm:"uniqueIndex;size:200;column:user_mini_openid;comment:微信小程序openid"`
	Status     int       `json:"status" gorm:"default:1;column:user_status;comment:状态:1正常 0禁用"`
	CheckReason string   `json:"checkReason" gorm:"size:500;column:user_check_reason;comment:审核原因"`
	Name       string    `json:"name" gorm:"size:100;column:user_name;comment:用户昵称"`
	Mobile     string    `json:"mobile" gorm:"size:20;column:user_mobile;comment:手机号"`
	Pic        string    `json:"avatar" gorm:"size:500;column:user_pic;comment:头像URL"`
	Forms      string    `json:"forms" gorm:"type:text;column:user_forms;comment:扩展表单数据JSON"`
	Obj        string    `json:"obj" gorm:"type:text;column:user_obj;comment:扩展对象数据JSON"`
	Password   string    `json:"-" gorm:"size:100;column:user_password;comment:密码(md5 hex)"`
	LoginCnt   int       `json:"loginCnt" gorm:"default:0;column:user_login_cnt;comment:登录次数"`
	LoginTime  int64     `json:"loginTime" gorm:"column:user_login_time;comment:最后登录时间"`
	AddTime    int64     `json:"addTime" gorm:"column:user_add_time;comment:创建时间"`
	AddIP      string    `json:"addIP" gorm:"column:user_add_ip;comment:创建IP"`
	EditTime   int64     `json:"editTime" gorm:"column:user_edit_time;comment:修改时间"`
	EditIP     string    `json:"editIP" gorm:"column:user_edit_ip;comment:修改IP"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`

	Role        string `json:"role" gorm:"-"`
	DeptName    string `json:"deptName" gorm:"-"`
	TopDeptName string `json:"topDeptName" gorm:"-"`
}

func (u *User) GetRole() string {
	if u.Status == 9 || u.Status == 1 {
		return "admin"
	}
	return "user"
}

func (u User) GetCreateTime() string {
	if u.AddTime == 0 { return "" }
	return time.UnixMilli(u.AddTime).Format("2006-01-02 15:04:05")
}

type News struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:新闻ID"`
	Title    string `json:"title" gorm:"size:200;column:news_title;comment:新闻标题"`
	Desc     string `json:"desc" gorm:"size:500;column:news_desc;comment:新闻简介"`
	Status   int    `json:"status" gorm:"default:1;column:news_status;comment:状态:1正常 0禁用"`
	DeptID          uint   `json:"deptId" gorm:"default:0;column:news_dept_id;comment:所属部门ID"`
	PublishDeptIds  string `json:"publishDeptIds" gorm:"size:500;column:news_publish_dept_ids;comment:发布部门ID列表,逗号分隔"`
	CreateBy        uint   `json:"createBy" gorm:"default:0;column:news_create_by;comment:创建管理员ID"`
	CateID   string `json:"cateId" gorm:"size:50;column:news_cate_id;comment:分类ID"`
	CateName string `json:"cateName" gorm:"size:50;column:news_cate_name;comment:分类名称"`
	Order    int    `json:"order" gorm:"default:9999;column:news_order;comment:排序值"`
	Vouch    int    `json:"vouch" gorm:"default:0;column:news_vouch;comment:推荐标记:1推荐"`
	Content  string `json:"content" gorm:"type:text;column:news_content;comment:新闻内容"`
	QR       string `json:"qr" gorm:"size:500;column:news_qr;comment:二维码URL"`
	ViewCnt  int    `json:"viewCnt" gorm:"default:0;column:news_view_cnt;comment:浏览次数"`
	Pic      string `json:"-" gorm:"type:text;column:news_pic;comment:图片列表JSON"`
	Img      string `json:"img" gorm:"-"`
	Forms    string `json:"forms" gorm:"type:text;column:news_forms;comment:扩展表单数据JSON"`
	Obj      string `json:"obj" gorm:"type:text;column:news_obj;comment:扩展对象数据JSON"`
	AddTime  int64  `json:"_createTime" gorm:"column:news_add_time;comment:创建时间"`
	EditTime int64  `json:"editTime" gorm:"column:news_edit_time;comment:修改时间"`
	AddIP    string `json:"NEWS_ADD_IP" gorm:"size:50;column:news_add_ip;comment:创建IP"`
	EditIP   string `json:"NEWS_EDIT_IP" gorm:"size:50;column:news_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Enroll struct {
	ID        uint   `json:"id" gorm:"primaryKey;comment:打卡项目ID"`
	Title     string `json:"title" gorm:"size:200;column:enroll_title;comment:打卡标题"`
	Desc      string `json:"desc" gorm:"-"`
	Img       string `json:"img" gorm:"-"`
	Status    int    `json:"status" gorm:"default:1;column:enroll_status;comment:状态:1正常 0停用"`
	DeptID          uint   `json:"deptId" gorm:"default:0;column:enroll_dept_id;comment:所属部门ID"`
	PublishDeptIds  string `json:"publishDeptIds" gorm:"size:500;column:enroll_publish_dept_ids;comment:发布部门ID列表,逗号分隔"`
	CreateBy        uint   `json:"createBy" gorm:"default:0;column:enroll_create_by;comment:创建管理员ID"`
	CateID    string `json:"cateId" gorm:"size:50;column:enroll_cate_id;comment:分类ID"`
	CateName  string `json:"cateName" gorm:"size:50;column:enroll_cate_name;comment:分类名称"`
	Start     int64  `json:"timeStart" gorm:"column:enroll_start;comment:开始时间"`
	End       int64  `json:"timeEnd" gorm:"column:enroll_end;comment:结束时间"`
	DayCnt    int    `json:"dayCnt" gorm:"column:enroll_day_cnt;comment:打卡天数"`
	Order     int    `json:"order" gorm:"default:9999;column:enroll_order;comment:排序值"`
	Vouch       int  `json:"vouch" gorm:"default:0;column:enroll_vouch;comment:推荐标记:1推荐"`
	AllowRepeat bool `json:"allowRepeat" gorm:"default:0;column:enroll_repeat;comment:允许重复打卡"`
	DailyLimit  int  `json:"dailyLimit" gorm:"default:1;column:enroll_limit;comment:每日打卡次数限制"`
	Forms       string `json:"forms" gorm:"type:text;column:enroll_forms;comment:打卡表单字段定义JSON"`
	Obj         string `json:"obj" gorm:"type:text;column:enroll_obj;comment:扩展对象数据JSON(封面/描述)"`
	JoinForms   string `json:"joinForms" gorm:"type:text;column:enroll_join_forms;comment:打卡表单字段定义JSON(兼容)"`
	QR        string `json:"qr" gorm:"size:500;column:enroll_qr;comment:二维码URL"`
	ViewCnt   int    `json:"viewCnt" gorm:"default:0;column:enroll_view_cnt;comment:浏览次数"`
	JoinCnt   int    `json:"joinCount" gorm:"default:0;column:enroll_join_cnt;comment:打卡总次数"`
	UserCnt      int                      `json:"userCnt" gorm:"default:0;column:enroll_user_cnt;comment:参与人数"`
	UserList     string                   `json:"-" gorm:"type:text;column:enroll_user_list;comment:用户列表JSON"`
	UserListArr  []map[string]interface{} `json:"userList" gorm:"-"`
	AddTime      int64                    `json:"_createTime" gorm:"column:enroll_add_time;comment:创建时间"`
	EditTime  int64  `json:"editTime" gorm:"column:enroll_edit_time;comment:修改时间"`
	AddIP     string `json:"ENROLL_ADD_IP" gorm:"size:50;column:enroll_add_ip;comment:创建IP"`
	EditIP    string `json:"ENROLL_EDIT_IP" gorm:"size:50;column:enroll_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	IsJoin        bool                     `json:"isJoin"`
	StartStr      string                   `json:"start"`
	EndStr        string                   `json:"end"`
	Content       []map[string]string      `json:"content" gorm:"-"`
	DayList       []map[string]string      `json:"dayList" gorm:"-"`
	RankList      []map[string]interface{} `json:"rankList" gorm:"-"`
	StatusDesc    string                   `json:"statusDesc" gorm:"-"`
	MyEnrollJoinID string                  `json:"myEnrollJoinId" gorm:"-"`
}

type EnrollJoin struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:打卡记录ID"`
	EnrollID string `json:"enrollId" gorm:"size:50;column:enroll_join_enroll_id;comment:打卡项目ID"`
	UserID   string `json:"userId" gorm:"size:200;column:enroll_join_user_id;comment:用户openid"`
	Day      string `json:"day" gorm:"size:20;column:enroll_join_day;comment:打卡日期(YYYY-MM-DD)"`
	Forms    string `json:"forms" gorm:"type:text;column:enroll_join_forms;comment:打卡表单数据JSON"`
	Status   int    `json:"status" gorm:"default:1;column:enroll_join_status;comment:状态:1已通过 0待审核 2未通过"`
	AddTime  int64  `json:"_createTime" gorm:"column:enroll_join_add_time;comment:创建时间"`
	EditTime int64  `json:"editTime" gorm:"column:enroll_join_edit_time;comment:修改时间"`
	AddIP    string `json:"ENROLL_JOIN_ADD_IP" gorm:"size:50;column:enroll_join_add_ip;comment:创建IP"`
	EditIP   string `json:"ENROLL_JOIN_EDIT_IP" gorm:"size:50;column:enroll_join_edit_ip;comment:修改IP"`
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

type EnrollUser struct {
	ID              uint   `json:"id" gorm:"primaryKey;comment:参与用户ID"`
	EnrollID        string `json:"enrollId" gorm:"size:50;column:enroll_user_enroll_id;comment:打卡项目ID"`
	MiniOpenID      string `json:"miniOpenId" gorm:"size:200;column:enroll_user_mini_openid;comment:用户openid"`
	Forms           string `json:"forms" gorm:"type:text;column:enroll_user_forms;comment:报名表单数据JSON"`
	JoinCnt         int    `json:"joinCnt" gorm:"default:0;column:enroll_user_join_cnt;comment:打卡次数"`
	DayCnt          int    `json:"dayCnt" gorm:"default:0;column:enroll_user_day_cnt;comment:打卡天数"`
	LastDay         string `json:"lastDay" gorm:"column:enroll_user_last_day;comment:最后打卡日期"`
	CheckedInToday  bool   `json:"checkedInToday" gorm:"-"`
	TodayJoinCnt    int    `json:"todayJoinCnt" gorm:"-"`
	EnrollTitle     string `json:"title" gorm:"-"`
	UserName        string `json:"userName" gorm:"-"`
	DeptName        string `json:"deptName" gorm:"-"`
	TopDeptName     string `json:"topDeptName" gorm:"-"`
	DailyLimit      int    `json:"dailyLimit" gorm:"-"`
	AddTime         int64  `json:"_createTime" gorm:"column:enroll_user_add_time;comment:创建时间"`
	EditTime        int64  `json:"editTime" gorm:"column:enroll_user_edit_time;comment:修改时间"`
	AddIP           string `json:"ENROLL_USER_ADD_IP" gorm:"size:50;column:enroll_user_add_ip;comment:创建IP"`
	EditIP          string `json:"ENROLL_USER_EDIT_IP" gorm:"size:50;column:enroll_user_edit_ip;comment:修改IP"`
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

type Favorite struct {
	ID      uint   `json:"id" gorm:"primaryKey;comment:收藏ID"`
	UserID  string `json:"userId" gorm:"size:200;column:fav_user_id;comment:用户openid"`
	Title   string `json:"title" gorm:"size:200;column:fav_title;comment:收藏标题"`
	Type    string `json:"type" gorm:"size:20;column:fav_type;comment:收藏类型"`
	OID     string `json:"oid" gorm:"size:50;column:fav_oid;comment:关联对象ID"`
	Path    string `json:"path" gorm:"size:500;column:fav_path;comment:收藏路径"`
	AddTime int64  `json:"_createTime" gorm:"column:fav_add_time;comment:创建时间"`
	EditTime int64 `json:"editTime" gorm:"column:fav_edit_time;comment:修改时间"`
	AddIP   string `json:"FAV_ADD_IP" gorm:"size:50;column:fav_add_ip;comment:创建IP"`
	EditIP  string `json:"FAV_EDIT_IP" gorm:"size:50;column:fav_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Admin struct {
	ID        uint   `json:"id" gorm:"primaryKey;comment:管理员ID"`
	Name      string `json:"name" gorm:"uniqueIndex;size:100;column:admin_name;comment:管理员用户名"`
	Password  string `json:"-" gorm:"size:100;column:admin_password;comment:密码(sha256 hex)"`
	Desc      string `json:"desc" gorm:"size:200;column:admin_desc;comment:管理员描述"`
	Pic       string `json:"pic" gorm:"size:500;column:admin_pic;comment:头像URL"`
	Phone     string `json:"phone" gorm:"size:20;column:admin_phone;comment:手机号"`
	Status    int    `json:"status" gorm:"default:1;column:admin_status;comment:状态:1正常 0禁用"`
	Type      int    `json:"type" gorm:"default:0;column:admin_type;comment:类型:1超级管理员"`
	RoleID    uint   `json:"roleId" gorm:"default:0;column:admin_role_id;comment:角色ID"`
	Token     string `json:"token" gorm:"size:100;column:admin_token;comment:登录token"`
	TokenTime int64  `json:"tokenTime" gorm:"column:admin_token_time;comment:token生成时间"`
	LoginCnt  int    `json:"loginCnt" gorm:"default:0;column:admin_login_cnt;comment:登录次数"`
	LoginTime int64  `json:"loginTime" gorm:"column:admin_login_time;comment:最后登录时间"`
	AddTime   int64  `json:"_createTime" gorm:"column:admin_add_time;comment:创建时间"`
	EditTime  int64  `json:"editTime" gorm:"column:admin_edit_time;comment:修改时间"`
	AddIP     string `json:"ADMIN_ADD_IP" gorm:"size:50;column:admin_add_ip;comment:创建IP"`
	EditIP    string `json:"ADMIN_EDIT_IP" gorm:"size:50;column:admin_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Log struct {
	ID        uint   `json:"id" gorm:"primaryKey;comment:日志ID"`
	Type      int    `json:"type" gorm:"column:log_type;comment:日志类型"`
	Content   string `json:"content" gorm:"type:text;column:log_content;comment:日志内容"`
	AdminID   string `json:"adminId" gorm:"size:50;column:log_admin_id;comment:管理员ID"`
	AdminName string `json:"adminName" gorm:"size:100;column:log_admin_name;comment:管理员用户名"`
	AdminDesc string `json:"adminDesc" gorm:"size:200;column:log_admin_desc;comment:管理员描述"`
	AddTime   int64  `json:"_createTime" gorm:"column:log_add_time;comment:创建时间"`
	AddIP     string `json:"LOG_ADD_IP" gorm:"size:50;column:log_add_ip;comment:创建IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Setup struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:配置ID"`
	Key      string `json:"key" gorm:"uniqueIndex;size:100;column:setup_key;comment:配置键名"`
	Value    string `json:"value" gorm:"type:text;column:setup_value;comment:配置值"`
	Type     string `json:"setup_type" gorm:"size:20;column:setup_type;comment:配置类型"`
	AddTime  int64  `json:"setup_add_time" gorm:"column:setup_add_time;comment:创建时间"`
	EditTime int64  `json:"setup_edit_time" gorm:"column:setup_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type UserDept struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	UserID    uint      `json:"userId" gorm:"index;column:user_dept_user_id;comment:用户ID"`
	DeptID    uint      `json:"deptId" gorm:"index;column:user_dept_dept_id;comment:部门ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Department struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:部门ID"`
	Name     string `json:"name" gorm:"size:100;column:dept_name;comment:部门名称"`
	ParentID uint   `json:"parentId" gorm:"default:0;column:dept_parent_id;comment:上级部门ID"`
	Sort     int    `json:"sort" gorm:"default:0;column:dept_sort;comment:排序"`
	Status   int    `json:"status" gorm:"default:1;column:dept_status;comment:状态:1正常 0禁用"`
	AddTime  int64  `json:"addTime" gorm:"column:dept_add_time;comment:创建时间"`
	EditTime int64  `json:"editTime" gorm:"column:dept_edit_time;comment:修改时间"`
	AddIP    string `json:"deptAddIp" gorm:"size:50;column:dept_add_ip;comment:创建IP"`
	EditIP   string `json:"deptEditIp" gorm:"size:50;column:dept_edit_ip;comment:修改IP"`
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
	ID        uint       `json:"id" gorm:"primaryKey;comment:菜单ID"`
	Name      string     `json:"name" gorm:"size:100;column:menu_name;comment:菜单名称"`
	ParentID  uint       `json:"parentId" gorm:"default:0;column:menu_parent_id;comment:上级菜单ID"`
	Path      string     `json:"path" gorm:"size:200;column:menu_path;comment:路由路径"`
	Perms     string     `json:"perms" gorm:"size:200;column:menu_perms;comment:权限标识(多个逗号分隔)"`
	Icon      string     `json:"icon" gorm:"size:100;column:menu_icon;comment:图标"`
	Sort      int        `json:"sort" gorm:"default:0;column:menu_sort;comment:排序"`
	Status    int        `json:"status" gorm:"default:1;column:menu_status;comment:状态:1正常 0禁用"`
	Type      int        `json:"type" gorm:"default:1;column:menu_type;comment:类型:0目录 1菜单 2按钮"`
	AddTime   int64      `json:"addTime" gorm:"column:menu_add_time;comment:创建时间"`
	EditTime  int64      `json:"editTime" gorm:"column:menu_edit_time;comment:修改时间"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Children  []*Menu    `json:"children" gorm:"-"`
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

type AdminDept struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	AdminID   uint      `json:"adminId" gorm:"index;column:admin_dept_admin_id;comment:管理员ID"`
	DeptID    uint      `json:"deptId" gorm:"index;column:admin_dept_dept_id;comment:部门ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type SysDict struct {
	ID        uint   `json:"id" gorm:"primaryKey;comment:字典ID"`
	TypeCode  string `json:"typeCode" gorm:"size:50;column:dict_type_code;index;comment:字典类型编码"`
	TypeName  string `json:"typeName" gorm:"size:100;column:dict_type_name;comment:字典类型名称"`
	Label     string `json:"label" gorm:"size:100;column:dict_label;comment:字典标签"`
	Value     string `json:"value" gorm:"size:200;column:dict_value;comment:字典值"`
	Sort      int    `json:"sort" gorm:"default:0;column:dict_sort;comment:排序"`
	Status    int    `json:"status" gorm:"default:1;column:dict_status;comment:状态(1正常 0停用)"`
	Remark    string `json:"remark" gorm:"size:500;column:dict_remark;comment:备注"`
	AddTime   int64  `json:"addTime" gorm:"column:dict_add_time;comment:创建时间"`
	EditTime  int64  `json:"editTime" gorm:"column:dict_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type UserFormField struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:字段ID"`
	Label    string `json:"label" gorm:"size:100;column:field_label;comment:字段名称"`
	Type     string `json:"type" gorm:"size:20;column:field_type;comment:字段类型(文本/数字/多行文本/选择/图片/定位)"`
	Required int    `json:"required" gorm:"default:0;column:field_required;comment:是否必填"`
	Options  string `json:"options" gorm:"size:500;column:field_options;comment:选项(逗号分隔)"`
	Sort     int    `json:"sort" gorm:"default:0;column:field_sort;comment:排序"`
	AddTime  int64  `json:"addTime" gorm:"column:field_add_time;comment:创建时间"`
	EditTime int64  `json:"editTime" gorm:"column:field_edit_time;comment:修改时间"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (n News) GetCreateTime() string {
	if n.AddTime == 0 { return "" }
	return time.UnixMilli(n.AddTime).Format("2006-01-02 15:04:05")
}

func (e Enroll) GetCreateTime() string {
	if e.AddTime == 0 { return "" }
	return time.UnixMilli(e.AddTime).Format("2006-01-02 15:04:05")
}

func (ej EnrollJoin) GetCreateTime() string {
	if ej.AddTime == 0 { return "" }
	return time.UnixMilli(ej.AddTime).Format("2006-01-02 15:04:05")
}

// Event models for 赛事活动
type Event struct {
	ID        uint   `json:"id" gorm:"primaryKey;comment:赛事活动ID"`
	Title     string `json:"title" gorm:"size:200;column:event_title;comment:标题"`
	Desc      string `json:"desc" gorm:"-"`
	Img       string `json:"img" gorm:"-"`
	Type      int    `json:"type" gorm:"default:1;column:event_type;comment:类型:1活动 2赛事"`
	Status    int    `json:"status" gorm:"default:1;column:event_status;comment:状态:1正常 0停用"`
	DeptID          uint   `json:"deptId" gorm:"default:0;column:event_dept_id;comment:所属部门ID"`
	PublishDeptIds  string `json:"publishDeptIds" gorm:"size:500;column:event_publish_dept_ids;comment:发布部门ID列表,逗号分隔"`
	CreateBy        uint   `json:"createBy" gorm:"default:0;column:event_create_by;comment:创建管理员ID"`
	CateID    string `json:"cateId" gorm:"size:50;column:event_cate_id;comment:分类ID"`
	CateName  string `json:"cateName" gorm:"size:50;column:event_cate_name;comment:分类名称"`
	RegStart  int64  `json:"regStart" gorm:"column:event_reg_start;comment:报名开始时间"`
	RegEnd    int64  `json:"regEnd" gorm:"column:event_reg_end;comment:报名结束时间"`
	EventStart int64 `json:"eventStart" gorm:"column:event_event_start;comment:活动开始时间"`
	EventEnd  int64  `json:"eventEnd" gorm:"column:event_event_end;comment:活动结束时间"`
	Order     int    `json:"order" gorm:"default:9999;column:event_order;comment:排序值"`
	Vouch     int    `json:"vouch" gorm:"default:0;column:event_vouch;comment:推荐标记:1推荐"`
	IsTop     int    `json:"isTop" gorm:"default:0;column:event_is_top;comment:置顶标记:1置顶"`
	Forms     string `json:"forms" gorm:"type:text;column:event_forms;comment:报名表单字段定义JSON"`
	Obj       string `json:"obj" gorm:"type:text;column:event_obj;comment:扩展对象数据JSON(封面/描述)"`
	QR        string `json:"qr" gorm:"size:500;column:event_qr;comment:二维码URL"`
	ViewCnt   int    `json:"viewCnt" gorm:"default:0;column:event_view_cnt;comment:浏览次数"`
	JoinCnt   int    `json:"joinCount" gorm:"default:0;column:event_join_cnt;comment:参与人次"`
	UserCnt   int    `json:"userCnt" gorm:"default:0;column:event_user_cnt;comment:参与人数"`
	AddTime     int64  `json:"_createTime" gorm:"column:event_add_time;comment:创建时间"`
	EditTime    int64  `json:"editTime" gorm:"column:event_edit_time;comment:修改时间"`
	AddIP       string `json:"EVENT_ADD_IP" gorm:"size:50;column:event_add_ip;comment:创建IP"`
	EditIP      string `json:"EVENT_EDIT_IP" gorm:"size:50;column:event_edit_ip;comment:修改IP"`
	ScoreFields string `json:"scoreFields" gorm:"type:text;column:event_score_fields;comment:评分项定义JSON"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`

	IsJoin       bool   `json:"isJoin"`
	RegStartStr  string `json:"regStartStr"`
	RegEndStr    string `json:"regEndStr"`
	EventStartStr string `json:"eventStartStr"`
	EventEndStr   string `json:"eventEndStr"`
	StatusDesc   string `json:"statusDesc" gorm:"-"`

	RoleName   string              `json:"roleName" gorm:"-"`
	Rules      string              `json:"rules" gorm:"-"`
	Organizers []map[string]string `json:"organizers" gorm:"-"`
	Assistants []map[string]string `json:"assistants" gorm:"-"`
	Referees   []map[string]string `json:"referees" gorm:"-"`
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

	UserName  string `json:"userName" gorm:"-"`
	UserAvatar string `json:"userAvatar" gorm:"-"`
	DeptName   string `json:"deptName" gorm:"-"`
	TopDeptName string `json:"topDeptName" gorm:"-"`
}

type EventDynamic struct {
	ID      uint   `json:"id" gorm:"primaryKey;comment:动态ID"`
	EventID uint   `json:"eventId" gorm:"index;column:event_dynamic_event_id;comment:赛事活动ID"`
	UserID  string `json:"userId" gorm:"size:200;column:event_dynamic_user_id;comment:发布者openid"`
	Title   string `json:"title" gorm:"size:200;column:event_dynamic_title;comment:动态标题"`
	Content string `json:"content" gorm:"type:text;column:event_dynamic_content;comment:动态内容"`
	Images  string `json:"images" gorm:"type:text;column:event_dynamic_images;comment:图片JSON数组"`
	Videos  string `json:"videos" gorm:"type:text;column:event_dynamic_videos;comment:视频JSON数组"`
	AddTime int64  `json:"_createTime" gorm:"column:event_dynamic_add_time;comment:创建时间"`
	EditTime int64 `json:"editTime" gorm:"column:event_dynamic_edit_time;comment:修改时间"`
	AddIP   string `json:"EVENT_DYNAMIC_ADD_IP" gorm:"size:50;column:event_dynamic_add_ip;comment:创建IP"`
	EditIP  string `json:"EVENT_DYNAMIC_EDIT_IP" gorm:"size:50;column:event_dynamic_edit_ip;comment:修改IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	UserName   string   `json:"userName" gorm:"-"`
	UserAvatar string   `json:"userAvatar" gorm:"-"`
	ImageList  []string `json:"imageList" gorm:"-"`
	VideoList  []string `json:"videoList" gorm:"-"`
}

type EventScore struct {
	ID            uint      `json:"id" gorm:"primaryKey;comment:成绩ID"`
	EventID       uint      `json:"eventId" gorm:"index;column:event_score_event_id;comment:赛事活动ID"`
	ParticipantID string    `json:"participantId" gorm:"size:200;index;column:event_score_participant_id;comment:参赛者openid"`
	Score         string    `json:"score" gorm:"type:text;column:event_score_score;comment:成绩"`
	JudgeID       string    `json:"judgeId" gorm:"size:200;column:event_score_judge_id;comment:裁判openid"`
	AddTime       int64     `json:"_createTime" gorm:"column:event_score_add_time;comment:创建时间"`
	EditTime      int64     `json:"editTime" gorm:"column:event_score_edit_time;comment:修改时间"`
	AddIP         string    `json:"EVENT_SCORE_ADD_IP" gorm:"size:50;column:event_score_add_ip;comment:创建IP"`
	EditIP        string    `json:"EVENT_SCORE_EDIT_IP" gorm:"size:50;column:event_score_edit_ip;comment:修改IP"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`

	ParticipantName   string `json:"participantName" gorm:"-"`
	ParticipantAvatar string `json:"participantAvatar" gorm:"-"`
	ParticipantDept   string `json:"participantDept" gorm:"-"`
	ParticipantTopDept string `json:"participantTopDept" gorm:"-"`
}

func (e Event) GetCreateTime() string {
	if e.AddTime == 0 { return "" }
	return time.UnixMilli(e.AddTime).Format("2006-01-02 15:04:05")
}

func (ep EventParticipant) GetCreateTime() string {
	if ep.AddTime == 0 { return "" }
	return time.UnixMilli(ep.AddTime).Format("2006-01-02 15:04:05")
}

func (ed EventDynamic) GetCreateTime() string {
	if ed.AddTime == 0 { return "" }
	return time.UnixMilli(ed.AddTime).Format("2006-01-02 15:04:05")
}

func (es EventScore) GetCreateTime() string {
	if es.AddTime == 0 { return "" }
	return time.UnixMilli(es.AddTime).Format("2006-01-02 15:04:05")
}

// ParseJSON parses a JSON string into a slice of strings
func ParseJSON(s string) []string {
	if s == "" || s == "[]" || s == "{}" {
		return nil
	}
	var result []string
	// Simple bracket removal
	if len(s) > 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	return append(result, s)
}

func init() {
	fmt.Println("models initialized")
}
