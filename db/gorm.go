package db

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Model 基础模型结构
type Model struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SoftDeleteModel 带软删除的基础模型结构
type SoftDeleteModel struct {
	Model
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TimestampModel 只有时间戳的基础模型
type TimestampModel struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transaction 执行数据库事务
func Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// WithContext 在已有数据库连接上设置上下文
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx)
}

// Paginate 分页查询辅助函数
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// Paginate 执行分页查询
func Paginate(page, pageSize int, db *gorm.DB, result interface{}) (*Pagination, error) {
	var total int64
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 克隆查询以获取总数
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Count(&total).Error; err != nil {
		return nil, err
	}

	// 执行分页查询
	if err := db.Limit(pageSize).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	// 计算总页数
	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &Pagination{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPage,
	}, nil
}

// OrderBy 排序辅助函数
func OrderBy(db *gorm.DB, field string, direction string) *gorm.DB {
	if direction == "" {
		direction = "asc"
	} else if direction != "asc" && direction != "desc" {
		direction = "asc"
	}

	return db.Order(field + " " + direction)
}

// Scope 定义查询范围
type Scope func(*gorm.DB) *gorm.DB

// WithScope 应用查询范围
func WithScope(db *gorm.DB, scopes ...Scope) *gorm.DB {
	for _, scope := range scopes {
		db = scope(db)
	}
	return db
}

// Preload 预加载关联
func Preload(db *gorm.DB, relations ...string) *gorm.DB {
	for _, relation := range relations {
		db = db.Preload(relation)
	}
	return db
}

// PreloadAll 预加载所有关联
func PreloadAll(db *gorm.DB) *gorm.DB {
	return db.Preload(clause.Associations)
}

// Lock 锁定查询
type LockMode string

const (
	// LockForUpdate FOR UPDATE锁
	LockForUpdate LockMode = "FOR UPDATE"
	// LockInShareMode SHARE MODE锁
	LockInShareMode LockMode = "LOCK IN SHARE MODE"
)

// WithLock 使用锁查询
func WithLock(db *gorm.DB, mode LockMode) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: string(mode)})
}

// Cache 查询缓存接口
type Cache interface {
	// Get 获取缓存数据
	Get(ctx context.Context, key string) (interface{}, error)
	// Set 设置缓存数据
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Delete 删除缓存数据
	Delete(ctx context.Context, key string) error
	// Has 检查缓存是否存在
	Has(ctx context.Context, key string) bool
}

// WithCache 使用缓存查询
func WithCache(db *gorm.DB, cache Cache, key string, expiration time.Duration) *gorm.DB {
	// 这里只是占位符，实际缓存实现需要通过插件或中间件方式实现
	return db
}

// EnumValue 枚举类型接口
type EnumValue interface {
	String() string
	Value() int
}

// JSONField JSON字段类型
type JSONField map[string]interface{}

// GormDataType 实现 schema.GormDataType 接口
func (JSONField) GormDataType() string {
	return "json"
}

// GormDBDataType 实现数据库特定的类型定义
func (j JSONField) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "postgres":
		return "JSON"
	case "sqlite":
		return "TEXT"
	}
	return "TEXT"
}

// Scan 实现 sql.Scanner 接口
func (j *JSONField) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("无效的扫描源，需要字节切片")
	}

	if len(bytes) == 0 {
		*j = make(JSONField)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONField) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// TimeRange 时间范围结构
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// BetweenTime 时间范围查询
func BetweenTime(field string, timeRange TimeRange) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), timeRange.Start, timeRange.End)
	}
}

// IsNull 空值查询
func IsNull(field string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s IS NULL", field))
	}
}

// IsNotNull 非空值查询
func IsNotNull(field string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s IS NOT NULL", field))
	}
}

// In 包含查询
func In(field string, values ...interface{}) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s IN ?", field), values)
	}
}

// NotIn 不包含查询
func NotIn(field string, values ...interface{}) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s NOT IN ?", field), values)
	}
}

// Like 模糊查询
func Like(field string, value string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	}
}

// StartsWith 前缀查询
func StartsWith(field string, value string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s LIKE ?", field), value+"%")
	}
}

// EndsWith 后缀查询
func EndsWith(field string, value string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value)
	}
}
