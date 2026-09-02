package workflowmodel

import "time"

const (
	DefinitionStatusDisabled  = 0
	DefinitionStatusDraft     = 1
	DefinitionStatusPublished = 2
)

type Definition struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:流程定义ID"`
	Key            string    `json:"key" gorm:"size:100;column:definition_key;uniqueIndex;comment:流程编码"`
	Name           string    `json:"name" gorm:"size:200;column:definition_name;index;comment:流程名称"`
	Description    string    `json:"description" gorm:"size:500;column:definition_description;comment:流程说明"`
	Category       string    `json:"category" gorm:"size:100;column:definition_category;index;comment:流程分类"`
	LogoURL        string    `json:"logoUrl" gorm:"size:500;column:definition_logo_url;comment:流程Logo地址"`
	Status         int       `json:"status" gorm:"default:1;column:definition_status;index;comment:状态:0停用 1草稿 2已发布"`
	CurrentVersion int       `json:"currentVersion" gorm:"default:0;column:definition_current_version;comment:当前发布版本"`
	DraftJSON      string    `json:"draftJson" gorm:"type:mediumtext;column:definition_draft_json;comment:设计器草稿JSON"`
	AddUserID      uint      `json:"addUserId" gorm:"column:definition_add_user_id;index;comment:创建人"`
	EditUserID     uint      `json:"editUserId" gorm:"column:definition_edit_user_id;comment:更新人"`
	AddTime        int64     `json:"addTime" gorm:"column:definition_add_time;comment:创建时间"`
	EditTime       int64     `json:"editTime" gorm:"column:definition_edit_time;comment:更新时间"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (Definition) TableName() string { return "workflow_definitions" }

type DefinitionVersion struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:流程版本ID"`
	DefinitionID   uint      `json:"definitionId" gorm:"column:definition_id;uniqueIndex:idx_workflow_definition_version;index;comment:流程定义ID"`
	Version        int       `json:"version" gorm:"column:definition_version;uniqueIndex:idx_workflow_definition_version;comment:版本号"`
	SourceJSON     string    `json:"sourceJson" gorm:"type:mediumtext;column:definition_source_json;comment:发布时设计器JSON"`
	BPMNXML        string    `json:"bpmnXml" gorm:"type:longtext;column:definition_bpmn_xml;comment:Flowable BPMN XML"`
	ValidationJSON string    `json:"validationJson" gorm:"type:mediumtext;column:definition_validation_json;comment:发布校验结果"`
	DeploymentID   string    `json:"deploymentId" gorm:"size:100;column:definition_deployment_id;index;comment:Flowable部署ID"`
	PublishedBy    uint      `json:"publishedBy" gorm:"column:definition_published_by;index;comment:发布人"`
	PublishedAt    int64     `json:"publishedAt" gorm:"column:definition_published_at;comment:发布时间"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (DefinitionVersion) TableName() string { return "workflow_definition_versions" }
