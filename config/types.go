package config

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

type StripeConfig struct {
	APIKey        string `yaml:"api_key"`
	WebhookSecret string `yaml:"webhook_secret"`
}

type PayPalConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type AlipayConfig struct {
	AppID      string `yaml:"app_id"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
}

type WeChatPayConfig struct {
	MchID    string `yaml:"mch_id"`
	APIKey   string `yaml:"api_key"`
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
}

type AppConfig struct {
	// ...其他配置...
	Stripe    StripeConfig    `yaml:"stripe"`
	Paypal    PayPalConfig    `yaml:"paypal"`
	Alipay    AlipayConfig    `yaml:"alipay"`
	WeChatPay WeChatPayConfig `yaml:"wechatpay"`
}
