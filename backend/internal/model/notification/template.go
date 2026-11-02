package notificationmodel

// DingTalkTemplate stores reusable manual-send configuration, never recipients.
type DingTalkTemplate struct {
	ID             uint   `gorm:"primaryKey;column:id;index:idx_dingtalk_notification_template_enabled_id,priority:2"`
	Name           string `gorm:"column:name;size:80;uniqueIndex:uk_dingtalk_notification_template_name"`
	Enabled        bool   `gorm:"column:enabled;index:idx_dingtalk_notification_template_enabled_id,priority:1"`
	Title          string `gorm:"column:title;size:255"`
	Content        string `gorm:"column:content;type:text"`
	MessageType    string `gorm:"column:message_type;size:20"`
	TargetKind     string `gorm:"column:target_kind;size:20"`
	DefinitionID   uint   `gorm:"column:definition_id"`
	DefinitionName string `gorm:"column:definition_name;size:200"`
	URL            string `gorm:"column:target_url;size:2048"`
	ButtonTitle    string `gorm:"column:button_title;size:20"`
	HeadColor      string `gorm:"column:head_color;size:8"`
	Version        int    `gorm:"column:version"`
	CreatedBy      uint   `gorm:"column:created_by"`
	UpdatedBy      uint   `gorm:"column:updated_by"`
	CreatedAt      int64  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt      int64  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (DingTalkTemplate) TableName() string { return "dingtalk_notification_templates" }
