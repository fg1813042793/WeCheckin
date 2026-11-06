package account

import "time"

const (
	PasswordChangeSourceSelf       = "self_change"
	PasswordChangeSourceAdminReset = "admin_reset"
	PasswordChangeSourceCreate     = "create"
	PasswordChangeSourceMigration  = "migration"
)

type PasswordHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey;column:id"`
	UserID       uint      `json:"userId" gorm:"index:idx_user_password_history_user_time,priority:1;column:user_id"`
	PasswordHash string    `json:"-" gorm:"size:255;column:password_hash"`
	ChangedAt    int64     `json:"changedAt" gorm:"index:idx_user_password_history_user_time,priority:2;column:changed_at"`
	ChangeSource string    `json:"changeSource" gorm:"size:32;column:change_source"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (PasswordHistory) TableName() string { return "user_password_histories" }
