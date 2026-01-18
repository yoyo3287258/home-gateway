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

// 鐗堟湰淇℃伅锛堝湪缂栬瘧鏃堕€氳繃 -ldflags 娉ㄥ叆锛?
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 鍛戒护琛屽弬鏁?
	var (
		configPath     string
		processorsPath string
		showVersion    bool
		selfUpdate     bool
	)

	flag.StringVar(&configPath, "config", "configs/config.yaml", "涓婚厤缃枃浠惰矾寰?)
	flag.StringVar(&configPath, "c", "configs/config.yaml", "涓婚厤缃枃浠惰矾寰?(绠€鍐?")
	flag.StringVar(&processorsPath, "processors", "configs/processors", "澶勭悊鍣ㄩ厤缃洰褰曟垨鏂囦欢璺緞")
	flag.StringVar(&processorsPath, "p", "configs/processors", "澶勭悊鍣ㄩ厤缃洰褰曟垨鏂囦欢璺緞 (绠€鍐?")
	flag.BoolVar(&showVersion, "version", false, "鏄剧ず鐗堟湰淇℃伅")
	flag.BoolVar(&showVersion, "v", false, "鏄剧ず鐗堟湰淇℃伅 (绠€鍐?")
	flag.BoolVar(&selfUpdate, "update", false, "妫€鏌ュ苟鏇存柊鍒版渶鏂扮増鏈?)
	flag.BoolVar(&selfUpdate, "U", false, "妫€鏌ュ苟鏇存柊鍒版渶鏂扮増鏈?(绠€鍐?")
	flag.Parse()

	// 鏄剧ず鐗堟湰
	if showVersion {
		fmt.Printf("Home Gateway %s\n", Version)
		fmt.Printf("鏋勫缓鏃堕棿: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	// 鑷洿鏂?
	if selfUpdate {
		if err := doSelfUpdate(); err != nil {
			fmt.Printf("鉂?鏇存柊澶辫触: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 纭繚閰嶇疆鏂囦欢璺緞鏄粷瀵硅矾寰?
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

	// 鎵撳嵃鍚姩淇℃伅
	fmt.Println("馃彔 Home Gateway 鍚姩涓?..")
	fmt.Printf("   鐗堟湰: %s\n", Version)
	fmt.Printf("   閰嶇疆: %s\n", configPath)
	fmt.Printf("   澶勭悊鍣? %s\n", processorsPath)

	// 鍔犺浇閰嶇疆
	configMgr := config.NewManager(configPath, processorsPath)
	if err := configMgr.Load(); err != nil {
		fmt.Printf("鉂?鍔犺浇閰嶇疆澶辫触: %v\n", err)
		os.Exit(1)
	}

	cfg := configMgr.Get()

	// 楠岃瘉閰嶇疆
	if err := cfg.Validate(); err != nil {
		fmt.Printf("鉂?閰嶇疆楠岃瘉澶辫触: %v\n", err)
		os.Exit(1)
	}

	// 鍒涘缓LLM瀹㈡埛绔?
	llmClient := llm.NewClient(&cfg.LLM)
	fmt.Printf("   LLM: %s (%s)\n", cfg.LLM.BaseURL, cfg.LLM.Model)

	// 鍒涘缓Kafka瀹㈡埛绔紙鍙€夛級
	var kafkaClient *kafka.Client
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		var err error
		kafkaClient, err = kafka.NewClient(&cfg.Kafka)
		if err != nil {
			fmt.Printf("鈿狅笍  Kafka杩炴帴澶辫触锛堝皢浠ユ棤Kafka妯″紡杩愯锛? %v\n", err)
		} else {
			fmt.Printf("   Kafka: %v\n", cfg.Kafka.Brokers)
		}
	} else {
		fmt.Println("   Kafka: 鏈厤缃紙浠ユ祴璇曟ā寮忚繍琛岋級")
	}

	// 鍚姩閰嶇疆鏂囦欢鐩戝惉
	if err := configMgr.WatchChanges(); err != nil {
		fmt.Printf("鈿狅笍  閰嶇疆鏂囦欢鐩戝惉鍚姩澶辫触: %v\n", err)
	}

	// 鍒涘缓澶勭悊鍣ㄥ拰鏈嶅姟鍣?
	handler := api.NewHandler(configMgr, llmClient, kafkaClient)
	server := api.NewServer(handler, cfg)

	// 澶勭悊鍣ㄦ暟閲?
	processors := configMgr.GetProcessors()
	enabledCount := 0
	for _, p := range processors {
		if p.Enabled {
			enabledCount++
		}
	}
	fmt.Printf("   澶勭悊鍣? %d 涓凡鍔犺浇\n", enabledCount)

	// 瀹夊叏閰嶇疆鐘舵€?
	if cfg.Security.APIToken != "" {
		fmt.Println("   馃攼 API Token璁よ瘉: 宸插惎鐢?)
	} else {
		fmt.Println("   鈿狅笍  API Token璁よ瘉: 鏈厤缃紙涓嶅畨鍏級")
	}
	if len(cfg.Security.IPWhitelist) > 0 {
		fmt.Printf("   馃攼 IP鐧藉悕鍗? %d 鏉¤鍒橽n", len(cfg.Security.IPWhitelist))
	}

	// 浼橀泤鍏抽棴
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		fmt.Println("\n馃洃 姝ｅ湪鍏抽棴鏈嶅姟...")
		if err := server.Stop(); err != nil {
			fmt.Printf("鍏抽棴鏈嶅姟鍣ㄥけ璐? %v\n", err)
		}
		if kafkaClient != nil {
			if err := kafkaClient.Close(); err != nil {
				fmt.Printf("鍏抽棴Kafka澶辫触: %v\n", err)
			}
		}
		os.Exit(0)
	}()

	// 鍚姩鏈嶅姟鍣?
	if err := server.Start(); err != nil {
		fmt.Printf("鉂?鏈嶅姟鍣ㄥ惎鍔ㄥけ璐? %v\n", err)
		os.Exit(1)
	}
}

// doSelfUpdate 鎵ц鑷洿鏂?
func doSelfUpdate() error {
	fmt.Println("馃攧 妫€鏌ユ洿鏂?..")
	
	updater := NewUpdater("home-gateway", "home-gateway")
	
	// 鑾峰彇鏈€鏂扮増鏈?
	latestVersion, downloadURL, err := updater.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("鑾峰彇鏈€鏂扮増鏈け璐? %w", err)
	}

	if latestVersion == Version {
		fmt.Printf("鉁?褰撳墠宸叉槸鏈€鏂扮増鏈?(%s)\n", Version)
		return nil
	}

	fmt.Printf("馃摝 鍙戠幇鏂扮増鏈? %s -> %s\n", Version, latestVersion)
	fmt.Printf("馃敆 涓嬭浇鍦板潃: %s\n", downloadURL)

	// 涓嬭浇骞舵浛鎹?
	if err := updater.DownloadAndReplace(downloadURL); err != nil {
		return fmt.Errorf("涓嬭浇鏇存柊澶辫触: %w", err)
	}

	fmt.Printf("鉁?鏇存柊瀹屾垚锛佽閲嶆柊鍚姩绋嬪簭銆俓n")
	return nil
}
