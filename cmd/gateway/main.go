package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/yoyo3287258/home-gateway/internal/api"
	"github.com/yoyo3287258/home-gateway/internal/config"
	"github.com/yoyo3287258/home-gateway/internal/kafka"
	"github.com/yoyo3287258/home-gateway/internal/llm"
)

// 版本信息（在编译时通过 -ldflags 注入）
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 命令行参数
	var (
		configPath     string
		processorsPath string
		showVersion    bool
		selfUpdate     bool
	)

	flag.StringVar(&configPath, "config", "configs/config.yaml", "主配置文件路径")
	flag.StringVar(&configPath, "c", "configs/config.yaml", "主配置文件路径 (简写)")
	flag.StringVar(&processorsPath, "processors", "configs/processors", "处理器配置目录或文件路径")
	flag.StringVar(&processorsPath, "p", "configs/processors", "处理器配置目录或文件路径 (简写)")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息 (简写)")
	flag.BoolVar(&selfUpdate, "update", false, "检查并更新到最新版本")
	flag.BoolVar(&selfUpdate, "U", false, "检查并更新到最新版本 (简写)")
	flag.Parse()

	// 显示版本
	if showVersion {
		fmt.Printf("Home Gateway %s\n", Version)
		fmt.Printf("构建时间: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	// 自更新
	if selfUpdate {
		if err := doSelfUpdate(); err != nil {
			fmt.Printf("❌ 更新失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 确保配置文件路径是绝对路径
	if !filepath.IsAbs(configPath) {
		execDir, _ := os.Executable()
		execDir = filepath.Dir(execDir)
		configPath = filepath.Join(execDir, configPath)
	}
	if !filepath.IsAbs(processorsPath) {
		execDir, _ := os.Executable()
		execDir = filepath.Dir(execDir)
		processorsPath = filepath.Join(execDir, processorsPath)
	}

	// 打印启动信息
	fmt.Println("🏠 Home Gateway 启动中...")
	fmt.Printf("   版本: %s\n", Version)
	fmt.Printf("   配置: %s\n", configPath)
	fmt.Printf("   处理器: %s\n", processorsPath)

	// 加载配置
	configMgr := config.NewManager(configPath, processorsPath)
	if err := configMgr.Load(); err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	cfg := configMgr.Get()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ 配置验证失败: %v\n", err)
		os.Exit(1)
	}

	// 创建LLM客户端
	llmClient := llm.NewClient(&cfg.LLM)
	fmt.Printf("   LLM: %s (%s)\n", cfg.LLM.BaseURL, cfg.LLM.Model)

	// 创建Kafka客户端（可选）
	var kafkaClient *kafka.Client
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		var err error
		kafkaClient, err = kafka.NewClient(&cfg.Kafka)
		if err != nil {
			fmt.Printf("⚠️  Kafka连接失败（将以无Kafka模式运行）: %v\n", err)
		} else {
			fmt.Printf("   Kafka: %v\n", cfg.Kafka.Brokers)
		}
	} else {
		fmt.Println("   Kafka: 未配置（以测试模式运行）")
	}

	// 启动配置文件监听
	if err := configMgr.WatchChanges(); err != nil {
		fmt.Printf("⚠️  配置文件监听启动失败: %v\n", err)
	}

	// 创建处理器和服务器
	handler := api.NewHandler(configMgr, llmClient, kafkaClient)
	server := api.NewServer(handler, cfg)

	// 处理器数量
	processors := configMgr.GetProcessors()
	enabledCount := 0
	for _, p := range processors {
		if p.Enabled {
			enabledCount++
		}
	}
	fmt.Printf("   处理器: %d 个已加载\n", enabledCount)

	// 安全配置状态
	if cfg.Security.APIToken != "" {
		fmt.Println("   🔐 API Token认证: 已启用")
	} else {
		fmt.Println("   ⚠️  API Token认证: 未配置（不安全）")
	}
	if len(cfg.Security.IPWhitelist) > 0 {
		fmt.Printf("   🔐 IP白名单: %d 条规则\n", len(cfg.Security.IPWhitelist))
	}

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		fmt.Println("\n🛑 正在关闭服务...")
		if err := server.Stop(); err != nil {
			fmt.Printf("关闭服务器失败: %v\n", err)
		}
		if kafkaClient != nil {
			if err := kafkaClient.Close(); err != nil {
				fmt.Printf("关闭Kafka失败: %v\n", err)
			}
		}
		os.Exit(0)
	}()

	// 启动服务器
	if err := server.Start(); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

// doSelfUpdate 执行自更新
func doSelfUpdate() error {
	fmt.Println("🔄 检查更新...")
	
	updater := NewUpdater("yoyo3287258", "home-gateway")
	
	// 获取最新版本
	latestVersion, downloadURL, err := updater.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("获取最新版本失败: %w", err)
	}

	if latestVersion == Version {
		fmt.Printf("✅ 当前已是最新版本 (%s)\n", Version)
		return nil
	}

	fmt.Printf("📦 发现新版本: %s -> %s\n", Version, latestVersion)
	fmt.Printf("🔗 下载地址: %s\n", downloadURL)

	// 下载并替换
	if err := updater.DownloadAndReplace(downloadURL); err != nil {
		return fmt.Errorf("下载更新失败: %w", err)
	}

	fmt.Printf("✅ 更新完成！请重新启动程序。\n")
	return nil
}
