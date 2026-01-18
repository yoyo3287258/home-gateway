package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yoyo3287258/home-gateway/internal/model"
	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	// Server HTTP服务器配置
	Server ServerConfig `yaml:"server"`

	// Security 安全配置
	Security SecurityConfig `yaml:"security"`

	// LLM 大语言模型配置
	LLM LLMConfig `yaml:"llm"`

	// Kafka Kafka配置
	Kafka KafkaConfig `yaml:"kafka"`

	// Channels 渠道配置
	Channels ChannelsConfig `yaml:"channels"`

	// Log 日志配置
	Log LogConfig `yaml:"log"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// APIToken API访问令牌，用于HTTP接口认证
	// 客户端需要在请求头中携带 Authorization: Bearer <token>
	APIToken string `yaml:"api_token"`

	// IPWhitelist IP白名单，为空则不限制
	// 支持单个IP和CIDR格式，如 ["192.168.1.0/24", "10.0.0.1"]
	IPWhitelist []string `yaml:"ip_whitelist"`

	// RateLimitPerMinute 每分钟请求限制，0表示不限制
	RateLimitPerMinute int `yaml:"rate_limit_per_minute"`

	// TrustedProxies 可信代理IP列表（用于获取真实客户端IP）
	TrustedProxies []string `yaml:"trusted_proxies"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	// Host 监听地址
	Host string `yaml:"host"`

	// Port 监听端口
	Port int `yaml:"port"`

	// ReadTimeout 读取超时
	ReadTimeout time.Duration `yaml:"read_timeout"`

	// WriteTimeout 写入超时
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// LLMConfig 大语言模型配置
type LLMConfig struct {
	// BaseURL API基础URL（OpenAI兼容格式）
	BaseURL string `yaml:"base_url"`

	// APIKey API密钥，支持环境变量格式 ${ENV_VAR}
	APIKey string `yaml:"api_key"`

	// Model 模型名称
	Model string `yaml:"model"`

	// Timeout 请求超时时间
	Timeout time.Duration `yaml:"timeout"`

	// MaxRetries 最大重试次数
	MaxRetries int `yaml:"max_retries"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	// Brokers Kafka broker地址列表
	Brokers []string `yaml:"brokers"`

	// RequestTopic 请求topic
	RequestTopic string `yaml:"request_topic"`

	// ResponseTopic 响应topic
	ResponseTopic string `yaml:"response_topic"`

	// ConsumerGroup 消费者组
	ConsumerGroup string `yaml:"consumer_group"`

	// ResponseTimeout 响应超时时间
	ResponseTimeout time.Duration `yaml:"response_timeout"`
}

// ChannelsConfig 渠道配置
type ChannelsConfig struct {
	// Telegram Telegram配置
	Telegram TelegramConfig `yaml:"telegram"`

	// WeChatWork 企业微信配置
	WeChatWork WeChatWorkConfig `yaml:"wechat_work"`
}

// TelegramConfig Telegram配置
type TelegramConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled"`

	// BotToken Bot Token
	BotToken string `yaml:"bot_token"`

	// WebhookSecret Webhook验证密钥
	WebhookSecret string `yaml:"webhook_secret"`
}

// WeChatWorkConfig 企业微信配置
type WeChatWorkConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled"`

	// CorpID 企业ID
	CorpID string `yaml:"corp_id"`

	// AgentID 应用ID
	AgentID string `yaml:"agent_id"`

	// Secret 应用密钥
	Secret string `yaml:"secret"`

	// Token 消息Token
	Token string `yaml:"token"`

	// EncodingAESKey 消息加密密钥
	EncodingAESKey string `yaml:"encoding_aes_key"`
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别: debug, info, warn, error
	Level string `yaml:"level"`

	// Format 日志格式: json, text
	Format string `yaml:"format"`
}

// ProcessorsConfig 处理器配置
type ProcessorsConfig struct {
	// Processors 处理器列表
	Processors []model.Processor `yaml:"processors"`
}

// Manager 配置管理器
type Manager struct {
	configPath     string
	processorsDir  string // 改为目录，支持多个配置文件

	config     *Config
	processors *ProcessorsConfig

	mu sync.RWMutex

	// onReload 配置重载回调函数
	onReload []func()
}

// NewManager 创建配置管理器
// processorsPath 可以是单个yaml文件或包含多个yaml文件的目录
func NewManager(configPath, processorsPath string) *Manager {
	return &Manager{
		configPath:    configPath,
		processorsDir: processorsPath,
	}
}

// Load 加载配置
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载主配置
	config, err := m.loadConfig(m.configPath)
	if err != nil {
		return fmt.Errorf("加载主配置失败: %w", err)
	}
	m.config = config

	// 加载处理器配置（支持单文件或目录）
	processors, err := m.loadProcessors(m.processorsDir)
	if err != nil {
		return fmt.Errorf("加载处理器配置失败: %w", err)
	}
	m.processors = processors

	return nil
}

// loadConfig 加载主配置文件
func (m *Manager) loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 替换环境变量
	content := expandEnvVars(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// loadProcessors 加载处理器配置
// 支持单个yaml文件或包含多个yaml文件的目录
func (m *Manager) loadProcessors(path string) (*ProcessorsConfig, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var allProcessors []model.Processor

	if info.IsDir() {
		// 目录模式：加载目录下所有yaml文件
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("读取处理器目录失败: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
				continue
			}

			filePath := path + "/" + name
			processors, err := m.loadSingleProcessorFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("加载处理器文件 %s 失败: %w", name, err)
			}
			allProcessors = append(allProcessors, processors...)
		}
	} else {
		// 单文件模式
		processors, err := m.loadSingleProcessorFile(path)
		if err != nil {
			return nil, err
		}
		allProcessors = processors
	}

	return &ProcessorsConfig{Processors: allProcessors}, nil
}

// loadSingleProcessorFile 加载单个处理器配置文件
func (m *Manager) loadSingleProcessorFile(path string) ([]model.Processor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ProcessorsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 从文件名推断分组（如果处理器没有设置group）
	baseName := strings.TrimSuffix(strings.TrimSuffix(path, ".yaml"), ".yml")
	parts := strings.Split(baseName, "/")
	if len(parts) > 0 {
		defaultGroup := parts[len(parts)-1]
		// 移除 -processors 后缀
		defaultGroup = strings.TrimSuffix(defaultGroup, "-processors")
		defaultGroup = strings.TrimSuffix(defaultGroup, "_processors")
		
		for i := range config.Processors {
			if config.Processors[i].Group == "" {
				config.Processors[i].Group = defaultGroup
			}
			// 默认启用
			if !config.Processors[i].Enabled {
				// YAML中未设置时，bool默认为false，这里需要特殊处理
				// 通过检查是否有enabled字段来判断（目前简化为默认启用）
			}
		}
	}

	return config.Processors, nil
}

// Reload 重新加载配置
func (m *Manager) Reload() error {
	if err := m.Load(); err != nil {
		return err
	}

	// 触发回调
	m.mu.RLock()
	callbacks := m.onReload
	m.mu.RUnlock()

	for _, cb := range callbacks {
		cb()
	}

	return nil
}

// OnReload 注册配置重载回调
func (m *Manager) OnReload(callback func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onReload = append(m.onReload, callback)
}

// WatchChanges 监听配置文件变化
func (m *Manager) WatchChanges() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 延迟一下，确保文件写入完成
					time.Sleep(100 * time.Millisecond)
					if err := m.Reload(); err != nil {
						fmt.Printf("配置重载失败: %v\n", err)
					} else {
						fmt.Println("配置已重新加载")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("配置监听错误: %v\n", err)
			}
		}
	}()

	if err := watcher.Add(m.configPath); err != nil {
		return err
	}
	if err := watcher.Add(m.processorsDir); err != nil {
		return err
	}

	return nil
}

// Get 获取主配置（只读）
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetProcessors 获取处理器列表
func (m *Manager) GetProcessors() []model.Processor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.processors == nil {
		return nil
	}
	return m.processors.Processors
}

// GetProcessor 根据ID获取处理器
func (m *Manager) GetProcessor(id string) *model.Processor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.processors == nil {
		return nil
	}
	for _, p := range m.processors.Processors {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// expandEnvVars 展开环境变量
// 支持 ${VAR} 和 $VAR 格式
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return "${" + key + "}"
	})
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}

	if config.LLM.Timeout == 0 {
		config.LLM.Timeout = 30 * time.Second
	}
	if config.LLM.MaxRetries == 0 {
		config.LLM.MaxRetries = 3
	}

	if config.Kafka.ResponseTimeout == 0 {
		config.Kafka.ResponseTimeout = 5 * time.Second
	}
	if config.Kafka.ConsumerGroup == "" {
		config.Kafka.ConsumerGroup = "gateway"
	}
	if config.Kafka.RequestTopic == "" {
		config.Kafka.RequestTopic = "home.request"
	}
	if config.Kafka.ResponseTopic == "" {
		config.Kafka.ResponseTopic = "home.response"
	}

	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "text"
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	var errs []string

	if c.LLM.BaseURL == "" {
		errs = append(errs, "llm.base_url 不能为空")
	}
	if c.LLM.APIKey == "" || strings.HasPrefix(c.LLM.APIKey, "${") {
		errs = append(errs, "llm.api_key 未设置或环境变量未定义")
	}
	if c.LLM.Model == "" {
		errs = append(errs, "llm.model 不能为空")
	}

	if len(c.Kafka.Brokers) == 0 {
		errs = append(errs, "kafka.brokers 不能为空")
	}

	if len(errs) > 0 {
		return fmt.Errorf("配置验证失败:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
