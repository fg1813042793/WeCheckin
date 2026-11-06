package infrastructure

import (
	"context"
	"errors"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	pm "wecheckin/backend/internal/model/publicportal"
	"wecheckin/backend/internal/modules/publicportal/application"
	"wecheckin/backend/internal/support/adminaccess"
	permissions "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

type GormStore struct {
	db   *gorm.DB
	inTx bool
}

func NewGormStore(db *gorm.DB) *GormStore { return &GormStore{db: db} }

var _ application.Store = (*GormStore)(nil)
var errDatabaseUnavailable = errors.New("public portal database is not initialized")
var errTransactionRequired = errors.New("public portal mutation requires a transaction")

func (s *GormStore) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	queryCtx, cancel := database.QueryContext(ctx)
	if s == nil || s.db == nil {
		cancel()
		return nil, func() {}, errDatabaseUnavailable
	}
	return s.db.WithContext(queryCtx), cancel, nil
}
func (s *GormStore) Transaction(ctx context.Context, fn func(application.Store) error) error {
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if fn == nil {
		return errors.New("public portal transaction callback is nil")
	}
	return storageError(db.Transaction(func(tx *gorm.DB) error { return fn(&GormStore{db: tx, inTx: true}) }))
}
func (s *GormStore) Actor(ctx context.Context, id uint) (application.Actor, error) {
	if id == 0 {
		return application.Actor{}, application.ErrForbidden
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return application.Actor{}, err
	}
	defer cancel()
	var admin model.Admin
	err = db.Select("id", "user_status", "user_role_id").Where("id = ? AND user_status = 1", id).Take(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return application.Actor{}, application.ErrForbidden
	}
	if err != nil {
		return application.Actor{}, err
	}
	roleIDs, err := permissions.ActiveRoleIDsForUserContext(db.Statement.Context, db, admin.ID, admin.RoleID)
	if err != nil {
		return application.Actor{}, err
	}
	if len(roleIDs) == 0 {
		return application.Actor{}, application.ErrForbidden
	}
	var roles []model.Role
	if err = db.Select("id", "role_name", "role_status").Where("id IN ? AND role_status = 1", roleIDs).Find(&roles).Error; err != nil {
		return application.Actor{}, err
	}
	if len(roles) == 0 {
		return application.Actor{}, application.ErrForbidden
	}
	actor := application.Actor{ID: admin.ID}
	for _, role := range roles {
		if adminaccess.IsReservedSuperAdminRole(role) {
			actor.ReservedSuperAdmin = true
		}
	}
	return actor, nil
}
func defaultSettings() pm.Settings {
	return pm.Settings{ID: 1, Title: "公开问卷与考试", Version: 1}
}
func (s *GormStore) Settings(ctx context.Context) (pm.Settings, error) {
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return pm.Settings{}, err
	}
	defer cancel()
	var item pm.Settings
	err = db.Where("id = ?", 1).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return defaultSettings(), nil
	}
	return item, storageError(err)
}
func (s *GormStore) LockSettings(ctx context.Context) (pm.Settings, error) {
	if s == nil || !s.inTx {
		return pm.Settings{}, errTransactionRequired
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return pm.Settings{}, err
	}
	defer cancel()
	seed := defaultSettings()
	// Inserts only the singleton data row. Schema is created solely by migrations.
	if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; err != nil {
		return pm.Settings{}, storageError(err)
	}
	var current pm.Settings
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", 1).Take(&current).Error
	return current, storageError(err)
}
func (s *GormStore) SaveSettings(ctx context.Context, item pm.Settings, version uint64) error {
	if s == nil || !s.inTx {
		return errTransactionRequired
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return changed(db.Model(&pm.Settings{}).Where("id = ? AND version = ?", 1, version).Select("enabled", "title", "description", "logo_url", "public_base_url", "show_ended", "version", "updated_by", "updated_at").Updates(&item))
}
func (s *GormStore) Items(ctx context.Context) ([]pm.Item, error) {
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	if s.inTx {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	items := []pm.Item{}
	err = db.Order("sort ASC").Order("id ASC").Limit(application.MaxItems + 1).Find(&items).Error
	return items, storageError(err)
}
func (s *GormStore) InsertItem(ctx context.Context, item *pm.Item) error {
	if s == nil || !s.inTx {
		return errTransactionRequired
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return changed(db.Create(item))
}
func (s *GormStore) UpdateItem(ctx context.Context, item pm.Item, version uint64) error {
	if s == nil || !s.inTx {
		return errTransactionRequired
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return changed(db.Model(&pm.Item{}).Where("id = ? AND version = ?", item.ID, version).Select("kind", "source_id", "title", "description", "enabled", "sort", "version", "updated_by", "updated_at").Updates(&item))
}
func (s *GormStore) DeleteItem(ctx context.Context, id, version uint64) error {
	if s == nil || !s.inTx {
		return errTransactionRequired
	}
	db, cancel, err := s.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return changed(db.Where("id = ? AND version = ?", id, version).Delete(&pm.Item{}))
}
func changed(result *gorm.DB) error {
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected != 1 {
		return application.ErrVersionConflict
	}
	return nil
}
func storageError(err error) error {
	var my *mysql.MySQLError
	if errors.As(err, &my) {
		switch my.Number {
		case 1146, 1054:
			return errors.Join(application.ErrSchemaMissing, err)
		case 1062:
			if strings.Contains(my.Message, "uk_public_portal_source") {
				return application.ErrDuplicateSource
			}
		}
	}
	return err
}
