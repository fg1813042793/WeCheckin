package infrastructure

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/pkg/database"
)

var (
	ErrDefinitionNotPublished = errors.New("流程定义尚未发布")
	ErrInstanceNotFound       = errors.New("流程实例不存在")
	ErrTaskNotFound           = errors.New("流程任务不存在")
)

type GormStore struct {
	db    *gorm.DB
	txCtx context.Context
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (store *GormStore) InTransaction(ctx context.Context, fn func(application.TransactionStore) error) error {
	if store == nil || store.db == nil {
		return errors.New("工作流数据库未初始化")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	return store.db.WithContext(queryCtx).Transaction(func(tx *gorm.DB) error {
		return fn(&GormStore{db: tx, txCtx: queryCtx})
	})
}

func (store *GormStore) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if store == nil || store.db == nil {
		return nil, func() {}, errors.New("工作流数据库未初始化")
	}
	if store.txCtx != nil {
		return store.db.WithContext(store.txCtx), func() {}, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return store.db.WithContext(queryCtx), cancel, nil
}
