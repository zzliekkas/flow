package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// 定义种子数据相关错误
var (
	ErrSeederFailed   = errors.New("种子数据执行失败")
	ErrSeederNotFound = errors.New("找不到指定的种子数据")
	ErrSeederExists   = errors.New("种子数据已存在")
	ErrInvalidSeeder  = errors.New("无效的种子数据")
)

// Seeder 种子数据接口
type Seeder interface {
	// Name 获取种子数据名称
	Name() string
	// Run 执行种子数据填充
	Run(db *gorm.DB) error
	// Dependencies 获取依赖的其他种子数据
	Dependencies() []string
	// SetOrder 设置执行顺序
	SetOrder(order int)
	// GetOrder 获取执行顺序
	GetOrder() int
}

// BaseSeeder 基础种子数据实现
type BaseSeeder struct {
	name         string
	run          func(db *gorm.DB) error
	dependencies []string
	order        int
}

// NewSeeder 创建新的种子数据
func NewSeeder(name string, run func(db *gorm.DB) error, dependencies ...string) Seeder {
	return &BaseSeeder{
		name:         name,
		run:          run,
		dependencies: dependencies,
		order:        0,
	}
}

// Name 获取种子数据名称
func (s *BaseSeeder) Name() string {
	return s.name
}

// Run 执行种子数据填充
func (s *BaseSeeder) Run(db *gorm.DB) error {
	if s.run == nil {
		return nil
	}
	return s.run(db)
}

// Dependencies 获取依赖的其他种子数据
func (s *BaseSeeder) Dependencies() []string {
	return s.dependencies
}

// SetOrder 设置执行顺序
func (s *BaseSeeder) SetOrder(order int) {
	s.order = order
}

// GetOrder 获取执行顺序
func (s *BaseSeeder) GetOrder() int {
	return s.order
}

// ModelSeeder 模型种子数据，用于从JSON/YAML文件填充数据
type ModelSeeder struct {
	BaseSeeder
	model        interface{}
	dataFile     string
	truncate     bool
	ignoreErrors bool
}

// NewModelSeeder 创建模型种子数据
func NewModelSeeder(name string, model interface{}, dataFile string, opts ...ModelSeederOption) *ModelSeeder {
	seeder := &ModelSeeder{
		BaseSeeder: BaseSeeder{
			name:         name,
			dependencies: []string{},
			order:        0,
		},
		model:        model,
		dataFile:     dataFile,
		truncate:     false,
		ignoreErrors: false,
	}

	// 应用选项
	for _, opt := range opts {
		opt(seeder)
	}

	// 设置运行函数
	seeder.run = seeder.runModelSeeder

	return seeder
}

// ModelSeederOption 模型种子数据选项
type ModelSeederOption func(*ModelSeeder)

// WithTruncate 设置是否在填充前清空表
func WithTruncate(truncate bool) ModelSeederOption {
	return func(s *ModelSeeder) {
		s.truncate = truncate
	}
}

// WithDependencies 设置依赖的其他种子数据
func WithDependencies(dependencies ...string) ModelSeederOption {
	return func(s *ModelSeeder) {
		s.dependencies = dependencies
	}
}

// WithIgnoreErrors 设置是否忽略错误
func WithIgnoreErrors(ignore bool) ModelSeederOption {
	return func(s *ModelSeeder) {
		s.ignoreErrors = ignore
	}
}

// WithOrder 设置执行顺序
func WithOrder(order int) ModelSeederOption {
	return func(s *ModelSeeder) {
		s.order = order
	}
}

// runModelSeeder 执行模型种子数据填充
func (s *ModelSeeder) runModelSeeder(db *gorm.DB) error {
	// 检查模型
	modelType := reflect.TypeOf(s.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 获取表名
	tableName := ""
	if tableNamer, ok := s.model.(interface{ TableName() string }); ok {
		tableName = tableNamer.TableName()
	} else {
		// 使用GORM的默认命名规则
		tableName = strings.ToLower(modelType.Name()) + "s"
	}

	// 如果需要，清空表
	if s.truncate {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName)).Error; err != nil {
			return fmt.Errorf("truncate table %s failed: %w", tableName, err)
		}
	}

	// 读取数据文件
	data, err := s.readDataFile()
	if err != nil {
		return fmt.Errorf("read data file failed: %w", err)
	}

	// 创建模型切片
	sliceType := reflect.SliceOf(modelType)
	modelsValue := reflect.New(sliceType)
	models := modelsValue.Interface()

	// 解析数据到模型切片
	if err := s.parseData(data, models); err != nil {
		return fmt.Errorf("parse data failed: %w", err)
	}

	// 执行批量插入
	return db.CreateInBatches(reflect.ValueOf(models).Elem().Interface(), 100).Error
}

// readDataFile 读取数据文件
func (s *ModelSeeder) readDataFile() ([]byte, error) {
	file, err := os.Open(s.dataFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// parseData 解析数据到模型切片
func (s *ModelSeeder) parseData(data []byte, models interface{}) error {
	ext := strings.ToLower(filepath.Ext(s.dataFile))

	switch ext {
	case ".json":
		return json.Unmarshal(data, models)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, models)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

// SeederRecord 种子数据记录
type SeederRecord struct {
	Name      string    `gorm:"primaryKey;size:255"`
	Batch     int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

// TableName 表名
func (SeederRecord) TableName() string {
	return "seeders"
}

// SeederManager 种子数据管理器
type SeederManager struct {
	db      *gorm.DB
	seeders map[string]Seeder
}

// NewSeederManager 创建种子数据管理器
func NewSeederManager(db *gorm.DB) *SeederManager {
	return &SeederManager{
		db:      db,
		seeders: make(map[string]Seeder),
	}
}

// Register 注册种子数据
func (m *SeederManager) Register(seeder Seeder) error {
	name := seeder.Name()
	if name == "" {
		return ErrInvalidSeeder
	}

	if _, exists := m.seeders[name]; exists {
		return ErrSeederExists
	}

	m.seeders[name] = seeder
	return nil
}

// ensureTable 确保种子数据表存在
func (m *SeederManager) ensureTable() error {
	return m.db.AutoMigrate(&SeederRecord{})
}

// GetRan 获取已执行的种子数据
func (m *SeederManager) GetRan() ([]string, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	var records []SeederRecord
	if err := m.db.Order("name").Find(&records).Error; err != nil {
		return nil, err
	}

	ran := make([]string, len(records))
	for i, record := range records {
		ran[i] = record.Name
	}

	return ran, nil
}

// GetSeeders 获取所有种子数据
func (m *SeederManager) GetSeeders() []Seeder {
	seeders := make([]Seeder, 0, len(m.seeders))
	for _, seeder := range m.seeders {
		seeders = append(seeders, seeder)
	}

	// 按照顺序排序
	sort.Slice(seeders, func(i, j int) bool {
		if seeders[i].GetOrder() == seeders[j].GetOrder() {
			return seeders[i].Name() < seeders[j].Name()
		}
		return seeders[i].GetOrder() < seeders[j].GetOrder()
	})

	return seeders
}

// GetPending 获取待执行的种子数据
func (m *SeederManager) GetPending() ([]Seeder, error) {
	ran, err := m.GetRan()
	if err != nil {
		return nil, err
	}

	// 创建已执行种子数据的集合
	ranSet := make(map[string]bool)
	for _, name := range ran {
		ranSet[name] = true
	}

	// 获取所有种子数据
	all := m.GetSeeders()

	// 过滤出未执行的种子数据
	pending := make([]Seeder, 0)
	for _, seeder := range all {
		if !ranSet[seeder.Name()] {
			pending = append(pending, seeder)
		}
	}

	return pending, nil
}

// Run 执行所有待执行的种子数据
func (m *SeederManager) Run() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	// 获取待执行的种子数据
	pending, err := m.GetPending()
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		return nil // 没有种子数据需要执行
	}

	// 获取当前批次号
	var batch int
	var maxBatch struct {
		MaxBatch int
	}
	if err := m.db.Model(&SeederRecord{}).Select("COALESCE(MAX(batch), 0) as max_batch").Scan(&maxBatch).Error; err != nil {
		return err
	}
	batch = maxBatch.MaxBatch + 1

	// 构建依赖图
	graph := make(map[string][]string)
	for _, seeder := range pending {
		graph[seeder.Name()] = seeder.Dependencies()
	}

	// 执行种子数据
	executed := make(map[string]bool)
	for _, seeder := range pending {
		if err := m.runSeeder(seeder, graph, executed, batch); err != nil {
			return err
		}
	}

	return nil
}

// runSeeder 执行单个种子数据(带依赖处理)
func (m *SeederManager) runSeeder(seeder Seeder, graph map[string][]string, executed map[string]bool, batch int) error {
	name := seeder.Name()

	// 如果已经执行过，直接返回
	if executed[name] {
		return nil
	}

	// 检查循环依赖
	visited := make(map[string]bool)
	if m.hasCyclicDependency(name, graph, visited, make(map[string]bool)) {
		return fmt.Errorf("cyclic dependency detected for seeder %s", name)
	}

	// 先执行依赖
	for _, dep := range seeder.Dependencies() {
		depSeeder, exists := m.seeders[dep]
		if !exists {
			return fmt.Errorf("dependency %s not found for seeder %s: %w", dep, name, ErrSeederNotFound)
		}

		if err := m.runSeeder(depSeeder, graph, executed, batch); err != nil {
			return err
		}
	}

	// 执行当前种子数据
	err := m.db.Transaction(func(tx *gorm.DB) error {
		// 执行种子数据
		if err := seeder.Run(tx); err != nil {
			return err
		}

		// 记录执行结果
		return tx.Create(&SeederRecord{
			Name:      name,
			Batch:     batch,
			CreatedAt: time.Now(),
		}).Error
	})

	if err != nil {
		return fmt.Errorf("failed to run seeder %s: %w", name, err)
	}

	// 标记为已执行
	executed[name] = true

	return nil
}

// hasCyclicDependency 检测循环依赖
func (m *SeederManager) hasCyclicDependency(name string, graph map[string][]string, visited, recStack map[string]bool) bool {
	if !visited[name] {
		visited[name] = true
		recStack[name] = true

		for _, dep := range graph[name] {
			if !visited[dep] && m.hasCyclicDependency(dep, graph, visited, recStack) {
				return true
			} else if recStack[dep] {
				return true
			}
		}
	}

	recStack[name] = false
	return false
}

// Reset 重置所有种子数据
func (m *SeederManager) Reset() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	// 清空种子数据记录表
	if err := m.db.Exec("DELETE FROM seeders").Error; err != nil {
		return err
	}

	return nil
}

// RunSeeder 执行指定名称的种子数据
func (m *SeederManager) RunSeeder(name string) error {
	seeder, exists := m.seeders[name]
	if !exists {
		return ErrSeederNotFound
	}

	// 获取当前批次号
	var batch int
	var maxBatch struct {
		MaxBatch int
	}
	if err := m.db.Model(&SeederRecord{}).Select("COALESCE(MAX(batch), 0) as max_batch").Scan(&maxBatch).Error; err != nil {
		return err
	}
	batch = maxBatch.MaxBatch + 1

	// 执行种子数据
	executed := make(map[string]bool)
	graph := make(map[string][]string)
	for _, s := range m.seeders {
		graph[s.Name()] = s.Dependencies()
	}

	return m.runSeeder(seeder, graph, executed, batch)
}

// LoadSeedersFromDirectory 从目录加载种子数据文件
func (m *SeederManager) LoadSeedersFromDirectory(directory string, modelType interface{}) error {
	// 检查目录是否存在
	info, err := os.Stat(directory)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", directory)
	}

	// 获取模型类型
	typ := reflect.TypeOf(modelType)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// 读取目录中的所有文件
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	// 遍历文件
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查文件扩展名
		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}

		// 获取种子数据名称
		name := strings.TrimSuffix(filename, ext)

		// 创建模型实例
		model := reflect.New(typ).Interface()

		// 创建种子数据
		seeder := NewModelSeeder(
			name,
			model,
			filepath.Join(directory, filename),
			WithTruncate(false),
		)

		// 注册种子数据
		if err := m.Register(seeder); err != nil {
			return err
		}
	}

	return nil
}

// Status 获取种子数据状态
func (m *SeederManager) Status() ([]map[string]interface{}, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	// 获取所有种子数据
	seeders := m.GetSeeders()

	// 获取已执行的种子数据
	ran, err := m.GetRan()
	if err != nil {
		return nil, err
	}

	// 创建已执行种子数据的集合
	ranMap := make(map[string]bool)
	for _, name := range ran {
		ranMap[name] = true
	}

	// 构建状态信息
	status := make([]map[string]interface{}, len(seeders))
	for i, seeder := range seeders {
		var statusText string
		if ranMap[seeder.Name()] {
			statusText = "已执行"
		} else {
			statusText = "待执行"
		}

		status[i] = map[string]interface{}{
			"name":         seeder.Name(),
			"ran":          ranMap[seeder.Name()],
			"status":       statusText,
			"order":        seeder.GetOrder(),
			"dependencies": seeder.Dependencies(),
		}
	}

	return status, nil
}

// RegisterSeeder 注册种子数据(全局函数)
func RegisterSeeder(manager *SeederManager, seeder Seeder) error {
	return manager.Register(seeder)
}
