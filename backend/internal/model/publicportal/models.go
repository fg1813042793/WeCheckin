package publicportal

// Settings is a singleton (id=1). Timestamps use Unix milliseconds.
type Settings struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement:false"`
	Enabled       bool
	Title         string `gorm:"size:100"`
	Description   string `gorm:"type:text"`
	LogoURL       string `gorm:"column:logo_url;size:2048"`
	PublicBaseURL string `gorm:"column:public_base_url;size:512"`
	ShowEnded     bool
	Version       uint64
	UpdatedBy     uint
	UpdatedAt     int64 `gorm:"autoUpdateTime:false"`
}

func (Settings) TableName() string { return "public_portal_settings" }

type Item struct {
	ID          uint64 `gorm:"primaryKey"`
	Kind        string `gorm:"size:16;uniqueIndex:uk_public_portal_source,priority:1"`
	SourceID    uint   `gorm:"uniqueIndex:uk_public_portal_source,priority:2"`
	Title       string `gorm:"size:255"`
	Description string `gorm:"type:text"`
	Enabled     bool
	Sort        int
	Version     uint64
	CreatedBy   uint
	UpdatedBy   uint
	CreatedAt   int64 `gorm:"autoCreateTime:false"`
	UpdatedAt   int64 `gorm:"autoUpdateTime:false"`
}

func (Item) TableName() string { return "public_portal_items" }
