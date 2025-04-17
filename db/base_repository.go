// Package db 提供数据库访问和仓储模式支持
package db

import (
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm"
)

// GenericRepositoryInterface 定义了泛型仓储接口
type GenericRepositoryInterface[T any] interface {
	// 基础操作方法
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, entity *T) error
	DeleteByID(ctx context.Context, id interface{}) error
	FindByID(ctx context.Context, id interface{}) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	FindByCondition(ctx context.Context, condition interface{}, args ...interface{}) ([]T, error)
	First(ctx context.Context, condition interface{}, args ...interface{}) (*T, error)
	Count(ctx context.Context, condition interface{}, args ...interface{}) (int64, error)
	Paginate(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]T, int64, error)

	// 工具方法
	WithTx(tx *gorm.DB) *GenericRepository[T]
	DB() *gorm.DB
	ModelName() string
}

// GenericRepository 提供通用的仓储实现，可用作所有特定模型仓储的基类
type GenericRepository[T any] struct {
	db *gorm.DB
}

// NewGenericRepository 创建一个新的基础仓储实例
func NewGenericRepository[T any](db *gorm.DB) *GenericRepository[T] {
	return &GenericRepository[T]{
		db: db,
	}
}

// WithTx 返回带有事务上下文的仓储实例
func (r *GenericRepository[T]) WithTx(tx *gorm.DB) *GenericRepository[T] {
	return &GenericRepository[T]{
		db: tx,
	}
}

// Create 创建一个新的实体记录
func (r *GenericRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update 更新实体记录
func (r *GenericRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete 删除实体记录
func (r *GenericRepository[T]) Delete(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Delete(entity).Error
}

// DeleteByID 根据ID删除实体记录
func (r *GenericRepository[T]) DeleteByID(ctx context.Context, id interface{}) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// FindByID 根据ID查找实体
func (r *GenericRepository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindAll 获取所有实体
func (r *GenericRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

// FindByCondition 根据条件查找实体
func (r *GenericRepository[T]) FindByCondition(ctx context.Context, condition interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where(condition, args...).Find(&entities).Error
	return entities, err
}

// First 获取满足条件的第一个实体
func (r *GenericRepository[T]) First(ctx context.Context, condition interface{}, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(condition, args...).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Count 计算满足条件的实体数量
func (r *GenericRepository[T]) Count(ctx context.Context, condition interface{}, args ...interface{}) (int64, error) {
	var count int64
	var entity T
	query := r.db.WithContext(ctx).Model(&entity)
	if condition != nil {
		query = query.Where(condition, args...)
	}
	err := query.Count(&count).Error
	return count, err
}

// Paginate 分页查询
func (r *GenericRepository[T]) Paginate(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]T, int64, error) {
	var entities []T
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx)
	if condition != nil {
		query = query.Where(condition, args...)
	}

	err := query.Model(new(T)).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []T{}, 0, nil
	}

	err = query.Offset(offset).Limit(pageSize).Find(&entities).Error
	return entities, total, err
}

// DB 返回原始的gorm.DB实例，用于高级查询
func (r *GenericRepository[T]) DB() *gorm.DB {
	return r.db
}

// ModelName 返回模型的名称
func (r *GenericRepository[T]) ModelName() string {
	var entity T
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
