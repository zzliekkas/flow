package db

import (
	"context"
	"database/sql"
	"errors"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// QueryBuilder 查询构建器
type QueryBuilder struct {
	// 数据库连接
	db *gorm.DB
	// 上下文
	ctx context.Context
	// 模型
	model interface{}
	// 是否使用事务
	tx bool
	// 分页参数
	page     int
	pageSize int
	// 预加载关系
	preloads []string
	// 自定义选择字段
	selects []string
	// 排序字段
	orders []string
	// 分组字段
	groups []string
	// 过滤条件
	scopes []Scope
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:       db,
		ctx:      context.Background(),
		page:     1,
		pageSize: 20,
		preloads: []string{},
		selects:  []string{},
		orders:   []string{},
		groups:   []string{},
		scopes:   []Scope{},
	}
}

// Model 设置模型
func (qb *QueryBuilder) Model(model interface{}) *QueryBuilder {
	qb.model = model
	return qb
}

// Context 设置上下文
func (qb *QueryBuilder) Context(ctx context.Context) *QueryBuilder {
	qb.ctx = ctx
	qb.db = qb.db.WithContext(ctx)
	return qb
}

// Transaction 启用事务
func (qb *QueryBuilder) Transaction() *QueryBuilder {
	qb.tx = true
	return qb
}

// Select 选择字段
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selects = append(qb.selects, fields...)
	return qb
}

// Where 添加条件
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Where(condition, args...)
	return qb
}

// Scope 应用查询范围
func (qb *QueryBuilder) Scope(scopes ...Scope) *QueryBuilder {
	qb.scopes = append(qb.scopes, scopes...)
	return qb
}

// Order 排序
func (qb *QueryBuilder) Order(order string) *QueryBuilder {
	qb.orders = append(qb.orders, order)
	return qb
}

// OrderBy 按字段排序
func (qb *QueryBuilder) OrderBy(field string, direction string) *QueryBuilder {
	if direction == "" {
		direction = "asc"
	} else if direction != "asc" && direction != "desc" {
		direction = "asc"
	}
	qb.orders = append(qb.orders, field+" "+direction)
	return qb
}

// Group 分组
func (qb *QueryBuilder) Group(group string) *QueryBuilder {
	qb.groups = append(qb.groups, group)
	return qb
}

// Preload 预加载关系
func (qb *QueryBuilder) Preload(relations ...string) *QueryBuilder {
	qb.preloads = append(qb.preloads, relations...)
	return qb
}

// PreloadAll 预加载所有关系
func (qb *QueryBuilder) PreloadAll() *QueryBuilder {
	qb.db = qb.db.Preload(clause.Associations)
	return qb
}

// Paginate 分页
func (qb *QueryBuilder) Paginate(page, pageSize int) *QueryBuilder {
	qb.page = page
	qb.pageSize = pageSize
	return qb
}

// Limit 限制数量
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.db = qb.db.Limit(limit)
	return qb
}

// Offset 偏移
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.db = qb.db.Offset(offset)
	return qb
}

// Lock 锁定查询
func (qb *QueryBuilder) Lock(mode LockMode) *QueryBuilder {
	qb.db = qb.db.Clauses(clause.Locking{Strength: string(mode)})
	return qb
}

// prepare 准备查询
func (qb *QueryBuilder) prepare() *gorm.DB {
	// 设置模型
	db := qb.db
	if qb.model != nil {
		db = db.Model(qb.model)
	}

	// 设置事务
	if qb.tx {
		tx := db.Begin()
		if tx.Error != nil {
			return db
		}
		db = tx
	}

	// 应用选择字段
	if len(qb.selects) > 0 {
		db = db.Select(qb.selects)
	}

	// 应用查询范围
	for _, scope := range qb.scopes {
		db = scope(db)
	}

	// 应用排序
	for _, order := range qb.orders {
		db = db.Order(order)
	}

	// 应用分组
	for _, group := range qb.groups {
		db = db.Group(group)
	}

	// 应用预加载
	for _, relation := range qb.preloads {
		db = db.Preload(relation)
	}

	return db
}

// Find 查询多条记录
func (qb *QueryBuilder) Find(dest interface{}) error {
	db := qb.prepare()
	return db.Find(dest).Error
}

// First 查询第一条记录
func (qb *QueryBuilder) First(dest interface{}) error {
	db := qb.prepare()
	err := db.First(dest).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// 记录不存在，返回空记录而不是错误
		return nil
	}
	return err
}

// Last 查询最后一条记录
func (qb *QueryBuilder) Last(dest interface{}) error {
	db := qb.prepare()
	err := db.Last(dest).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// Take 查询一条记录
func (qb *QueryBuilder) Take(dest interface{}) error {
	db := qb.prepare()
	err := db.Take(dest).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// FindOne 查询单条记录
func (qb *QueryBuilder) FindOne(dest interface{}) error {
	return qb.First(dest)
}

// FindByID 根据ID查询
func (qb *QueryBuilder) FindByID(id interface{}, dest interface{}) error {
	db := qb.prepare()
	err := db.First(dest, id).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// Count 计数
func (qb *QueryBuilder) Count(count *int64) error {
	db := qb.prepare()
	return db.Count(count).Error
}

// Exists 检查记录是否存在
func (qb *QueryBuilder) Exists() (bool, error) {
	var count int64
	db := qb.prepare()
	err := db.Limit(1).Count(&count).Error
	return count > 0, err
}

// Pluck 提取单个列的值
func (qb *QueryBuilder) Pluck(column string, dest interface{}) error {
	db := qb.prepare()
	return db.Pluck(column, dest).Error
}

// PaginateQuery 分页查询
func (qb *QueryBuilder) PaginateQuery(result interface{}) (*Pagination, error) {
	if qb.page < 1 {
		qb.page = 1
	}
	if qb.pageSize < 1 {
		qb.pageSize = 20
	}

	var total int64
	db := qb.prepare()

	// 计算总数
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (qb.page - 1) * qb.pageSize
	if err := db.Limit(qb.pageSize).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	// 计算总页数
	totalPage := int(total) / qb.pageSize
	if int(total)%qb.pageSize > 0 {
		totalPage++
	}

	return &Pagination{
		Page:      qb.page,
		PageSize:  qb.pageSize,
		Total:     total,
		TotalPage: totalPage,
	}, nil
}

// Create 创建记录
func (qb *QueryBuilder) Create(value interface{}) error {
	db := qb.prepare()
	return db.Create(value).Error
}

// Save 保存记录
func (qb *QueryBuilder) Save(value interface{}) error {
	db := qb.prepare()
	return db.Save(value).Error
}

// Update 更新记录
func (qb *QueryBuilder) Update(column string, value interface{}) error {
	db := qb.prepare()
	return db.Update(column, value).Error
}

// Updates 批量更新
func (qb *QueryBuilder) Updates(values interface{}) error {
	db := qb.prepare()
	return db.Updates(values).Error
}

// UpdateColumn 更新列
func (qb *QueryBuilder) UpdateColumn(column string, value interface{}) error {
	db := qb.prepare()
	return db.UpdateColumn(column, value).Error
}

// UpdateColumns 批量更新列
func (qb *QueryBuilder) UpdateColumns(values interface{}) error {
	db := qb.prepare()
	return db.UpdateColumns(values).Error
}

// Delete 删除记录
func (qb *QueryBuilder) Delete(value interface{}, conds ...interface{}) error {
	db := qb.prepare()
	return db.Delete(value, conds...).Error
}

// Exec 执行原生SQL
func (qb *QueryBuilder) Exec(sql string, values ...interface{}) error {
	db := qb.prepare()
	return db.Exec(sql, values...).Error
}

// Raw 执行原生查询
func (qb *QueryBuilder) Raw(sql string, values ...interface{}) *gorm.DB {
	db := qb.prepare()
	return db.Raw(sql, values...)
}

// Commit 提交事务
func (qb *QueryBuilder) Commit() error {
	if !qb.tx {
		return errors.New("不在事务中")
	}
	return qb.db.Commit().Error
}

// Rollback 回滚事务
func (qb *QueryBuilder) Rollback() error {
	if !qb.tx {
		return errors.New("不在事务中")
	}
	return qb.db.Rollback().Error
}

// ToSQL 获取SQL
func (qb *QueryBuilder) ToSQL() string {
	stmt := qb.prepare().Statement
	return stmt.SQL.String()
}

// Scan 扫描结果
func (qb *QueryBuilder) Scan(dest interface{}) error {
	db := qb.prepare()
	return db.Scan(dest).Error
}

// ScanRows 扫描多行
func (qb *QueryBuilder) ScanRows(rows *sql.Rows, dest interface{}) error {
	return qb.db.ScanRows(rows, dest)
}

// Row 获取单行
func (qb *QueryBuilder) Row() *sql.Row {
	db := qb.prepare()
	return db.Row()
}

// Rows 获取多行
func (qb *QueryBuilder) Rows() (*sql.Rows, error) {
	db := qb.prepare()
	return db.Rows()
}

// QueryMap 将查询结果转换为map
func (qb *QueryBuilder) QueryMap() ([]map[string]interface{}, error) {
	rows, err := qb.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// 创建结果map
		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// 处理null值
			if val == nil {
				result[col] = nil
				continue
			}

			// 尝试将字节数组转换为字符串
			if b, ok := val.([]byte); ok {
				result[col] = string(b)
			} else {
				result[col] = val
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// QueryStruct 将查询结果转换为结构体
func (qb *QueryBuilder) QueryStruct(dest interface{}) error {
	// 检查dest是否为指向切片的指针
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return errors.New("目标必须是指向切片的指针")
	}

	rows, err := qb.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	sliceValue := destValue.Elem()
	elemType := sliceValue.Type().Elem()

	for rows.Next() {
		// 创建新元素
		newElem := reflect.New(elemType).Elem()

		// 创建指向元素的指针，用于扫描
		elemToScan := newElem
		if elemType.Kind() == reflect.Ptr {
			elemToScan = reflect.New(elemType.Elem())
		}

		// 使用GORM扫描行到结构体
		if err := qb.db.ScanRows(rows, elemToScan.Addr().Interface()); err != nil {
			return err
		}

		// 将元素添加到切片
		if elemType.Kind() == reflect.Ptr {
			sliceValue.Set(reflect.Append(sliceValue, elemToScan))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, newElem))
		}
	}

	return nil
}
