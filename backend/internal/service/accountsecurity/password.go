package accountsecurity

import (
	"context"
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/passwordsession"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

var (
	ErrCurrentPasswordInvalid = errors.New("当前密码不正确")
	ErrPasswordReused         = errors.New("新密码不能与当前或最近使用的密码相同")
)

const (
	PasswordChangeSourceSelf       = "self_change"
	PasswordChangeSourceAdminReset = "admin_reset"
	PasswordChangeSourceCreate     = "create"
	passwordHistoryRetention       = 20
)

type Credential struct {
	UserID             uint
	PasswordHash       string
	PasswordChangedAt  int64
	PasswordMustChange bool
}

type PasswordUpdate struct {
	UserID       uint
	NewHash      string
	PreviousHash string
	ChangedAt    int64
	MustChange   bool
	ChangeSource string
}

type PasswordStore interface {
	WithinTransaction(ctx context.Context, fn func(PasswordStore) error) error
	CredentialForUpdate(ctx context.Context, userID uint) (Credential, error)
	RecentPasswordHashes(ctx context.Context, userID uint, limit int) ([]string, error)
	ApplyPasswordUpdate(ctx context.Context, update PasswordUpdate) error
}

type PolicyLoader func(context.Context) (Policy, error)

type PasswordService struct {
	store      PasswordStore
	loadPolicy PolicyLoader
}

func NewPasswordService(store PasswordStore, loadPolicy PolicyLoader) *PasswordService {
	return &PasswordService{store: store, loadPolicy: loadPolicy}
}

func (s *PasswordService) ChangeOwnPasswordContext(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	return s.change(ctx, userID, currentPassword, newPassword, PasswordChangeSourceSelf, false, true)
}

func (s *PasswordService) SetInitialPasswordContext(ctx context.Context, userID uint, password string) error {
	return s.change(ctx, userID, "", password, PasswordChangeSourceCreate, true, false)
}

func (s *PasswordService) ResetTemporaryPasswordContext(ctx context.Context, userID uint, password string) error {
	return s.change(ctx, userID, "", password, PasswordChangeSourceAdminReset, true, false)
}

func (s *PasswordService) change(ctx context.Context, userID uint, currentPassword, newPassword, source string, temporary, verifyCurrent bool) error {
	policy, err := s.loadPolicy(ctx)
	if err != nil {
		return err
	}
	return s.store.WithinTransaction(ctx, func(store PasswordStore) error {
		credential, err := store.CredentialForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		if verifyCurrent && !passwordutil.Verify(credential.PasswordHash, currentPassword) {
			return ErrCurrentPasswordInvalid
		}
		if !temporary || !policy.Enabled || !policy.AllowWeakTemporaryPassword {
			if err := ValidatePermanentPassword(policy, newPassword); err != nil {
				return err
			}
		}
		if passwordutil.Verify(credential.PasswordHash, newPassword) {
			return ErrPasswordReused
		}
		if policy.Enabled && policy.HistoryCount > 0 {
			hashes, err := store.RecentPasswordHashes(ctx, userID, policy.HistoryCount)
			if err != nil {
				return err
			}
			for _, hash := range hashes {
				if passwordutil.Verify(hash, newPassword) {
					return ErrPasswordReused
				}
			}
		}
		newHash, err := passwordutil.Hash(newPassword)
		if err != nil {
			return err
		}
		mustChange := false
		if policy.Enabled && temporary {
			mustChange = (source == PasswordChangeSourceCreate && policy.ForceChangeOnCreate) ||
				(source == PasswordChangeSourceAdminReset && policy.ForceChangeOnReset)
		}
		return store.ApplyPasswordUpdate(ctx, PasswordUpdate{
			UserID: userID, NewHash: newHash, PreviousHash: credential.PasswordHash,
			ChangedAt: time.Now().UnixMilli(), MustChange: mustChange, ChangeSource: source,
		})
	})
}

type GormPasswordStore struct{ db *gorm.DB }

func NewGormPasswordStore(db *gorm.DB) *GormPasswordStore { return &GormPasswordStore{db: db} }

func (s *GormPasswordStore) WithinTransaction(ctx context.Context, fn func(PasswordStore) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GormPasswordStore{db: tx})
	})
}

func (s *GormPasswordStore) CredentialForUpdate(ctx context.Context, userID uint) (Credential, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Select("id", "user_password", "user_password_changed_at", "user_password_must_change").
		Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
		return Credential{}, err
	}
	return Credential{UserID: user.ID, PasswordHash: user.Password, PasswordChangedAt: user.PasswordChangedAt, PasswordMustChange: user.PasswordMustChange == 1}, nil
}

func (s *GormPasswordStore) RecentPasswordHashes(ctx context.Context, userID uint, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}
	var hashes []string
	err := s.db.WithContext(ctx).Model(&model.PasswordHistory{}).Where("user_id = ?", userID).
		Order("changed_at DESC, id DESC").Limit(limit).Pluck("password_hash", &hashes).Error
	return hashes, err
}

func (s *GormPasswordStore) ApplyPasswordUpdate(ctx context.Context, update PasswordUpdate) error {
	now := update.ChangedAt
	if update.PreviousHash != "" {
		history := model.PasswordHistory{UserID: update.UserID, PasswordHash: update.PreviousHash, ChangedAt: now, ChangeSource: update.ChangeSource}
		if err := s.db.WithContext(ctx).Create(&history).Error; err != nil {
			return err
		}
	}
	result := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", update.UserID).Updates(map[string]interface{}{
		"user_password": update.NewHash, "user_password_changed_at": now,
		"user_password_must_change": boolInt(update.MustChange), "user_edit_time": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	var staleIDs []uint
	// MySQL requires LIMIT with OFFSET, even when selecting every remaining row.
	if err := s.db.WithContext(ctx).Model(&model.PasswordHistory{}).Where("user_id = ?", update.UserID).
		Order("changed_at DESC, id DESC").Limit(math.MaxInt).Offset(passwordHistoryRetention).Pluck("id", &staleIDs).Error; err != nil {
		return err
	}
	if len(staleIDs) > 0 {
		return s.db.WithContext(ctx).Where("id IN ?", staleIDs).Delete(&model.PasswordHistory{}).Error
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func defaultPasswordService() *PasswordService {
	return NewPasswordService(NewGormPasswordStore(database.GetDB()), LoadPolicyContext)
}

func ChangeOwnPasswordContext(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	if err := defaultPasswordService().ChangeOwnPasswordContext(ctx, userID, currentPassword, newPassword); err != nil {
		return err
	}
	return passwordsession.InvalidateAllContext(ctx, userID)
}

func SetInitialPasswordContext(ctx context.Context, userID uint, password string) error {
	return defaultPasswordService().SetInitialPasswordContext(ctx, userID, password)
}

func ResetTemporaryPasswordContext(ctx context.Context, userID uint, password string) error {
	if err := defaultPasswordService().ResetTemporaryPasswordContext(ctx, userID, password); err != nil {
		return err
	}
	return passwordsession.InvalidateAllContext(ctx, userID)
}

func SetInitialPasswordWithDBContext(ctx context.Context, db *gorm.DB, userID uint, password string) error {
	return NewPasswordService(NewGormPasswordStore(db), LoadPolicyContext).SetInitialPasswordContext(ctx, userID, password)
}

func ResetTemporaryPasswordWithDBContext(ctx context.Context, db *gorm.DB, userID uint, password string) error {
	return NewPasswordService(NewGormPasswordStore(db), LoadPolicyContext).ResetTemporaryPasswordContext(ctx, userID, password)
}
