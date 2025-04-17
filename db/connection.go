package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	// 使用下划线导入使得驱动程序能够自注册
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/zzliekkas/flow/config"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 数据库驱动类型常量
const (
	// MySQL 数据库
	MySQL = "mysql"
	// PostgreSQL 数据库
	PostgreSQL = "postgres"
	// SQLite 数据库
	SQLite = "sqlite"
)

// 定义错误类型
var (
	// ErrUnsupportedDriver 不支持的驱动类型错误
	ErrUnsupportedDriver = errors.New("不支持的数据库驱动类型")
	// ErrConnectionNotFound 连接未找到错误
	ErrConnectionNotFound = errors.New("数据库连接未找到")
	// ErrInvalidConfiguration 无效的数据库配置
	ErrInvalidConfiguration = errors.New("无效的数据库配置")
	// ErrDatabaseNotFound 未找到指定的数据库
	ErrDatabaseNotFound = errors.New("未找到指定的数据库连接")
	// ErrConnectionFailed 数据库连接失败
	ErrConnectionFailed = errors.New("数据库连接失败")
)

// Config 数据库配置
type Config struct {
	// 驱动类型：mysql, postgres, sqlite等
	Driver string `yaml:"driver" json:"driver"`

	// 连接信息
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`

	// 其他连接参数
	Charset  string `yaml:"charset" json:"charset"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`
	TimeZone string `yaml:"timezone" json:"timezone"`

	// 连接池配置
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`

	// 日志配置
	LogLevel      logger.LogLevel `yaml:"log_level" json:"log_level"`
	SlowThreshold time.Duration   `yaml:"slow_threshold" json:"slow_threshold"`

	// 主从配置
	Replicas []ReplicaConfig `yaml:"replicas" json:"replicas"`

	// 健康检查配置
	HealthCheck        bool          `yaml:"health_check" json:"health_check"`
	HealthCheckPeriod  time.Duration `yaml:"health_check_period" json:"health_check_period"`
	HealthCheckTimeout time.Duration `yaml:"health_check_timeout" json:"health_check_timeout"`
	HealthCheckSQL     string        `yaml:"health_check_sql" json:"health_check_sql"`
}

// ReplicaConfig 从库配置
type ReplicaConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`
	Weight   int    `yaml:"weight" json:"weight"` // 权重，用于负载均衡
}

// Manager 数据库连接管理器
type Manager struct {
	// 数据库连接映射
	connections map[string]*gorm.DB
	// 默认连接名称
	defaultConnection string
	// 配置项
	configs map[string]Config
	// 互斥锁
	mutex sync.RWMutex
	// 健康状态
	healthStatus map[string]bool
	// 健康检查上下文
	healthCtx context.Context
	// 健康检查取消函数
	healthCancel context.CancelFunc
}

// NewManager 创建数据库连接管理器
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		connections:  make(map[string]*gorm.DB),
		configs:      make(map[string]Config),
		healthStatus: make(map[string]bool),
		healthCtx:    ctx,
		healthCancel: cancel,
	}
}

// SetDefaultConnection 设置默认数据库连接
func (m *Manager) SetDefaultConnection(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.defaultConnection = name
}

// Register 注册数据库配置
func (m *Manager) Register(name string, config Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证配置
	if config.Driver == "" {
		return ErrInvalidConfiguration
	}

	// 设置默认值
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = 10
	}
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = 100
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = time.Hour
	}
	if config.HealthCheckPeriod <= 0 {
		config.HealthCheckPeriod = 30 * time.Second
	}
	if config.HealthCheckTimeout <= 0 {
		config.HealthCheckTimeout = 5 * time.Second
	}
	if config.HealthCheckSQL == "" {
		config.HealthCheckSQL = "SELECT 1"
	}

	// 保存配置
	m.configs[name] = config

	// 如果是第一个配置，设为默认
	if m.defaultConnection == "" {
		m.defaultConnection = name
	}

	return nil
}

// Connect 建立数据库连接
func (m *Manager) Connect(name string) (*gorm.DB, error) {
	m.mutex.RLock()
	db, exists := m.connections[name]
	if exists {
		m.mutex.RUnlock()
		return db, nil
	}

	config, exists := m.configs[name]
	m.mutex.RUnlock()

	if !exists {
		return nil, ErrDatabaseNotFound
	}

	// 创建新连接
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 双重检查，避免并发情况下重复创建连接
	if db, exists = m.connections[name]; exists {
		return db, nil
	}

	var err error
	db, err = m.createConnection(config)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 保存连接
	m.connections[name] = db
	m.healthStatus[name] = true

	// 启动健康检查
	if config.HealthCheck {
		go m.startHealthCheck(name, config)
	}

	return db, nil
}

// Default 获取默认数据库连接
func (m *Manager) Default() (*gorm.DB, error) {
	m.mutex.RLock()
	defaultName := m.defaultConnection
	m.mutex.RUnlock()

	if defaultName == "" {
		return nil, errors.New("未设置默认数据库连接")
	}

	return m.Connect(defaultName)
}

// Connection 获取指定名称的数据库连接
func (m *Manager) Connection(name string) (*gorm.DB, error) {
	return m.Connect(name)
}

// HasConnection 检查是否存在指定名称的数据库连接
func (m *Manager) HasConnection(name string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.connections[name]
	return exists
}

// Close 关闭所有数据库连接
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 取消健康检查
	if m.healthCancel != nil {
		m.healthCancel()
	}

	var lastErr error
	for name, db := range m.connections {
		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			continue
		}

		if err := sqlDB.Close(); err != nil {
			lastErr = err
		}

		delete(m.connections, name)
		delete(m.healthStatus, name)
	}

	return lastErr
}

// IsHealthy 检查数据库连接是否健康
func (m *Manager) IsHealthy(name string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status, exists := m.healthStatus[name]
	if !exists {
		return false
	}

	return status
}

// AllHealthStatus 获取所有数据库连接的健康状态
func (m *Manager) AllHealthStatus() map[string]bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 复制一份状态信息，避免外部修改
	status := make(map[string]bool, len(m.healthStatus))
	for k, v := range m.healthStatus {
		status[k] = v
	}

	return status
}

// createConnection 根据配置创建数据库连接
func (m *Manager) createConnection(config Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case MySQL:
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
			config.Charset,
			config.TimeZone,
		)
		dialector = mysqldriver.Open(dsn)

	case PostgreSQL:
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			config.Host,
			config.Port,
			config.Username,
			config.Password,
			config.Database,
			config.SSLMode,
			config.TimeZone,
		)
		dialector = postgres.Open(dsn)

	case SQLite:
		dialector = sqlite.Open(config.Database)

	default:
		return nil, ErrUnsupportedDriver
	}

	// 创建日志配置
	logConfig := logger.Config{
		SlowThreshold:             config.SlowThreshold,
		LogLevel:                  config.LogLevel,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	}

	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // 使用标准日志输出
			logConfig,
		),
	}

	return gorm.Open(dialector, gormConfig)
}

// startHealthCheck 启动数据库健康检查
func (m *Manager) startHealthCheck(name string, config Config) {
	ticker := time.NewTicker(config.HealthCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			healthy := m.checkHealth(name, config)
			m.mutex.Lock()
			m.healthStatus[name] = healthy
			m.mutex.Unlock()

		case <-m.healthCtx.Done():
			return
		}
	}
}

// checkHealth 检查数据库连接健康状态
func (m *Manager) checkHealth(name string, config Config) bool {
	m.mutex.RLock()
	db, exists := m.connections[name]
	m.mutex.RUnlock()

	if !exists {
		return false
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false
	}

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), config.HealthCheckTimeout)
	defer cancel()

	// 执行健康检查SQL
	return sqlDB.PingContext(ctx) == nil
}

// FromConfig 从配置中加载数据库配置
func (m *Manager) FromConfig(configManager *config.Manager) error {
	if configManager == nil {
		return errors.New("配置管理器不能为空")
	}

	// 获取默认连接名称
	defaultConn := configManager.GetString("database.default")
	if defaultConn != "" {
		m.SetDefaultConnection(defaultConn)
	}

	// 获取所有连接配置
	connectionsMap := configManager.GetStringMap("database.connections")
	if len(connectionsMap) == 0 {
		return nil // 没有配置数据库连接，可能是不需要数据库的应用
	}

	// 遍历并注册数据库配置
	for name, configData := range connectionsMap {
		if configMap, ok := configData.(map[string]interface{}); ok {
			config := Config{
				Driver:   getString(configMap, "driver", ""),
				Host:     getString(configMap, "host", "localhost"),
				Port:     getInt(configMap, "port", 3306),
				Database: getString(configMap, "database", ""),
				Username: getString(configMap, "username", ""),
				Password: getString(configMap, "password", ""),
				Charset:  getString(configMap, "charset", "utf8mb4"),
				SSLMode:  getString(configMap, "sslmode", "disable"),
				TimeZone: getString(configMap, "timezone", "Local"),

				MaxIdleConns:    getInt(configMap, "max_idle_conns", 10),
				MaxOpenConns:    getInt(configMap, "max_open_conns", 100),
				ConnMaxLifetime: getDuration(configMap, "conn_max_lifetime", time.Hour),
				ConnMaxIdleTime: getDuration(configMap, "conn_max_idle_time", time.Hour),

				HealthCheck:        getBool(configMap, "health_check", true),
				HealthCheckPeriod:  getDuration(configMap, "health_check_period", 30*time.Second),
				HealthCheckTimeout: getDuration(configMap, "health_check_timeout", 5*time.Second),
				HealthCheckSQL:     getString(configMap, "health_check_sql", "SELECT 1"),
			}

			// 从库配置
			if replicasData, exists := configMap["replicas"]; exists {
				if replicasList, ok := replicasData.([]interface{}); ok {
					for _, replicaData := range replicasList {
						if replicaMap, ok := replicaData.(map[string]interface{}); ok {
							replica := ReplicaConfig{
								Host:     getString(replicaMap, "host", config.Host),
								Port:     getInt(replicaMap, "port", config.Port),
								Username: getString(replicaMap, "username", config.Username),
								Password: getString(replicaMap, "password", config.Password),
								SSLMode:  getString(replicaMap, "sslmode", config.SSLMode),
								Weight:   getInt(replicaMap, "weight", 1),
							}
							config.Replicas = append(config.Replicas, replica)
						}
					}
				}
			}

			if err := m.Register(name, config); err != nil {
				return err
			}
		}
	}

	return nil
}

// 辅助函数：从map中获取字符串值
func getString(m map[string]interface{}, key, defaultValue string) string {
	if val, exists := m[key]; exists {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// 辅助函数：从map中获取整数值
func getInt(m map[string]interface{}, key string, defaultValue int) int {
	if val, exists := m[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// 辅助函数：从map中获取布尔值
func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := m[key]; exists {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// 辅助函数：从map中获取时间间隔值
func getDuration(m map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	if val, exists := m[key]; exists {
		switch v := val.(type) {
		case time.Duration:
			return v
		case int:
			return time.Duration(v) * time.Second
		case int64:
			return time.Duration(v) * time.Second
		case float64:
			return time.Duration(v) * time.Second
		case string:
			if duration, err := time.ParseDuration(v); err == nil {
				return duration
			}
		}
	}
	return defaultValue
}
