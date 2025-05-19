package config

// Manager 配置管理器类型别名，与Config相同
// 这是为了向后兼容，新代码应直接使用Config类型
type Manager = Config

// Kd100Config 定义快递100新版配置
type Kd100Config struct {
	Key       string `mapstructure:"key" yaml:"key"`
	Customer  string `mapstructure:"customer" yaml:"customer"`
	Secret    string `mapstructure:"secret" yaml:"secret"`
	BaseURL   string `mapstructure:"baseUrl" yaml:"baseUrl"`
	Salt      string `mapstructure:"salt" yaml:"salt"`
	NotifyURL string `mapstructure:"notifyUrl" yaml:"notifyUrl"`
	Timeout   int    `mapstructure:"timeout" yaml:"timeout"`
}
