package db

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 迁移错误定义
var (
	ErrMigrationFailed    = errors.New("迁移执行失败")
	ErrMigrationNotFound  = errors.New("找不到指定的迁移")
	ErrMigrationExists    = errors.New("迁移已存在")
	ErrInvalidMigration   = errors.New("无效的迁移")
	ErrInvalidMigrationID = errors.New("无效的迁移ID")
)

// 定义迁移相关常量
const (
	// 迁移表名
	MigrationTable = "migrations"
	// 迁移文件模板
	MigrationTemplate = `package migrations

import (
	"github.com/zzliekkas/flow/db"
	"gorm.io/gorm"
)

func init() {
	db.RegisterMigration("%s", "%s", up%s, down%s)
}

// up%s 执行迁移
func up%s(db *gorm.DB) error {
	// TODO: 实现迁移逻辑
	return nil
}

// down%s 回滚迁移
func down%s(db *gorm.DB) error {
	// TODO: 实现回滚逻辑
	return nil
}
`
)

// Migration 迁移接口
type Migration interface {
	// ID 获取迁移ID
	ID() string
	// Name 获取迁移名称
	Name() string
	// Up 执行迁移
	Up(db *gorm.DB) error
	// Down 回滚迁移
	Down(db *gorm.DB) error
}

// BaseMigration 基础迁移实现
type BaseMigration struct {
	id   string
	name string
	up   func(db *gorm.DB) error
	down func(db *gorm.DB) error
}

// ID 获取迁移ID
func (m *BaseMigration) ID() string {
	return m.id
}

// Name 获取迁移名称
func (m *BaseMigration) Name() string {
	return m.name
}

// Up 执行迁移
func (m *BaseMigration) Up(db *gorm.DB) error {
	if m.up == nil {
		return nil
	}
	return m.up(db)
}

// Down 回滚迁移
func (m *BaseMigration) Down(db *gorm.DB) error {
	if m.down == nil {
		return nil
	}
	return m.down(db)
}

// NewMigration 创建新的迁移
func NewMigration(id, name string, up, down func(db *gorm.DB) error) Migration {
	return &BaseMigration{
		id:   id,
		name: name,
		up:   up,
		down: down,
	}
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID        string    `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Batch     int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

// TableName 表名
func (MigrationRecord) TableName() string {
	return "migrations"
}

// Migrator 迁移管理器
type Migrator struct {
	db         *gorm.DB
	migrations map[string]Migration
	directory  string
}

// NewMigrator 创建迁移管理器
func NewMigrator(db *gorm.DB, directory string) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make(map[string]Migration),
		directory:  directory,
	}
}

// Register 注册迁移
func (m *Migrator) Register(migration Migration) error {
	id := migration.ID()
	if id == "" {
		return ErrInvalidMigrationID
	}

	if _, exists := m.migrations[id]; exists {
		return ErrMigrationExists
	}

	m.migrations[id] = migration
	return nil
}

// ensureTable 确保迁移表存在
func (m *Migrator) ensureTable() error {
	return m.db.AutoMigrate(&MigrationRecord{})
}

// GetRan 获取已执行的迁移
func (m *Migrator) GetRan() ([]string, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	var records []MigrationRecord
	if err := m.db.Order("id").Find(&records).Error; err != nil {
		return nil, err
	}

	ran := make([]string, len(records))
	for i, record := range records {
		ran[i] = record.ID
	}

	return ran, nil
}

// GetMigrations 获取所有迁移
func (m *Migrator) GetMigrations() []Migration {
	migrations := make([]Migration, 0, len(m.migrations))
	for _, migration := range m.migrations {
		migrations = append(migrations, migration)
	}

	// 按ID排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID() < migrations[j].ID()
	})

	return migrations
}

// GetPending 获取待执行的迁移
func (m *Migrator) GetPending() ([]Migration, error) {
	ran, err := m.GetRan()
	if err != nil {
		return nil, err
	}

	// 创建已执行迁移集合，用于快速查找
	ranMap := make(map[string]bool)
	for _, id := range ran {
		ranMap[id] = true
	}

	// 获取所有未执行的迁移
	migrations := m.GetMigrations()
	pending := make([]Migration, 0)

	for _, migration := range migrations {
		if !ranMap[migration.ID()] {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// Migrate 执行所有待执行的迁移
func (m *Migrator) Migrate() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	pending, err := m.GetPending()
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		return nil // 没有待执行的迁移
	}

	// 获取当前批次号
	var batch int
	var record MigrationRecord
	err = m.db.Order("batch desc").First(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	batch = record.Batch + 1

	// 执行迁移
	for _, migration := range pending {
		// 使用事务执行迁移
		err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行迁移
			if err := migration.Up(tx); err != nil {
				return err
			}

			// 记录迁移
			return tx.Create(&MigrationRecord{
				ID:        migration.ID(),
				Name:      migration.Name(),
				Batch:     batch,
				CreatedAt: time.Now(),
			}).Error
		})

		if err != nil {
			return fmt.Errorf("迁移 %s 失败: %w", migration.ID(), err)
		}
	}

	return nil
}

// Rollback 回滚最后一批次的迁移
func (m *Migrator) Rollback() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	// 查找最后一批次的迁移
	var records []MigrationRecord
	var lastBatch int

	err := m.db.Order("batch desc").Limit(1).Pluck("batch", &lastBatch).Error
	if err != nil {
		return err
	}

	err = m.db.Where("batch = ?", lastBatch).Order("id desc").Find(&records).Error
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // 没有可回滚的迁移
	}

	// 执行回滚
	for _, record := range records {
		migration, exists := m.migrations[record.ID]
		if !exists {
			return fmt.Errorf("找不到迁移 %s: %w", record.ID, ErrMigrationNotFound)
		}

		// 使用事务执行回滚
		err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行回滚
			if err := migration.Down(tx); err != nil {
				return err
			}

			// 删除迁移记录
			return tx.Delete(&MigrationRecord{}, "id = ?", record.ID).Error
		})

		if err != nil {
			return fmt.Errorf("回滚 %s 失败: %w", record.ID, err)
		}
	}

	return nil
}

// Reset 重置所有迁移
func (m *Migrator) Reset() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	// 查找所有已执行的迁移
	var records []MigrationRecord
	err := m.db.Order("id desc").Find(&records).Error
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // 没有可重置的迁移
	}

	// 执行重置
	for _, record := range records {
		migration, exists := m.migrations[record.ID]
		if !exists {
			return fmt.Errorf("找不到迁移 %s: %w", record.ID, ErrMigrationNotFound)
		}

		// 使用事务执行重置
		err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行回滚
			if err := migration.Down(tx); err != nil {
				return err
			}

			// 删除迁移记录
			return tx.Delete(&MigrationRecord{}, "id = ?", record.ID).Error
		})

		if err != nil {
			return fmt.Errorf("重置 %s 失败: %w", record.ID, err)
		}
	}

	return nil
}

// RollbackTo 回滚到指定的迁移
func (m *Migrator) RollbackTo(id string) error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	// 查找比目标迁移更新的所有迁移
	var records []MigrationRecord
	err := m.db.Where("id > ?", id).Order("id desc").Find(&records).Error
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // 没有可回滚的迁移
	}

	// 执行回滚
	for _, record := range records {
		migration, exists := m.migrations[record.ID]
		if !exists {
			return fmt.Errorf("找不到迁移 %s: %w", record.ID, ErrMigrationNotFound)
		}

		// 使用事务执行回滚
		err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行回滚
			if err := migration.Down(tx); err != nil {
				return err
			}

			// 删除迁移记录
			return tx.Delete(&MigrationRecord{}, "id = ?", record.ID).Error
		})

		if err != nil {
			return fmt.Errorf("回滚 %s 失败: %w", record.ID, err)
		}
	}

	return nil
}

// Status 获取迁移状态
func (m *Migrator) Status() ([]map[string]interface{}, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	// 获取所有迁移
	migrations := m.GetMigrations()

	// 获取已执行的迁移
	ran, err := m.GetRan()
	if err != nil {
		return nil, err
	}

	// 创建已执行迁移集合，用于快速查找
	ranMap := make(map[string]bool)
	for _, id := range ran {
		ranMap[id] = true
	}

	// 构建状态信息
	status := make([]map[string]interface{}, len(migrations))
	for i, migration := range migrations {
		var statusText string
		if ranMap[migration.ID()] {
			statusText = "已执行"
		} else {
			statusText = "待执行"
		}

		status[i] = map[string]interface{}{
			"id":     migration.ID(),
			"name":   migration.Name(),
			"ran":    ranMap[migration.ID()],
			"status": statusText,
		}
	}

	return status, nil
}

// CreateMigrationFile 创建新的迁移文件
func (m *Migrator) CreateMigrationFile(name string) (string, error) {
	// 生成迁移ID
	id := time.Now().Format("20060102150405")

	// 转换名称为下划线格式
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")

	// 完整的迁移文件名
	filename := fmt.Sprintf("%s_%s.go", id, name)

	// 完整的迁移文件路径
	filepath := fmt.Sprintf("%s/%s", m.directory, filename)

	// 创建迁移目录
	if err := os.MkdirAll(m.directory, 0755); err != nil {
		return "", fmt.Errorf("创建迁移目录失败: %w", err)
	}

	// 使用 Pascalize 函数处理名称
	pascalName := toPascalCase(name)

	// 格式化迁移文件内容
	content := fmt.Sprintf(
		MigrationTemplate,
		id, name,
		pascalName, pascalName,
		pascalName, pascalName,
		pascalName, pascalName,
	)

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入迁移文件失败: %w", err)
	}

	return filepath, nil
}

// toPascalCase 将字符串转换为 Pascal 命名风格
func toPascalCase(s string) string {
	// 分割字符串
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// RegisterMigration 注册迁移
func RegisterMigration(id, name string, up, down func(db *gorm.DB) error) {
	// 这是一个全局注册函数，需要在初始化阶段实现
	// 此处只是占位符
}

// LoadMigrationsFromDirectory 从目录加载迁移
func (m *Migrator) LoadMigrationsFromDirectory(directory string) error {
	// 读取目录下的所有文件
	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("读取迁移目录失败: %w", err)
	}

	// 遍历文件
	for _, file := range files {
		// 仅处理Go文件
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		// 解析文件名
		parts := strings.SplitN(strings.TrimSuffix(file.Name(), ".go"), "_", 2)
		if len(parts) != 2 {
			continue
		}

		migrationID := parts[0]
		migrationName := parts[1]

		// 这里应该通过反射或导入包的方式加载迁移
		// 实现示例：
		fmt.Printf("发现迁移文件: %s, ID: %s, 名称: %s\n", file.Name(), migrationID, migrationName)

		// 实际项目中应该添加迁移加载逻辑
	}

	return nil
}

// LoadMigrationsFromConfig 从配置中加载迁移
func (m *Migrator) LoadMigrationsFromConfig(config []map[string]interface{}) error {
	// 从配置中加载迁移
	for _, migrationConfig := range config {
		id, ok := migrationConfig["id"].(string)
		if !ok {
			return ErrInvalidMigrationID
		}

		name, ok := migrationConfig["name"].(string)
		if !ok {
			name = id
		}

		upFunc, ok := migrationConfig["up"].(func(db *gorm.DB) error)
		if !ok {
			return ErrInvalidMigration
		}

		downFunc, ok := migrationConfig["down"].(func(db *gorm.DB) error)
		if !ok {
			downFunc = func(db *gorm.DB) error { return nil }
		}

		migration := NewMigration(id, name, upFunc, downFunc)
		if err := m.Register(migration); err != nil {
			return err
		}
	}

	return nil
}

// ParseMigrationID 解析迁移ID
func ParseMigrationID(id string) (time.Time, error) {
	if len(id) < 14 {
		return time.Time{}, ErrInvalidMigrationID
	}

	year, err := strconv.Atoi(id[0:4])
	if err != nil {
		return time.Time{}, err
	}

	month, err := strconv.Atoi(id[4:6])
	if err != nil {
		return time.Time{}, err
	}

	day, err := strconv.Atoi(id[6:8])
	if err != nil {
		return time.Time{}, err
	}

	hour, err := strconv.Atoi(id[8:10])
	if err != nil {
		return time.Time{}, err
	}

	minute, err := strconv.Atoi(id[10:12])
	if err != nil {
		return time.Time{}, err
	}

	second, err := strconv.Atoi(id[12:14])
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}
