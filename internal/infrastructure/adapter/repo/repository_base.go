package repo

import (
	"context"
	"fmt"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
	domainModel "github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
	"gorm.io/gorm"
)

// RepositoryBase 仓储基类
// 封装 GORM 事务 + 领域事件自动发布逻辑
type RepositoryBase struct {
	DB       *gorm.DB
	EventBus appEvent.EventBus
}

// NewRepositoryBase 创建仓储基类
func NewRepositoryBase(db *gorm.DB, eventBus appEvent.EventBus) RepositoryBase {
	return RepositoryBase{DB: db, EventBus: eventBus}
}

// SaveWithEvents 在事务中保存聚合根，成功后自动发布领域事件
// saveFn 在事务中执行持久化操作，事件发布在事务提交后执行
func (b *RepositoryBase) SaveWithEvents(ctx context.Context, root domainModel.DomainEventAware, saveFn func(tx *gorm.DB) error) error {
	err := b.DB.WithContext(ctx).Transaction(saveFn)
	if err != nil {
		return fmt.Errorf("持久化失败: %w", err)
	}

	events := root.ClearDomainEvents()
	if len(events) > 0 && b.EventBus != nil {
		return b.EventBus.Publish(ctx, events...)
	}
	return nil
}