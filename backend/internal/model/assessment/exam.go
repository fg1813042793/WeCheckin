package assessment

// ExamQuestion 题库题目
// 复用 formkit 题型，schema 存放完整 formkit 配置 JSON
type ExamQuestion struct {
	ID           uint   `gorm:"primaryKey;column:exam_q_id" json:"id"`
	Type         string `gorm:"column:exam_q_type;size:32" json:"type"`
	Title        string `gorm:"column:exam_q_title;size:255" json:"title"`
	Schema       string `gorm:"column:exam_q_schema;type:text" json:"schema"`
	Options      string `gorm:"column:exam_q_options;type:text" json:"options"`
	Answer       string `gorm:"column:exam_q_answer;type:text" json:"answer"`
	Score        int    `gorm:"column:exam_q_score;default:0" json:"score"`
	Category     string `gorm:"column:exam_q_category;size:64" json:"category"`
	Tags         string `gorm:"column:exam_q_tags;size:255" json:"tags"`
	Analysis     string `gorm:"column:exam_q_analysis;type:text" json:"analysis"`
	Difficulty   int    `gorm:"column:exam_q_difficulty;default:1" json:"difficulty"`
	Status       int    `gorm:"column:exam_q_status;default:1" json:"status"`
	DeptID       uint   `gorm:"column:create_dept_id;default:0" json:"deptId"`
	CreateBy     uint   `gorm:"column:create_by;default:0" json:"createBy"`
	UpdateBy     uint   `gorm:"column:update_by;default:0" json:"updateBy"`
	UpdateDeptID uint   `gorm:"column:update_dept_id;default:0" json:"updateDeptId"`
	AddTime      int64  `gorm:"column:add_time" json:"addTime"`
	EditTime     int64  `gorm:"column:edit_time" json:"editTime"`
}

func (ExamQuestion) TableName() string { return "exam_question" }

// ExamPaper 试卷：一组题 + 配置
type ExamPaper struct {
	ID           uint   `gorm:"primaryKey;column:exam_p_id" json:"id"`
	Title        string `gorm:"column:exam_p_title;size:255" json:"title"`
	Description  string `gorm:"column:exam_p_desc;type:text" json:"description"`
	QuestionIDs  string `gorm:"column:exam_p_question_ids;type:text" json:"questionIds"` // JSON 数组 [1,2,3]
	TotalScore   int    `gorm:"column:exam_p_total_score;default:0" json:"totalScore"`
	TimeLimit    int    `gorm:"column:exam_p_time_limit;default:60" json:"timeLimit"`  // 答题时长（分钟）
	PassScore    int    `gorm:"column:exam_p_pass_score;default:0" json:"passScore"`   // 及格分
	Shuffle      int    `gorm:"column:exam_p_shuffle;default:0" json:"shuffle"`        // 1=题目乱序
	ShowAnswer   int    `gorm:"column:exam_p_show_answer;default:0" json:"showAnswer"` // 1=交卷后立即显示答案
	Category     string `gorm:"column:exam_p_category;size:64" json:"category"`
	Status       int    `gorm:"column:exam_p_status;default:1" json:"status"`
	DeptID       uint   `gorm:"column:create_dept_id;default:0" json:"deptId"`
	CreateBy     uint   `gorm:"column:create_by;default:0" json:"createBy"`
	UpdateBy     uint   `gorm:"column:update_by;default:0" json:"updateBy"`
	UpdateDeptID uint   `gorm:"column:update_dept_id;default:0" json:"updateDeptId"`
	AddTime      int64  `gorm:"column:add_time" json:"addTime"`
	EditTime     int64  `gorm:"column:edit_time" json:"editTime"`
}

func (ExamPaper) TableName() string { return "exam_paper" }

// Exam 考试：一场具体考试（绑试卷 + 时间窗口 + 限员规则）
type Exam struct {
	ID           uint   `gorm:"primaryKey;column:exam_id" json:"id"`
	Title        string `gorm:"column:exam_title;size:255" json:"title"`
	Description  string `gorm:"column:exam_desc;type:text" json:"description"`
	Category     string `gorm:"column:exam_category;size:64" json:"category"`
	Tags         string `gorm:"column:exam_tags;size:255" json:"tags"`
	Visibility   int    `gorm:"column:exam_visibility;default:0" json:"visibility"`
	AllowMulti   int    `gorm:"column:exam_allow_multi;default:0" json:"allowMulti"`
	Anonymous    int    `gorm:"column:exam_anonymous;default:0" json:"anonymous"`
	ShowResult   int    `gorm:"column:exam_show_result;default:0" json:"showResult"`
	PaperID      uint   `gorm:"column:exam_paper_id" json:"paperId"`
	Schema       string `gorm:"column:exam_schema;type:text" json:"schema"`
	Settings     string `gorm:"column:exam_settings;type:text" json:"settings"`
	StartTime    int64  `gorm:"column:exam_start_time" json:"startTime"`
	EndTime      int64  `gorm:"column:exam_end_time" json:"endTime"`
	Duration     int    `gorm:"column:exam_duration;default:60" json:"duration"`
	MaxAttempts  int    `gorm:"column:exam_max_attempts;default:1" json:"maxAttempts"`
	ShowScore    int    `gorm:"column:exam_show_score;default:1" json:"showScore"`
	MaxResponse  int    `gorm:"column:exam_max_response;default:0" json:"maxResponse"`
	DeptIds      string `gorm:"column:exam_dept_ids;size:512" json:"deptIds"`
	Mode         string `gorm:"column:exam_mode;size:16;default:'exam'" json:"mode"`
	PublishDepts string `gorm:"column:exam_publish_dept_ids;size:512" json:"publishDepts"`
	QR           string `gorm:"column:exam_qr;size:512" json:"qr"`
	Status       int    `gorm:"column:exam_status;default:1" json:"status"`
	Order        int    `gorm:"column:exam_order;default:0" json:"order"`
	DeptID       uint   `gorm:"column:create_dept_id;default:0" json:"deptId"`
	CreateBy     uint   `gorm:"column:create_by;default:0" json:"createBy"`
	UpdateBy     uint   `gorm:"column:update_by;default:0" json:"updateBy"`
	UpdateDeptID uint   `gorm:"column:update_dept_id;default:0" json:"updateDeptId"`
	AddTime      int64  `gorm:"column:add_time" json:"addTime"`
	EditTime     int64  `gorm:"column:edit_time" json:"editTime"`
}

func (Exam) TableName() string { return "exam" }

// ExamRecord 考试记录：用户的一次考试
type ExamRecord struct {
	ID           uint   `gorm:"primaryKey;column:exam_r_id" json:"id"`
	ExamID       uint   `gorm:"column:exam_r_exam_id" json:"examId"`
	PaperID      uint   `gorm:"column:exam_r_paper_id" json:"paperId"`
	UserID       string `gorm:"column:exam_r_user_id;size:128" json:"userId"`
	Answers      string `gorm:"column:exam_r_answers;type:text" json:"answers"` // JSON {qid: value}
	Score        int    `gorm:"column:exam_r_score;default:0" json:"score"`
	TotalScore   int    `gorm:"column:exam_r_total_score;default:0" json:"totalScore"`
	Pass         int    `gorm:"column:exam_r_pass;default:0" json:"pass"`     // 1=通过 0=未通过
	Status       int    `gorm:"column:exam_r_status;default:0" json:"status"` // 0=进行中 1=已提交 2=已批改
	StartTime    int64  `gorm:"column:exam_r_start_time" json:"startTime"`
	SubmitTime   int64  `gorm:"column:exam_r_submit_time" json:"submitTime"`
	TimeSpent    int    `gorm:"column:exam_r_time_spent;default:0" json:"timeSpent"` // 秒
	AddIP        string `gorm:"column:exam_r_add_ip;size:64" json:"addIp"`
	IsAutoSubmit int    `gorm:"column:exam_r_auto_submit;default:0" json:"isAutoSubmit"`
	Device       string `gorm:"column:exam_r_device;size:512" json:"device"`
	DeviceID     string `gorm:"column:exam_r_device_id;size:128" json:"deviceId"`
	Session      string `gorm:"column:exam_r_session;size:128" json:"session"`
	Result       string `gorm:"column:exam_r_result;type:text" json:"result"`
}

func (ExamRecord) TableName() string { return "exam_record" }

// ExamResource 考试资源（背景图、页眉图等）
type ExamResource struct {
	ID       uint   `gorm:"primaryKey;column:exam_res_id" json:"id"`
	ExamID   uint   `gorm:"column:exam_res_exam_id;index" json:"examId"`
	Type     string `gorm:"column:exam_res_type;size:32" json:"type"` // bg / header
	URL      string `gorm:"column:exam_res_url;size:512" json:"url"`
	Filename string `gorm:"column:exam_res_filename;size:255" json:"filename"`
	Path     string `gorm:"column:exam_res_path;size:512" json:"path"`
	Domain   string `gorm:"column:exam_res_domain;size:255" json:"domain"`
	AddTime  int64  `gorm:"column:exam_res_add_time" json:"addTime"`
}

func (ExamResource) TableName() string { return "exam_resource" }
