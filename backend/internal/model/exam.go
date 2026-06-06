package model

// ExamQuestion 题库题目
// 复用 formkit 题型，但额外带：正确答案 (Answer)、分值 (Score)、分类 (Category)、解析 (Analysis)
type ExamQuestion struct {
	ID         uint   `gorm:"primaryKey;column:exam_q_id" json:"id"`
	Type       string `gorm:"column:exam_q_type;size:32" json:"type"`            // input/select/radio/checkbox/textarea
	Title      string `gorm:"column:exam_q_title;size:255" json:"title"`          // 题干
	Options    string `gorm:"column:exam_q_options;type:text" json:"options"`     // JSON 数组 [{label,value}]
	Answer     string `gorm:"column:exam_q_answer;type:text" json:"answer"`       // 正确答案（JSON：string/[]string）
	Score      int    `gorm:"column:exam_q_score;default:0" json:"score"`         // 分值
	Category   string `gorm:"column:exam_q_category;size:64" json:"category"`     // 分类
	Tags       string `gorm:"column:exam_q_tags;size:255" json:"tags"`            // 逗号分隔标签
	Analysis   string `gorm:"column:exam_q_analysis;type:text" json:"analysis"`   // 解析
	Difficulty int    `gorm:"column:exam_q_difficulty;default:1" json:"difficulty"` // 1=易 2=中 3=难
	Status     int    `gorm:"column:exam_q_status;default:1" json:"status"`      // 1=启用 0=停用
	DeptID     uint   `gorm:"column:exam_q_dept_id;default:0" json:"deptId"`      // 归属部门
	CreateBy   uint   `gorm:"column:exam_q_create_by;default:0" json:"createBy"`
	AddTime    int64  `gorm:"column:exam_q_add_time" json:"addTime"`
}

func (ExamQuestion) TableName() string { return "exam_question" }

// ExamPaper 试卷：一组题 + 配置
type ExamPaper struct {
	ID          uint   `gorm:"primaryKey;column:exam_p_id" json:"id"`
	Title       string `gorm:"column:exam_p_title;size:255" json:"title"`
	Description string `gorm:"column:exam_p_desc;type:text" json:"description"`
	QuestionIDs string `gorm:"column:exam_p_question_ids;type:text" json:"questionIds"` // JSON 数组 [1,2,3]
	TotalScore  int    `gorm:"column:exam_p_total_score;default:0" json:"totalScore"`
	TimeLimit   int    `gorm:"column:exam_p_time_limit;default:60" json:"timeLimit"`        // 答题时长（分钟）
	PassScore   int    `gorm:"column:exam_p_pass_score;default:0" json:"passScore"`         // 及格分
	Shuffle     int    `gorm:"column:exam_p_shuffle;default:0" json:"shuffle"`            // 1=题目乱序
	ShowAnswer  int    `gorm:"column:exam_p_show_answer;default:0" json:"showAnswer"`      // 1=交卷后立即显示答案
	Category    string `gorm:"column:exam_p_category;size:64" json:"category"`
	Status      int    `gorm:"column:exam_p_status;default:1" json:"status"`
	DeptID      uint   `gorm:"column:exam_p_dept_id;default:0" json:"deptId"`
	CreateBy    uint   `gorm:"column:exam_p_create_by;default:0" json:"createBy"`
	AddTime     int64  `gorm:"column:exam_p_add_time" json:"addTime"`
}

func (ExamPaper) TableName() string { return "exam_paper" }

// Exam 考试：一场具体考试（绑试卷 + 时间窗口 + 限员规则）
type Exam struct {
	ID           uint   `gorm:"primaryKey;column:exam_id" json:"id"`
	Title        string `gorm:"column:exam_title;size:255" json:"title"`
	PaperID      uint   `gorm:"column:exam_paper_id" json:"paperId"`
	StartTime    int64  `gorm:"column:exam_start_time" json:"startTime"`    // 开始时间 (ms)
	EndTime      int64  `gorm:"column:exam_end_time" json:"endTime"`        // 结束时间 (ms)
	Duration     int    `gorm:"column:exam_duration;default:60" json:"duration"` // 单人答题时长（分钟），0=不限
	MaxAttempts  int    `gorm:"column:exam_max_attempts;default:1" json:"maxAttempts"`
	ShowScore    int    `gorm:"column:exam_show_score;default:1" json:"showScore"` // 1=交卷后立即显示分数
	PublishDepts string `gorm:"column:exam_publish_dept_ids;size:512" json:"publishDepts"`
	QR           string `gorm:"column:exam_qr;size:512" json:"qr"`
	Status       int    `gorm:"column:exam_status;default:1" json:"status"`
	Order        int    `gorm:"column:exam_order;default:0" json:"order"`
	DeptID       uint   `gorm:"column:exam_dept_id;default:0" json:"deptId"`
	CreateBy     uint   `gorm:"column:exam_create_by;default:0" json:"createBy"`
	AddTime      int64  `gorm:"column:exam_add_time" json:"addTime"`
}

func (Exam) TableName() string { return "exam" }

// ExamRecord 考试记录：用户的一次考试
type ExamRecord struct {
	ID         uint   `gorm:"primaryKey;column:exam_r_id" json:"id"`
	ExamID     uint   `gorm:"column:exam_r_exam_id" json:"examId"`
	PaperID    uint   `gorm:"column:exam_r_paper_id" json:"paperId"`
	UserID     string `gorm:"column:exam_r_user_id;size:128" json:"userId"`
	Answers    string `gorm:"column:exam_r_answers;type:text" json:"answers"`     // JSON {qid: value}
	Score      int    `gorm:"column:exam_r_score;default:0" json:"score"`
	TotalScore int    `gorm:"column:exam_r_total_score;default:0" json:"totalScore"`
	Pass       int    `gorm:"column:exam_r_pass;default:0" json:"pass"`               // 1=通过 0=未通过
	Status     int    `gorm:"column:exam_r_status;default:0" json:"status"`           // 0=进行中 1=已提交 2=已批改
	StartTime  int64  `gorm:"column:exam_r_start_time" json:"startTime"`
	SubmitTime int64  `gorm:"column:exam_r_submit_time" json:"submitTime"`
	TimeSpent  int    `gorm:"column:exam_r_time_spent;default:0" json:"timeSpent"`     // 秒
	AddIP      string `gorm:"column:exam_r_add_ip;size:64" json:"addIp"`
	Result     string `gorm:"column:exam_r_result;type:text" json:"result"`
}

func (ExamRecord) TableName() string { return "exam_record" }
