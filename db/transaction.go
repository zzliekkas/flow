// Package db 提供数据库访问和事务管理功能
package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TransactionManager 提供事务管理功能
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建新的事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

// Transaction 在事务上下文中执行函数，自动处理提交和回滚
func (tm *TransactionManager) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(fn)
}

// RunInTransaction 执行事务并返回结果，适用于需要返回值的情况
func (tm *TransactionManager) RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})
	return result, err
}

// Begin 开始一个事务
func (tm *TransactionManager) Begin(ctx context.Context) *gorm.DB {
	return tm.db.WithContext(ctx).Begin()
}

// Commit 提交事务
func (tm *TransactionManager) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback 回滚事务
func (tm *TransactionManager) Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// WithTransaction 在事务上下文中操作多个仓储
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(tx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建一个新的上下文，包含事务信息
		txCtx := NewTransactionContext(ctx, tx)
		return fn(txCtx)
	})
}

// TransactionKey 表示事务上下文的键
type TransactionKey struct{}

// NewTransactionContext 创建一个包含事务信息的上下文
func NewTransactionContext(parent context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(parent, TransactionKey{}, tx)
}

// GetTransaction 从上下文中获取事务
func GetTransaction(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(TransactionKey{}).(*gorm.DB)
	return tx, ok
}

// TxOptions 表示事务选项
type TxOptions struct {
	Isolation string
	ReadOnly  bool
}

// BeginWithOptionsResult 包含事务开始的结果
type BeginWithOptionsResult struct {
	TX  *gorm.DB
	Err error
}

// BeginWithOptions 使用选项开始一个事务，返回事务和可能的错误
func (tm *TransactionManager) BeginWithOptions(ctx context.Context, opts *TxOptions) BeginWithOptionsResult {
	if opts == nil {
		return BeginWithOptionsResult{
			TX: tm.db.WithContext(ctx).Begin(),
		}
	}

	// 将我们的选项转换为SQL隔离级别语句
	var isolationLevel string
	if opts.Isolation != "" {
		isolationLevel = "ISOLATION LEVEL " + opts.Isolation
	}

	// 开始事务并设置选项
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return BeginWithOptionsResult{TX: tx, Err: tx.Error}
	}

	// 设置隔离级别（如果指定）
	if isolationLevel != "" {
		if err := tx.Exec("SET TRANSACTION " + isolationLevel).Error; err != nil {
			tx.Rollback()
			return BeginWithOptionsResult{
				Err: fmt.Errorf("设置事务隔离级别失败: %w", err),
			}
		}
	}

	// 设置只读模式（如果指定）
	if opts.ReadOnly {
		if err := tx.Exec("SET TRANSACTION READ ONLY").Error; err != nil {
			tx.Rollback()
			return BeginWithOptionsResult{
				Err: fmt.Errorf("设置事务只读模式失败: %w", err),
			}
		}
	}

	return BeginWithOptionsResult{TX: tx}
}
