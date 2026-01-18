package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yoyo3287258/home-gateway/internal/config"
)

// Server HTTP API鏈嶅姟鍣?
type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
	handler    *Handler
	cfg        *config.Config
	startTime  time.Time
}

// NewServer 鍒涘缓HTTP鏈嶅姟鍣?
func NewServer(handler *Handler, cfg *config.Config) *Server {
	// 璁剧疆Gin妯″紡
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	
	// 鍩虹涓棿浠?
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware())
	engine.Use(CORSMiddleware())

	// 瀹夊叏涓棿浠?
	engine.Use(IPWhitelistMiddleware(&cfg.Security))
	engine.Use(RateLimitMiddleware(&cfg.Security))

	s := &Server{
		engine:    engine,
		handler:   handler,
		cfg:       cfg,
		startTime: time.Now(),
	}

	// 娉ㄥ唽璺敱
	s.setupRoutes()

	return s
}

// setupRoutes 璁剧疆璺敱
func (s *Server) setupRoutes() {
	cfg := s.cfg

	// API v1 璺敱缁?
	v1 := s.engine.Group("/api/v1")
	{
		// 鍋ュ悍妫€鏌ワ紙涓嶉渶瑕佽璇侊級
		v1.GET("/health", s.handler.Health)

		// 闇€瑕丄PI Token璁よ瘉鐨勬帴鍙?
		protected := v1.Group("")
		protected.Use(APITokenAuthMiddleware(&cfg.Security))
		{
			// 鑾峰彇澶勭悊鍣ㄥ垪琛?
			protected.GET("/processors", s.handler.ListProcessors)

			// 閫氱敤鍛戒护鎺ュ彛
			protected.POST("/command", s.handler.Command)

			// 閰嶇疆閲嶈浇
			protected.POST("/config/reload", s.handler.ReloadConfig)
		}

		// Webhook鎺ュ彛锛堜娇鐢ㄥ悇鑷笭閬撶殑楠岃瘉鏈哄埗锛屼笉闇€瑕丄PI Token锛?
		webhook := v1.Group("/webhook")
		{
			// Telegram Webhook锛堥€氳繃 webhook_secret 楠岃瘉锛?
			webhook.POST("/telegram", s.handler.TelegramWebhook)
			
			// 浼佷笟寰俊Webhook锛堥鐣欙級
			webhook.POST("/wechat-work", s.handler.WeChatWorkWebhook)
		}
	}

	// 鏍硅矾寰勯噸瀹氬悜鍒板仴搴锋鏌?
	s.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/api/v1/health")
	})
}

// Start 鍚姩鏈嶅姟鍣?
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	fmt.Printf("馃殌 鏈嶅姟鍣ㄥ惎鍔ㄥ湪 http://%s\n", addr)
	fmt.Printf("馃摎 API鏂囨。: http://%s/api/v1/health\n", addr)

	return s.httpServer.ListenAndServe()
}

// Stop 鍋滄鏈嶅姟鍣?
func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

// GetStartTime 鑾峰彇鍚姩鏃堕棿
func (s *Server) GetStartTime() time.Time {
	return s.startTime
}

// Engine 鑾峰彇Gin寮曟搸锛堢敤浜庢祴璇曪級
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
