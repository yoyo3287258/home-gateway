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

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码设置颜色
		var statusColor string
		switch {
		case statusCode >= 500:
			statusColor = "\033[31m" // 红色
		case statusCode >= 400:
			statusColor = "\033[33m" // 黄色
		case statusCode >= 300:
			statusColor = "\033[36m" // 青色
		default:
			statusColor = "\033[32m" // 绿色
		}
		resetColor := "\033[0m"

		fmt.Printf("[GW] %s |%s %3d %s| %13v | %15s | %-7s %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusColor,
			statusCode,
			resetColor,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// CORSMiddleware CORS跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// TraceIDMiddleware TraceID中间件
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取TraceID，如果没有则生成
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}

		// 设置到Context和响应头
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// generateTraceID 生成TraceID
func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// APITokenAuthMiddleware API Token认证中间件
// 验证请求头中的 Authorization: Bearer <token>
func APITokenAuthMiddleware(securityCfg *config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果没有配置token，跳过验证（开发模式）
		if securityCfg.APIToken == "" {
			c.Next()
			return
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization格式错误，应为: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token != securityCfg.APIToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的API Token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(securityCfg *config.SecurityConfig) gin.HandlerFunc {
	// 预解析CIDR
	var networks []*net.IPNet
	var singleIPs []net.IP
	
	for _, cidr := range securityCfg.IPWhitelist {
		if strings.Contains(cidr, "/") {
			_, network, err := net.ParseCIDR(cidr)
			if err == nil {
				networks = append(networks, network)
			}
		} else {
			ip := net.ParseIP(cidr)
			if ip != nil {
				singleIPs = append(singleIPs, ip)
			}
		}
	}

	return func(c *gin.Context) {
		// 如果没有配置白名单，跳过验证
		if len(securityCfg.IPWhitelist) == 0 {
			c.Next()
			return
		}

		clientIP := net.ParseIP(c.ClientIP())
		if clientIP == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "无法解析客户端IP",
			})
			c.Abort()
			return
		}

		// 检查是否在白名单中
		allowed := false
		
		// 检查单个IP
		for _, ip := range singleIPs {
			if ip.Equal(clientIP) {
				allowed = true
				break
			}
		}

		// 检查CIDR
		if !allowed {
			for _, network := range networks {
				if network.Contains(clientIP) {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": fmt.Sprintf("IP %s 不在白名单中", c.ClientIP()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter 简单的请求限速器
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// NewRateLimiter 创建限速器
func NewRateLimiter(limitPerMinute int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limitPerMinute,
		window:   time.Minute,
	}
}

// Allow 检查是否允许请求
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// 清理过期请求
	if times, ok := r.requests[key]; ok {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		r.requests[key] = valid
	}

	// 检查是否超过限制
	if len(r.requests[key]) >= r.limit {
		return false
	}

	// 记录请求
	r.requests[key] = append(r.requests[key], now)
	return true
}

// RateLimitMiddleware 请求限速中间件
func RateLimitMiddleware(securityCfg *config.SecurityConfig) gin.HandlerFunc {
	if securityCfg.RateLimitPerMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	limiter := NewRateLimiter(securityCfg.RateLimitPerMinute)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": fmt.Sprintf("请求过于频繁，限制为每分钟%d次", securityCfg.RateLimitPerMinute),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
