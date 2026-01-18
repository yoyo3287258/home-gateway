package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yoyo3287258/home-gateway/internal/config"
)

// Server HTTP API服务器
type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
	handler    *Handler
	cfg        *config.Config
	startTime  time.Time
}

// NewServer 创建HTTP服务器
func NewServer(handler *Handler, cfg *config.Config) *Server {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	
	// 基础中间件
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware())
	engine.Use(CORSMiddleware())

	// 安全中间件
	engine.Use(IPWhitelistMiddleware(&cfg.Security))
	engine.Use(RateLimitMiddleware(&cfg.Security))

	s := &Server{
		engine:    engine,
		handler:   handler,
		cfg:       cfg,
		startTime: time.Now(),
	}

	// 注册路由
	s.setupRoutes()

	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	cfg := s.cfg

	// API v1 路由组
	v1 := s.engine.Group("/api/v1")
	{
		// 健康检查（不需要认证）
		v1.GET("/health", s.handler.Health)

		// 需要API Token认证的接口
		protected := v1.Group("")
		protected.Use(APITokenAuthMiddleware(&cfg.Security))
		{
			// 获取处理器列表
			protected.GET("/processors", s.handler.ListProcessors)

			// 通用命令接口
			protected.POST("/command", s.handler.Command)

			// 配置重载
			protected.POST("/config/reload", s.handler.ReloadConfig)
		}

		// Webhook接口（使用各自渠道的验证机制，不需要API Token）
		webhook := v1.Group("/webhook")
		{
			// Telegram Webhook（通过 webhook_secret 验证）
			webhook.POST("/telegram", s.handler.TelegramWebhook)
			
			// 企业微信Webhook（预留）
			webhook.POST("/wechat-work", s.handler.WeChatWorkWebhook)
		}
	}

	// 根路径重定向到健康检查
	s.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/api/v1/health")
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	fmt.Printf("🚀 服务器启动在 http://%s\n", addr)
	fmt.Printf("📚 API文档: http://%s/api/v1/health\n", addr)

	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

// GetStartTime 获取启动时间
func (s *Server) GetStartTime() time.Time {
	return s.startTime
}

// Engine 获取Gin引擎（用于测试）
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
