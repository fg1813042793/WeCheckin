package model

import (
	"time"
	"fmt"
)

type User struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	MiniOpenID string `json:"USER_MINI_OPENID" gorm:"uniqueIndex;size:200;column:USER_MINI_OPENID"`
	Status     int    `json:"USER_STATUS" gorm:"default:1;column:USER_STATUS"`
	CheckReason string `json:"USER_CHECK_REASON" gorm:"size:500;column:USER_CHECK_REASON"`
	Name       string `json:"USER_NAME" gorm:"size:100;column:USER_NAME"`
	Mobile     string `json:"USER_MOBILE" gorm:"size:20;column:USER_MOBILE"`
	Pic        string `json:"USER_PIC" gorm:"size:500;column:USER_PIC"`
	Forms      string `json:"USER_FORMS" gorm:"type:text;column:USER_FORMS"`
	Obj        string `json:"USER_OBJ" gorm:"type:text;column:USER_OBJ"`
	LoginCnt   int    `json:"USER_LOGIN_CNT" gorm:"default:0;column:USER_LOGIN_CNT"`
	LoginTime  int64  `json:"USER_LOGIN_TIME" gorm:"column:USER_LOGIN_TIME"`
	AddTime    int64  `json:"USER_ADD_TIME" gorm:"column:USER_ADD_TIME"`
	AddIP      string `json:"USER_ADD_IP" gorm:"column:USER_ADD_IP"`
	EditTime   int64  `json:"USER_EDIT_TIME" gorm:"column:USER_EDIT_TIME"`
	EditIP     string `json:"USER_EDIT_IP" gorm:"column:USER_EDIT_IP"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (u User) GetCreateTime() string {
	if u.AddTime == 0 { return "" }
	return time.UnixMilli(u.AddTime).Format("2006-01-02 15:04:05")
}

type News struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Title    string `json:"title" gorm:"size:200;column:NEWS_TITLE"`
	Desc     string `json:"desc" gorm:"size:500;column:NEWS_DESC"`
	Status   int    `json:"status" gorm:"default:1;column:NEWS_STATUS"`
	CateID   string `json:"cateId" gorm:"size:50;column:NEWS_CATE_ID"`
	CateName string `json:"cateName" gorm:"size:50;column:NEWS_CATE_NAME"`
	Order    int    `json:"order" gorm:"default:9999;column:NEWS_ORDER"`
	Vouch    int    `json:"vouch" gorm:"default:0;column:NEWS_VOUCH"`
	Content  string `json:"content" gorm:"type:text;column:NEWS_CONTENT"`
	QR       string `json:"qr" gorm:"size:500;column:NEWS_QR"`
	ViewCnt  int    `json:"viewCnt" gorm:"default:0;column:NEWS_VIEW_CNT"`
	Pic      string `json:"img" gorm:"type:text;column:NEWS_PIC"`
	Forms    string `json:"forms" gorm:"type:text;column:NEWS_FORMS"`
	Obj      string `json:"obj" gorm:"type:text;column:NEWS_OBJ"`
	AddTime  int64  `json:"_createTime" gorm:"column:NEWS_ADD_TIME"`
	EditTime int64  `json:"editTime" gorm:"column:NEWS_EDIT_TIME"`
	AddIP    string `json:"NEWS_ADD_IP" gorm:"size:50;column:NEWS_ADD_IP"`
	EditIP   string `json:"NEWS_EDIT_IP" gorm:"size:50;column:NEWS_EDIT_IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Enroll struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Title     string `json:"title" gorm:"size:200;column:ENROLL_TITLE"`
	Desc      string `json:"desc" gorm:"-"`
	Img       string `json:"img" gorm:"-"`
	Status    int    `json:"status" gorm:"default:1;column:ENROLL_STATUS"`
	CateID    string `json:"cateId" gorm:"size:50;column:ENROLL_CATE_ID"`
	CateName  string `json:"cateName" gorm:"size:50;column:ENROLL_CATE_NAME"`
	Start     int64  `json:"timeStart" gorm:"column:ENROLL_START"`
	End       int64  `json:"timeEnd" gorm:"column:ENROLL_END"`
	DayCnt    int    `json:"dayCnt" gorm:"column:ENROLL_DAY_CNT"`
	Order     int    `json:"order" gorm:"default:9999;column:ENROLL_ORDER"`
	Vouch     int    `json:"vouch" gorm:"default:0;column:ENROLL_VOUCH"`
	Forms     string `json:"forms" gorm:"type:text;column:ENROLL_FORMS"`
	Obj       string `json:"obj" gorm:"type:text;column:ENROLL_OBJ"`
	JoinForms string `json:"joinForms" gorm:"type:text;column:ENROLL_JOIN_FORMS"`
	QR        string `json:"qr" gorm:"size:500;column:ENROLL_QR"`
	ViewCnt   int    `json:"viewCnt" gorm:"default:0;column:ENROLL_VIEW_CNT"`
	JoinCnt   int    `json:"joinCount" gorm:"default:0;column:ENROLL_JOIN_CNT"`
	UserCnt   int    `json:"userCnt" gorm:"default:0;column:ENROLL_USER_CNT"`
	UserList  string `json:"userList" gorm:"type:text;column:ENROLL_USER_LIST"`
	AddTime   int64  `json:"_createTime" gorm:"column:ENROLL_ADD_TIME"`
	EditTime  int64  `json:"editTime" gorm:"column:ENROLL_EDIT_TIME"`
	AddIP     string `json:"ENROLL_ADD_IP" gorm:"size:50;column:ENROLL_ADD_IP"`
	EditIP    string `json:"ENROLL_EDIT_IP" gorm:"size:50;column:ENROLL_EDIT_IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	IsJoin bool `json:"isJoin"`
}

type EnrollJoin struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	EnrollID string `json:"enrollId" gorm:"size:50;column:ENROLL_JOIN_ENROLL_ID"`
	UserID   string `json:"userId" gorm:"size:200;column:ENROLL_JOIN_USER_ID"`
	Day      string `json:"day" gorm:"size:20;column:ENROLL_JOIN_DAY"`
	Forms    string `json:"forms" gorm:"type:text;column:ENROLL_JOIN_FORMS"`
	Status   int    `json:"status" gorm:"default:1;column:ENROLL_JOIN_STATUS"`
	AddTime  int64  `json:"_createTime" gorm:"column:ENROLL_JOIN_ADD_TIME"`
	EditTime int64  `json:"editTime" gorm:"column:ENROLL_JOIN_EDIT_TIME"`
	AddIP    string `json:"ENROLL_JOIN_ADD_IP" gorm:"size:50;column:ENROLL_JOIN_ADD_IP"`
	EditIP   string `json:"ENROLL_JOIN_EDIT_IP" gorm:"size:50;column:ENROLL_JOIN_EDIT_IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	EnrollTitle string `json:"enrollTitle" gorm:"-"`
	Content     string `json:"content" gorm:"-"`
	Images      string `json:"images" gorm:"-"`
	Location    string `json:"location" gorm:"-"`
}

type EnrollUser struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	EnrollID   string `json:"enrollId" gorm:"size:50;column:ENROLL_USER_ENROLL_ID"`
	MiniOpenID string `json:"miniOpenId" gorm:"size:200;column:ENROLL_USER_MINI_OPENID"`
	JoinCnt    int    `json:"joinCnt" gorm:"default:0;column:ENROLL_USER_JOIN_CNT"`
	DayCnt     int    `json:"dayCnt" gorm:"default:0;column:ENROLL_USER_DAY_CNT"`
	LastDay    string `json:"lastDay" gorm:"column:ENROLL_USER_LAST_DAY"`
	AddTime    int64  `json:"_createTime" gorm:"column:ENROLL_USER_ADD_TIME"`
	EditTime   int64  `json:"editTime" gorm:"column:ENROLL_USER_EDIT_TIME"`
	AddIP      string `json:"ENROLL_USER_ADD_IP" gorm:"size:50;column:ENROLL_USER_ADD_IP"`
	EditIP     string `json:"ENROLL_USER_EDIT_IP" gorm:"size:50;column:ENROLL_USER_EDIT_IP"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type Favorite struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	UserID  string `json:"userId" gorm:"size:200;column:FAV_USER_ID"`
	Title   string `json:"title" gorm:"size:200;column:FAV_TITLE"`
	Type    string `json:"type" gorm:"size:20;column:FAV_TYPE"`
	OID     string `json:"oid" gorm:"size:50;column:FAV_OID"`
	Path    string `json:"path" gorm:"size:500;column:FAV_PATH"`
	AddTime int64  `json:"_createTime" gorm:"column:FAV_ADD_TIME"`
	EditTime int64 `json:"editTime" gorm:"column:FAV_EDIT_TIME"`
	AddIP   string `json:"FAV_ADD_IP" gorm:"size:50;column:FAV_ADD_IP"`
	EditIP  string `json:"FAV_EDIT_IP" gorm:"size:50;column:FAV_EDIT_IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Admin struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"uniqueIndex;size:100;column:ADMIN_NAME"`
	Password  string `json:"-" gorm:"size:100;column:ADMIN_PASSWORD"`
	Desc      string `json:"desc" gorm:"size:200;column:ADMIN_DESC"`
	Phone     string `json:"phone" gorm:"size:20;column:ADMIN_PHONE"`
	Status    int    `json:"status" gorm:"default:1;column:ADMIN_STATUS"`
	Type      int    `json:"type" gorm:"default:0;column:ADMIN_TYPE"`
	Token     string `json:"token" gorm:"size:100;column:ADMIN_TOKEN"`
	TokenTime int64  `json:"tokenTime" gorm:"column:ADMIN_TOKEN_TIME"`
	LoginCnt  int    `json:"loginCnt" gorm:"default:0;column:ADMIN_LOGIN_CNT"`
	LoginTime int64  `json:"loginTime" gorm:"column:ADMIN_LOGIN_TIME"`
	AddTime   int64  `json:"_createTime" gorm:"column:ADMIN_ADD_TIME"`
	EditTime  int64  `json:"editTime" gorm:"column:ADMIN_EDIT_TIME"`
	AddIP     string `json:"ADMIN_ADD_IP" gorm:"size:50;column:ADMIN_ADD_IP"`
	EditIP    string `json:"ADMIN_EDIT_IP" gorm:"size:50;column:ADMIN_EDIT_IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Log struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Type      int    `json:"type" gorm:"column:LOG_TYPE"`
	Content   string `json:"content" gorm:"type:text;column:LOG_CONTENT"`
	AdminID   string `json:"adminId" gorm:"size:50;column:LOG_ADMIN_ID"`
	AdminName string `json:"adminName" gorm:"size:100;column:LOG_ADMIN_NAME"`
	AdminDesc string `json:"adminDesc" gorm:"size:200;column:LOG_ADMIN_DESC"`
	AddTime   int64  `json:"_createTime" gorm:"column:LOG_ADD_TIME"`
	AddIP     string `json:"LOG_ADD_IP" gorm:"size:50;column:LOG_ADD_IP"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Setup struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Key      string `json:"key" gorm:"uniqueIndex;size:100;column:SETUP_KEY"`
	Value    string `json:"value" gorm:"type:text;column:SETUP_VALUE"`
	Type     string `json:"SETUP_TYPE" gorm:"size:20;column:SETUP_TYPE"`
	AddTime  int64  `json:"SETUP_ADD_TIME" gorm:"column:SETUP_ADD_TIME"`
	EditTime int64  `json:"SETUP_EDIT_TIME" gorm:"column:SETUP_EDIT_TIME"`
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
