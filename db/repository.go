package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// Repository 存储库接口
type Repository interface {
	// 基础操作方法
	Find(id interface{}) (interface{}, error)
	FindBy(conditions map[string]interface{}) (interface{}, error)
	All() (interface{}, error)
	Create(entity interface{}) error
	Update(entity interface{}) error
	Delete(id interface{}) error

	// 构建器相关
	Query() *QueryBuilder
	WithContext(ctx context.Context) Repository
	WithTx(tx *gorm.DB) Repository
}

// BaseRepository 基础存储库实现
type BaseRepository struct {
	// 数据库连接
	db *gorm.DB
	// 上下文
	ctx context.Context
	// 模型实例
	modelType reflect.Type
	// 模型名称
	modelName string
	// 表名
	tableName string
	// 主键名称
	primaryKey string
}

// NewRepository 创建新的存储库
func NewRepository(db *gorm.DB, model interface{}) *BaseRepository {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 创建一个实例以获取表名
	modelInstance := reflect.New(modelType).Interface()
	statement := &gorm.Statement{DB: db}
	_ = statement.Parse(modelInstance)

	return &BaseRepository{
		db:         db,
		ctx:        context.Background(),
		modelType:  modelType,
		modelName:  modelType.Name(),
		tableName:  statement.Table,
		primaryKey: "id", // 默认主键为id
	}
}

// WithPrimaryKey 设置主键名称
func (r *BaseRepository) WithPrimaryKey(key string) *BaseRepository {
	r.primaryKey = key
	return r
}

// WithContext 设置上下文
func (r *BaseRepository) WithContext(ctx context.Context) Repository {
	clone := *r
	clone.ctx = ctx
	clone.db = r.db.WithContext(ctx)
	return &clone
}

// WithTx 使用事务
func (r *BaseRepository) WithTx(tx *gorm.DB) Repository {
	clone := *r
	clone.db = tx
	return &clone
}

// Query 创建查询构建器
func (r *BaseRepository) Query() *QueryBuilder {
	// 创建一个新的模型实例
	model := reflect.New(r.modelType).Interface()
	return NewQueryBuilder(r.db).Model(model).Context(r.ctx)
}

// newModel 创建一个新的模型实例
func (r *BaseRepository) newModel() interface{} {
	return reflect.New(r.modelType).Interface()
}

// newSlice 创建一个新的模型切片
func (r *BaseRepository) newSlice() any {
	sliceType := reflect.SliceOf(reflect.PtrTo(r.modelType))
	return reflect.New(sliceType).Interface()
}

// Find 根据ID查找实体
func (r *BaseRepository) Find(id interface{}) (interface{}, error) {
	model := r.newModel()
	err := r.Query().Where(fmt.Sprintf("%s = ?", r.primaryKey), id).First(model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// FindBy 根据条件查找实体
func (r *BaseRepository) FindBy(conditions map[string]interface{}) (interface{}, error) {
	model := r.newModel()
	query := r.Query()

	for field, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	err := query.First(model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// All 获取所有实体
func (r *BaseRepository) All() (interface{}, error) {
	slice := r.newSlice()
	err := r.Query().Find(slice)
	if err != nil {
		return nil, err
	}

	// 解引用切片
	return reflect.ValueOf(slice).Elem().Interface(), nil
}

// Create 创建实体
func (r *BaseRepository) Create(entity interface{}) error {
	return r.db.WithContext(r.ctx).Create(entity).Error
}

// Update 更新实体
func (r *BaseRepository) Update(entity interface{}) error {
	return r.db.WithContext(r.ctx).Save(entity).Error
}

// Delete 删除实体
func (r *BaseRepository) Delete(id interface{}) error {
	model := r.newModel()
	result := r.db.WithContext(r.ctx).Where(fmt.Sprintf("%s = ?", r.primaryKey), id).Delete(model)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("entity not found")
	}

	return nil
}

// Paginate 分页查询
func (r *BaseRepository) Paginate(page, pageSize int) (*Pagination, error) {
	slice := r.newSlice()
	return r.Query().Paginate(page, pageSize).PaginateQuery(slice)
}

// Transaction 执行事务
func (r *BaseRepository) Transaction(fn func(Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(r.WithTx(tx))
	})
}

// Count 计数
func (r *BaseRepository) Count() (int64, error) {
	var count int64
	err := r.Query().Count(&count)
	return count, err
}

// Exists 检查是否存在
func (r *BaseRepository) Exists(id interface{}) (bool, error) {
	var count int64
	err := r.Query().Where(fmt.Sprintf("%s = ?", r.primaryKey), id).Count(&count)
	return count > 0, err
}
