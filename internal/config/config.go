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

// Config 涓婚厤缃粨鏋?
type Config struct {
	// Server HTTP鏈嶅姟鍣ㄩ厤缃?
	Server ServerConfig `yaml:"server"`

	// Security 瀹夊叏閰嶇疆
	Security SecurityConfig `yaml:"security"`

	// LLM 澶ц瑷€妯″瀷閰嶇疆
	LLM LLMConfig `yaml:"llm"`

	// Kafka Kafka閰嶇疆
	Kafka KafkaConfig `yaml:"kafka"`

	// Channels 娓犻亾閰嶇疆
	Channels ChannelsConfig `yaml:"channels"`

	// Log 鏃ュ織閰嶇疆
	Log LogConfig `yaml:"log"`
}

// SecurityConfig 瀹夊叏閰嶇疆
type SecurityConfig struct {
	// APIToken API璁块棶浠ょ墝锛岀敤浜嶩TTP鎺ュ彛璁よ瘉
	// 瀹㈡埛绔渶瑕佸湪璇锋眰澶翠腑鎼哄甫 Authorization: Bearer <token>
	APIToken string `yaml:"api_token"`

	// IPWhitelist IP鐧藉悕鍗曪紝涓虹┖鍒欎笉闄愬埗
	// 鏀寔鍗曚釜IP鍜孋IDR鏍煎紡锛屽 ["192.168.1.0/24", "10.0.0.1"]
	IPWhitelist []string `yaml:"ip_whitelist"`

	// RateLimitPerMinute 姣忓垎閽熻姹傞檺鍒讹紝0琛ㄧず涓嶉檺鍒?
	RateLimitPerMinute int `yaml:"rate_limit_per_minute"`

	// TrustedProxies 鍙俊浠ｇ悊IP鍒楄〃锛堢敤浜庤幏鍙栫湡瀹炲鎴风IP锛?
	TrustedProxies []string `yaml:"trusted_proxies"`
}

// ServerConfig HTTP鏈嶅姟鍣ㄩ厤缃?
type ServerConfig struct {
	// Host 鐩戝惉鍦板潃
	Host string `yaml:"host"`

	// Port 鐩戝惉绔彛
	Port int `yaml:"port"`

	// ReadTimeout 璇诲彇瓒呮椂
	ReadTimeout time.Duration `yaml:"read_timeout"`

	// WriteTimeout 鍐欏叆瓒呮椂
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// LLMConfig 澶ц瑷€妯″瀷閰嶇疆
type LLMConfig struct {
	// BaseURL API鍩虹URL锛圤penAI鍏煎鏍煎紡锛?
	BaseURL string `yaml:"base_url"`

	// APIKey API瀵嗛挜锛屾敮鎸佺幆澧冨彉閲忔牸寮?${ENV_VAR}
	APIKey string `yaml:"api_key"`

	// Model 妯″瀷鍚嶇О
	Model string `yaml:"model"`

	// Timeout 璇锋眰瓒呮椂鏃堕棿
	Timeout time.Duration `yaml:"timeout"`

	// MaxRetries 鏈€澶ч噸璇曟鏁?
	MaxRetries int `yaml:"max_retries"`
}

// KafkaConfig Kafka閰嶇疆
type KafkaConfig struct {
	// Brokers Kafka broker鍦板潃鍒楄〃
	Brokers []string `yaml:"brokers"`

	// RequestTopic 璇锋眰topic
	RequestTopic string `yaml:"request_topic"`

	// ResponseTopic 鍝嶅簲topic
	ResponseTopic string `yaml:"response_topic"`

	// ConsumerGroup 娑堣垂鑰呯粍
	ConsumerGroup string `yaml:"consumer_group"`

	// ResponseTimeout 鍝嶅簲瓒呮椂鏃堕棿
	ResponseTimeout time.Duration `yaml:"response_timeout"`
}

// ChannelsConfig 娓犻亾閰嶇疆
type ChannelsConfig struct {
	// Telegram Telegram閰嶇疆
	Telegram TelegramConfig `yaml:"telegram"`

	// WeChatWork 浼佷笟寰俊閰嶇疆
	WeChatWork WeChatWorkConfig `yaml:"wechat_work"`
}

// TelegramConfig Telegram閰嶇疆
type TelegramConfig struct {
	// Enabled 鏄惁鍚敤
	Enabled bool `yaml:"enabled"`

	// BotToken Bot Token
	BotToken string `yaml:"bot_token"`

	// WebhookSecret Webhook楠岃瘉瀵嗛挜
	WebhookSecret string `yaml:"webhook_secret"`
}

// WeChatWorkConfig 浼佷笟寰俊閰嶇疆
type WeChatWorkConfig struct {
	// Enabled 鏄惁鍚敤
	Enabled bool `yaml:"enabled"`

	// CorpID 浼佷笟ID
	CorpID string `yaml:"corp_id"`

	// AgentID 搴旂敤ID
	AgentID string `yaml:"agent_id"`

	// Secret 搴旂敤瀵嗛挜
	Secret string `yaml:"secret"`

	// Token 娑堟伅Token
	Token string `yaml:"token"`

	// EncodingAESKey 娑堟伅鍔犲瘑瀵嗛挜
	EncodingAESKey string `yaml:"encoding_aes_key"`
}

// LogConfig 鏃ュ織閰嶇疆
type LogConfig struct {
	// Level 鏃ュ織绾у埆: debug, info, warn, error
	Level string `yaml:"level"`

	// Format 鏃ュ織鏍煎紡: json, text
	Format string `yaml:"format"`
}

// ProcessorsConfig 澶勭悊鍣ㄩ厤缃?
type ProcessorsConfig struct {
	// Processors 澶勭悊鍣ㄥ垪琛?
	Processors []model.Processor `yaml:"processors"`
}

// Manager 閰嶇疆绠＄悊鍣?
type Manager struct {
	configPath     string
	processorsDir  string // 鏀逛负鐩綍锛屾敮鎸佸涓厤缃枃浠?

	config     *Config
	processors *ProcessorsConfig

	mu sync.RWMutex

	// onReload 閰嶇疆閲嶈浇鍥炶皟鍑芥暟
	onReload []func()
}

// NewManager 鍒涘缓閰嶇疆绠＄悊鍣?
// processorsPath 鍙互鏄崟涓獃aml鏂囦欢鎴栧寘鍚涓獃aml鏂囦欢鐨勭洰褰?
func NewManager(configPath, processorsPath string) *Manager {
	return &Manager{
		configPath:    configPath,
		processorsDir: processorsPath,
	}
}

// Load 鍔犺浇閰嶇疆
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 鍔犺浇涓婚厤缃?
	config, err := m.loadConfig(m.configPath)
	if err != nil {
		return fmt.Errorf("鍔犺浇涓婚厤缃け璐? %w", err)
	}
	m.config = config

	// 鍔犺浇澶勭悊鍣ㄩ厤缃紙鏀寔鍗曟枃浠舵垨鐩綍锛?
	processors, err := m.loadProcessors(m.processorsDir)
	if err != nil {
		return fmt.Errorf("鍔犺浇澶勭悊鍣ㄩ厤缃け璐? %w", err)
	}
	m.processors = processors

	return nil
}

// loadConfig 鍔犺浇涓婚厤缃枃浠?
func (m *Manager) loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 鏇挎崲鐜鍙橀噺
	content := expandEnvVars(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}

	// 璁剧疆榛樿鍊?
	setDefaults(&config)

	return &config, nil
}

// loadProcessors 鍔犺浇澶勭悊鍣ㄩ厤缃?
// 鏀寔鍗曚釜yaml鏂囦欢鎴栧寘鍚涓獃aml鏂囦欢鐨勭洰褰?
func (m *Manager) loadProcessors(path string) (*ProcessorsConfig, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var allProcessors []model.Processor

	if info.IsDir() {
		// 鐩綍妯″紡锛氬姞杞界洰褰曚笅鎵€鏈墆aml鏂囦欢
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("璇诲彇澶勭悊鍣ㄧ洰褰曞け璐? %w", err)
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
				return nil, fmt.Errorf("鍔犺浇澶勭悊鍣ㄦ枃浠?%s 澶辫触: %w", name, err)
			}
			allProcessors = append(allProcessors, processors...)
		}
	} else {
		// 鍗曟枃浠舵ā寮?
		processors, err := m.loadSingleProcessorFile(path)
		if err != nil {
			return nil, err
		}
		allProcessors = processors
	}

	return &ProcessorsConfig{Processors: allProcessors}, nil
}

// loadSingleProcessorFile 鍔犺浇鍗曚釜澶勭悊鍣ㄩ厤缃枃浠?
func (m *Manager) loadSingleProcessorFile(path string) ([]model.Processor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ProcessorsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 浠庢枃浠跺悕鎺ㄦ柇鍒嗙粍锛堝鏋滃鐞嗗櫒娌℃湁璁剧疆group锛?
	baseName := strings.TrimSuffix(strings.TrimSuffix(path, ".yaml"), ".yml")
	parts := strings.Split(baseName, "/")
	if len(parts) > 0 {
		defaultGroup := parts[len(parts)-1]
		// 绉婚櫎 -processors 鍚庣紑
		defaultGroup = strings.TrimSuffix(defaultGroup, "-processors")
		defaultGroup = strings.TrimSuffix(defaultGroup, "_processors")
		
		for i := range config.Processors {
			if config.Processors[i].Group == "" {
				config.Processors[i].Group = defaultGroup
			}
			// 榛樿鍚敤
			if !config.Processors[i].Enabled {
				// YAML涓湭璁剧疆鏃讹紝bool榛樿涓篺alse锛岃繖閲岄渶瑕佺壒娈婂鐞?
				// 閫氳繃妫€鏌ユ槸鍚︽湁enabled瀛楁鏉ュ垽鏂紙鐩墠绠€鍖栦负榛樿鍚敤锛?
			}
		}
	}

	return config.Processors, nil
}

// Reload 閲嶆柊鍔犺浇閰嶇疆
func (m *Manager) Reload() error {
	if err := m.Load(); err != nil {
		return err
	}

	// 瑙﹀彂鍥炶皟
	m.mu.RLock()
	callbacks := m.onReload
	m.mu.RUnlock()

	for _, cb := range callbacks {
		cb()
	}

	return nil
}

// OnReload 娉ㄥ唽閰嶇疆閲嶈浇鍥炶皟
func (m *Manager) OnReload(callback func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onReload = append(m.onReload, callback)
}

// WatchChanges 鐩戝惉閰嶇疆鏂囦欢鍙樺寲
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
					// 寤惰繜涓€涓嬶紝纭繚鏂囦欢鍐欏叆瀹屾垚
					time.Sleep(100 * time.Millisecond)
					if err := m.Reload(); err != nil {
						fmt.Printf("閰嶇疆閲嶈浇澶辫触: %v\n", err)
					} else {
						fmt.Println("閰嶇疆宸查噸鏂板姞杞?)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("閰嶇疆鐩戝惉閿欒: %v\n", err)
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

// Get 鑾峰彇涓婚厤缃紙鍙锛?
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetProcessors 鑾峰彇澶勭悊鍣ㄥ垪琛?
func (m *Manager) GetProcessors() []model.Processor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.processors == nil {
		return nil
	}
	return m.processors.Processors
}

// GetProcessor 鏍规嵁ID鑾峰彇澶勭悊鍣?
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

// expandEnvVars 灞曞紑鐜鍙橀噺
// 鏀寔 ${VAR} 鍜?$VAR 鏍煎紡
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return "${" + key + "}"
	})
}

// setDefaults 璁剧疆榛樿鍊?
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

// Validate 楠岃瘉閰嶇疆
func (c *Config) Validate() error {
	var errs []string

	if c.LLM.BaseURL == "" {
		errs = append(errs, "llm.base_url 涓嶈兘涓虹┖")
	}
	if c.LLM.APIKey == "" || strings.HasPrefix(c.LLM.APIKey, "${") {
		errs = append(errs, "llm.api_key 鏈缃垨鐜鍙橀噺鏈畾涔?)
	}
	if c.LLM.Model == "" {
		errs = append(errs, "llm.model 涓嶈兘涓虹┖")
	}

	if len(c.Kafka.Brokers) == 0 {
		errs = append(errs, "kafka.brokers 涓嶈兘涓虹┖")
	}

	if len(errs) > 0 {
		return fmt.Errorf("閰嶇疆楠岃瘉澶辫触:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
