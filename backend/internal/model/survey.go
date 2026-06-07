package model

// Survey 问卷/表单定义（SurveyKing 风格的独立子系统）
// 一份 Survey = 一个 schema（题目配置）+ 投放设置 + 业务配置
type Survey struct {
	ID          uint   `gorm:"primaryKey;column:survey_id" json:"id" form:"id"`
	Title       string `gorm:"column:survey_title;size:255" json:"title" form:"title"`              // 标题
	Desc        string `gorm:"column:survey_desc;type:text" json:"description" form:"description"`  // 描述
	Schema      string `gorm:"column:survey_schema;type:text" json:"-" form:"schema"`               // formkit schema (JSON)
	Category    string `gorm:"column:survey_category;size:64" json:"category" form:"category"`      // 分类
	Tags        string `gorm:"column:survey_tags;size:255" json:"tags" form:"tags"`                 // 标签 (逗号)
	Cover       string `gorm:"column:survey_cover;size:512" json:"cover" form:"cover"`              // 封面
	Visibility  int    `gorm:"column:survey_visibility;default:0" json:"visibility" form:"visibility"`
	AllowMulti  int    `gorm:"column:survey_allow_multi;default:0" json:"allowMulti" form:"allowMulti"`
	StartTime   int64  `gorm:"column:survey_start_time" json:"startTime" form:"startTime"`
	EndTime     int64  `gorm:"column:survey_end_time" json:"endTime" form:"endTime"`
	MaxResponse int    `gorm:"column:survey_max_response;default:0" json:"maxResponse" form:"maxResponse"`
	ShowResult  int    `gorm:"column:survey_show_result;default:0" json:"showResult" form:"showResult"`
	Anonymous   int    `gorm:"column:survey_anonymous;default:0" json:"anonymous" form:"anonymous"`
	DeptIDs     string `gorm:"column:survey_dept_ids;size:512" json:"deptIds" form:"deptIds"`
	QR          string `gorm:"column:survey_qr;size:512" json:"qr" form:"qr"`
	Status      int    `gorm:"column:survey_status;default:1" json:"status" form:"status"`
	Mode        string `gorm:"column:survey_mode;size:16;default:'survey'" json:"mode" form:"mode"`
	Order       int    `gorm:"column:survey_order;default:0" json:"order" form:"order"`
	DeptID      uint   `gorm:"column:survey_dept_id;default:0" json:"deptId" form:"deptId"`
	CreateBy    uint   `gorm:"column:survey_create_by;default:0" json:"createBy" form:"createBy"`
	AddTime     int64  `gorm:"column:survey_add_time" json:"addTime" form:"addTime"`
	EditTime    int64  `gorm:"column:survey_edit_time" json:"editTime" form:"editTime"`
	Settings string `gorm:"column:survey_settings;type:text" json:"settings" form:"settings"`
}

func (Survey) TableName() string { return "survey" }

// SurveyResponse 答卷（一用户一次提交）
type SurveyResponse struct {
	ID         uint   `gorm:"primaryKey;column:survey_resp_id" json:"id"`
	SurveyID   uint   `gorm:"column:survey_resp_survey_id" json:"surveyId"`
	UserID     string `gorm:"column:survey_resp_user_id;size:128" json:"userId"`        // 匿名时为空
	Nickname   string `gorm:"column:survey_resp_nickname;size:128" json:"nickname"`     // 冗余（防用户改名）
	Answers    string `gorm:"column:survey_resp_answers;type:text" json:"-"`           // {qid: value} JSON
	Duration   int    `gorm:"column:survey_resp_duration;default:0" json:"duration"`   // 答题时长(秒)
	Status     int    `gorm:"column:survey_resp_status;default:1" json:"status"`       // 1=已完成 0=草稿
	IP         string `gorm:"column:survey_resp_ip;size:64" json:"ip"`
	Device     string `gorm:"column:survey_resp_device;size:255" json:"device"`
	StartTime  int64  `gorm:"column:survey_resp_start_time" json:"startTime"`
	SubmitTime int64  `gorm:"column:survey_resp_submit_time" json:"submitTime"`
	AddTime    int64  `gorm:"column:survey_resp_add_time" json:"addTime"`
}

func (SurveyResponse) TableName() string { return "survey_response" }

// SurveyChannel 投放渠道（链接/二维码/嵌入）
type SurveyChannel struct {
	ID        uint   `gorm:"primaryKey;column:survey_ch_id" json:"id"`
	SurveyID  uint   `gorm:"column:survey_ch_survey_id" json:"surveyId"`
	Name      string `gorm:"column:survey_ch_name;size:128" json:"name"`
	Type      string `gorm:"column:survey_ch_type;size:32" json:"type"`  // link/qr/embed/shorturl
	URL       string `gorm:"column:survey_ch_url;size:512" json:"url"`
	Extra     string `gorm:"column:survey_ch_extra;type:text" json:"extra"` // 渠道额外参数
	VisitCnt  int    `gorm:"column:survey_ch_visit_cnt;default:0" json:"visitCnt"`
	SubmitCnt int    `gorm:"column:survey_ch_submit_cnt;default:0" json:"submitCnt"`
	Status    int    `gorm:"column:survey_ch_status;default:1" json:"status"`
	AddTime   int64  `gorm:"column:survey_ch_add_time" json:"addTime"`
}

func (SurveyChannel) TableName() string { return "survey_channel" }

// SurveyAILog AI 生成记录（审计 + 成本统计）
type SurveyAILog struct {
	ID         uint   `gorm:"primaryKey;column:survey_ai_id" json:"id"`
	Prompt     string `gorm:"column:survey_ai_prompt;type:text" json:"prompt"`
	SchemaOut  string `gorm:"column:survey_ai_schema_out;type:text" json:"schemaOut"`
	SurveyID   uint   `gorm:"column:survey_ai_survey_id" json:"surveyId"`
	TokensUsed int    `gorm:"column:survey_ai_tokens_used;default:0" json:"tokensUsed"`
	Model      string `gorm:"column:survey_ai_model;size:64" json:"model"`
	CostUSD    int    `gorm:"column:survey_ai_cost_usd_micro;default:0" json:"costUsdMicro"` // 微美元
	Status     int    `gorm:"column:survey_ai_status;default:1" json:"status"`               // 1=成功 0=失败
	ErrorMsg   string `gorm:"column:survey_ai_error;size:512" json:"errorMsg"`
	AdminID    uint   `gorm:"column:survey_ai_admin_id" json:"adminId"`
	AddTime    int64  `gorm:"column:survey_ai_add_time" json:"addTime"`
}

func (SurveyAILog) TableName() string { return "survey_ai_log" }

// SurveyResource 问卷资源（背景图、页眉图等）
type SurveyResource struct {
	ID       uint   `gorm:"primaryKey;column:survey_res_id" json:"id"`
	SurveyID uint   `gorm:"column:survey_res_survey_id;index" json:"surveyId"`
	Type     string `gorm:"column:survey_res_type;size:32" json:"type"` // bg / header
	URL      string `gorm:"column:survey_res_url;size:512" json:"url"`
	Filename string `gorm:"column:survey_res_filename;size:255" json:"filename"`
	Path     string `gorm:"column:survey_res_path;size:512" json:"path"`
	Domain   string `gorm:"column:survey_res_domain;size:255" json:"domain"`
	AddTime  int64  `gorm:"column:survey_res_add_time" json:"addTime"`
}

func (SurveyResource) TableName() string { return "survey_resource" }

// SurveyQuestion 问卷题库题目
type SurveyQuestion struct {
	ID       uint   `gorm:"primaryKey;column:survey_q_id" json:"id"`
	Type     string `gorm:"column:survey_q_type;size:32" json:"type"`
	Title    string `gorm:"column:survey_q_title;size:255" json:"title"`
	Schema   string `gorm:"column:survey_q_schema;type:text" json:"schema"`       // 完整 formkit schema JSON
	Category string `gorm:"column:survey_q_category;size:64" json:"category"`
	Tags     string `gorm:"column:survey_q_tags;size:255" json:"tags"`
	Status   int    `gorm:"column:survey_q_status;default:1" json:"status"`
	DeptID   uint   `gorm:"column:survey_q_dept_id;default:0" json:"deptId"`
	CreateBy uint   `gorm:"column:survey_q_create_by;default:0" json:"createBy"`
	AddTime  int64  `gorm:"column:survey_q_add_time" json:"addTime"`
}

func (SurveyQuestion) TableName() string { return "survey_question" }
