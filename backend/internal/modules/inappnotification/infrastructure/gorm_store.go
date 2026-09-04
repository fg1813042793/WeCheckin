package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/inappnotification/application"
	"wecheckin/backend/internal/support/notificationstyle"
	"wecheckin/backend/pkg/database"
)

type GormStore struct {
	db  *gorm.DB
	now func() time.Time
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db, now: time.Now}
}

func (store *GormStore) ResolveRecipients(ctx context.Context, rule application.RecipientRule) (application.RecipientResolution, error) {
	if store == nil || store.db == nil {
		return application.RecipientResolution{}, errors.New("in-app notification database is not initialized")
	}
	db, cancel := store.withContext(ctx)
	defer cancel()

	switch rule.Scope {
	case application.ScopeAll:
		userIDs, err := activeUserIDs(db, nil)
		return application.RecipientResolution{UserIDs: userIDs}, err
	case application.ScopeUsers:
		requested := normalizeIDs(rule.UserIDs)
		userIDs, err := activeUserIDs(db, requested)
		return application.RecipientResolution{
			UserIDs:      userIDs,
			SkippedCount: len(requested) - len(userIDs),
		}, err
	case application.ScopeDepartments:
		var departments []model.Department
		if err := db.Select("id", "dept_parent_id").Where("dept_status = ?", 1).Find(&departments).Error; err != nil {
			return application.RecipientResolution{}, err
		}
		departmentIDs := expandDepartmentIDs(rule.DepartmentIDs, departments)
		if len(departmentIDs) == 0 {
			return application.RecipientResolution{}, nil
		}
		var userIDs []uint
		err := db.Model(&model.User{}).
			Distinct("users.id").
			Joins("JOIN user_depts ON user_depts.user_dept_user_id = users.id").
			Where("users.user_status = ? AND user_depts.user_dept_dept_id IN ?", 1, departmentIDs).
			Order("users.id ASC").
			Pluck("users.id", &userIDs).Error
		return application.RecipientResolution{UserIDs: userIDs}, err
	default:
		return application.RecipientResolution{}, application.ErrInvalidScope
	}
}

func (store *GormStore) Deliver(ctx context.Context, batch application.DeliveryBatch) (application.DeliveryResult, error) {
	if store == nil || store.db == nil {
		return application.DeliveryResult{}, errors.New("in-app notification database is not initialized")
	}
	userIDs := normalizeIDs(batch.UserIDs)
	if len(userIDs) == 0 {
		return application.DeliveryResult{}, application.ErrNoRecipients
	}

	db, cancel := store.withContext(ctx)
	defer cancel()
	deliverySourceID := strings.TrimSpace(batch.DeliveryKey)
	if deliverySourceID == "" {
		deliverySourceID = batch.SourceID
	}
	deliveryKeys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		deliveryKeys = append(deliveryKeys, deliveryKey(batch.SourceType, deliverySourceID, userID))
	}
	result := application.DeliveryResult{}
	err := db.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&model.Notify{}).
			Where("notify_delivery_key IN ?", deliveryKeys).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			result = application.DeliveryResult{SentCount: int(existing), Replayed: true}
			return nil
		}

		now := store.now
		if now == nil {
			now = time.Now
		}
		addTime := now().UnixMilli()
		rows := make([]model.Notify, 0, len(userIDs))
		for index, userID := range userIDs {
			key := deliveryKeys[index]
			rows = append(rows, model.Notify{
				Title:       batch.Title,
				Content:     batch.Content,
				Type:        batch.Type,
				SourceType:  batch.SourceType,
				SourceID:    batch.SourceID,
				UserID:      strconv.FormatUint(uint64(userID), 10),
				DeliveryKey: &key,
				IsRead:      0,
				AddTime:     addTime,
			})
		}
		if err := tx.CreateInBatches(rows, 500).Error; err != nil {
			return err
		}
		result = application.DeliveryResult{SentCount: len(rows)}
		return nil
	})
	if err == nil {
		return result, nil
	}
	if !isDuplicateKeyError(err) {
		return application.DeliveryResult{}, err
	}

	var existing int64
	if countErr := db.Model(&model.Notify{}).
		Where("notify_delivery_key IN ?", deliveryKeys).
		Count(&existing).Error; countErr != nil {
		return application.DeliveryResult{}, fmt.Errorf("verify replay after duplicate delivery: %w", countErr)
	}
	if existing == 0 {
		return application.DeliveryResult{}, err
	}
	return application.DeliveryResult{SentCount: int(existing), Replayed: true}, nil
}

func (store *GormStore) List(ctx context.Context, userID string, page, pageSize int) ([]application.Notification, int64, error) {
	if store == nil || store.db == nil {
		return nil, 0, errors.New("in-app notification database is not initialized")
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	query := db.Model(&model.Notify{}).Where("notify_user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]model.Notify, 0)
	if err := query.Order("notify_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]application.Notification, 0, len(rows))
	for _, row := range rows {
		items = append(items, application.Notification{
			ID:         row.ID,
			Title:      row.Title,
			Content:    row.Content,
			Type:       row.Type,
			SourceType: row.SourceType,
			SourceID:   row.SourceID,
			IsRead:     row.IsRead,
			AddTime:    row.AddTime,
		})
	}
	return items, total, nil
}

func (store *GormStore) UnreadCount(ctx context.Context, userID string) (int64, error) {
	if store == nil || store.db == nil {
		return 0, errors.New("in-app notification database is not initialized")
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	var count int64
	err := db.Model(&model.Notify{}).
		Where("notify_user_id = ? AND notify_is_read = 0", userID).
		Count(&count).Error
	return count, err
}

func (store *GormStore) MarkRead(ctx context.Context, userID string, notificationID uint) (bool, error) {
	if store == nil || store.db == nil {
		return false, errors.New("in-app notification database is not initialized")
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	var row model.Notify
	err := db.Select("notify_id").
		Where("notify_id = ? AND notify_user_id = ?", notificationID, userID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := db.Model(&model.Notify{}).
		Where("notify_id = ? AND notify_user_id = ?", notificationID, userID).
		UpdateColumn("notify_is_read", 1).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (store *GormStore) MarkAllRead(ctx context.Context, userID string) error {
	if store == nil || store.db == nil {
		return errors.New("in-app notification database is not initialized")
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	return db.Model(&model.Notify{}).
		Where("notify_user_id = ? AND notify_is_read = 0", userID).
		UpdateColumn("notify_is_read", 1).Error
}

func (store *GormStore) RecipientOptions(ctx context.Context) (application.RecipientOptions, error) {
	if store == nil || store.db == nil {
		return application.RecipientOptions{}, errors.New("in-app notification database is not initialized")
	}
	db, cancel := store.withContext(ctx)
	defer cancel()
	var users []model.User
	if err := db.Select("id", "user_name", "user_mobile", "user_status").
		Where("user_status = ?", 1).
		Order("id ASC").
		Find(&users).Error; err != nil {
		return application.RecipientOptions{}, err
	}
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	departmentsByUser := make(map[uint][]uint, len(users))
	if len(userIDs) > 0 {
		var relations []model.UserDept
		if err := db.Select("user_dept_user_id", "user_dept_dept_id").
			Where("user_dept_user_id IN ?", userIDs).
			Order("user_dept_user_id ASC, user_dept_dept_id ASC").
			Find(&relations).Error; err != nil {
			return application.RecipientOptions{}, err
		}
		for _, relation := range relations {
			departmentsByUser[relation.UserID] = append(departmentsByUser[relation.UserID], relation.DeptID)
		}
	}
	userOptions := make([]application.RecipientUserOption, 0, len(users))
	for _, user := range users {
		userOptions = append(userOptions, application.RecipientUserOption{
			ID: user.ID, Name: user.Name, Mobile: user.Mobile, Status: user.Status,
			DeptIDs: normalizeIDs(departmentsByUser[user.ID]),
		})
	}
	var departments []model.Department
	if err := db.Select("id", "dept_name", "dept_parent_id", "dept_sort").
		Where("dept_status = ?", 1).
		Order("dept_sort ASC, id ASC").
		Find(&departments).Error; err != nil {
		return application.RecipientOptions{}, err
	}
	return application.RecipientOptions{
		Users: userOptions, Departments: buildDepartmentOptions(departments),
	}, nil
}

func (store *GormStore) NotificationStyles(ctx context.Context) (notificationstyle.Config, error) {
	if store == nil {
		return notificationstyle.Config{}, errors.New("in-app notification database is not initialized")
	}
	return notificationstyle.Load(ctx, store.db)
}

func (store *GormStore) SaveNotificationStyles(ctx context.Context, config notificationstyle.Config) (notificationstyle.Config, error) {
	if store == nil {
		return notificationstyle.Config{}, errors.New("in-app notification database is not initialized")
	}
	return notificationstyle.Save(ctx, store.db, config)
}

func activeUserIDs(db *gorm.DB, requested []uint) ([]uint, error) {
	query := db.Model(&model.User{}).Where("user_status = ?", 1)
	if requested != nil {
		if len(requested) == 0 {
			return []uint{}, nil
		}
		query = query.Where("id IN ?", requested)
	}
	var userIDs []uint
	err := query.Order("id ASC").Pluck("id", &userIDs).Error
	return userIDs, err
}

func expandDepartmentIDs(selected []uint, departments []model.Department) []uint {
	selected = normalizeIDs(selected)
	known := make(map[uint]struct{}, len(departments))
	children := make(map[uint][]uint)
	for _, department := range departments {
		if department.ID == 0 {
			continue
		}
		known[department.ID] = struct{}{}
		children[department.ParentID] = append(children[department.ParentID], department.ID)
	}
	queue := make([]uint, 0, len(selected))
	seen := make(map[uint]struct{}, len(selected))
	for _, departmentID := range selected {
		if _, exists := known[departmentID]; !exists {
			continue
		}
		seen[departmentID] = struct{}{}
		queue = append(queue, departmentID)
	}
	for index := 0; index < len(queue); index++ {
		for _, childID := range children[queue[index]] {
			if _, exists := seen[childID]; exists {
				continue
			}
			seen[childID] = struct{}{}
			queue = append(queue, childID)
		}
	}
	result := make([]uint, 0, len(seen))
	for departmentID := range seen {
		result = append(result, departmentID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func buildDepartmentOptions(departments []model.Department) []*application.RecipientDepartmentOption {
	nodes := make(map[uint]*application.RecipientDepartmentOption, len(departments))
	for _, department := range departments {
		if department.ID == 0 {
			continue
		}
		nodes[department.ID] = &application.RecipientDepartmentOption{ID: department.ID, Name: department.Name}
	}
	roots := make([]*application.RecipientDepartmentOption, 0)
	for _, department := range departments {
		node := nodes[department.ID]
		if node == nil {
			continue
		}
		parent := nodes[department.ParentID]
		if department.ParentID == 0 || parent == nil {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return roots
}

func normalizeIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func deliveryKey(sourceType, sourceID string, userID uint) string {
	payload := strings.Join([]string{sourceType, sourceID, strconv.FormatUint(uint64(userID), 10)}, "\x00")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (store *GormStore) withContext(parent context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := database.QueryContext(parent)
	return store.db.WithContext(ctx), cancel
}

var _ application.Store = (*GormStore)(nil)
