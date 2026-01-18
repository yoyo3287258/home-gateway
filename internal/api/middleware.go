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

// LoggerMiddleware 鏃ュ織涓棿浠?
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 澶勭悊璇锋眰
		c.Next()

		// 璁板綍鏃ュ織
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// 鏍规嵁鐘舵€佺爜璁剧疆棰滆壊
		var statusColor string
		switch {
		case statusCode >= 500:
			statusColor = "\033[31m" // 绾㈣壊
		case statusCode >= 400:
			statusColor = "\033[33m" // 榛勮壊
		case statusCode >= 300:
			statusColor = "\033[36m" // 闈掕壊
		default:
			statusColor = "\033[32m" // 缁胯壊
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

// CORSMiddleware CORS璺ㄥ煙涓棿浠?
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

// TraceIDMiddleware TraceID涓棿浠?
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 浠庤姹傚ご鑾峰彇TraceID锛屽鏋滄病鏈夊垯鐢熸垚
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}

		// 璁剧疆鍒癈ontext鍜屽搷搴斿ご
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// generateTraceID 鐢熸垚TraceID
func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// APITokenAuthMiddleware API Token璁よ瘉涓棿浠?
// 楠岃瘉璇锋眰澶翠腑鐨?Authorization: Bearer <token>
func APITokenAuthMiddleware(securityCfg *config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 濡傛灉娌℃湁閰嶇疆token锛岃烦杩囬獙璇侊紙寮€鍙戞ā寮忥級
		if securityCfg.APIToken == "" {
			c.Next()
			return
		}

		// 鑾峰彇Authorization澶?
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缂哄皯Authorization澶?,
			})
			c.Abort()
			return
		}

		// 瑙ｆ瀽Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization鏍煎紡閿欒锛屽簲涓? Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token != securityCfg.APIToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "鏃犳晥鐨凙PI Token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPWhitelistMiddleware IP鐧藉悕鍗曚腑闂翠欢
func IPWhitelistMiddleware(securityCfg *config.SecurityConfig) gin.HandlerFunc {
	// 棰勮В鏋怌IDR
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
		// 濡傛灉娌℃湁閰嶇疆鐧藉悕鍗曪紝璺宠繃楠岃瘉
		if len(securityCfg.IPWhitelist) == 0 {
			c.Next()
			return
		}

		clientIP := net.ParseIP(c.ClientIP())
		if clientIP == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "鏃犳硶瑙ｆ瀽瀹㈡埛绔疘P",
			})
			c.Abort()
			return
		}

		// 妫€鏌ユ槸鍚﹀湪鐧藉悕鍗曚腑
		allowed := false
		
		// 妫€鏌ュ崟涓狪P
		for _, ip := range singleIPs {
			if ip.Equal(clientIP) {
				allowed = true
				break
			}
		}

		// 妫€鏌IDR
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
				"message": fmt.Sprintf("IP %s 涓嶅湪鐧藉悕鍗曚腑", c.ClientIP()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter 绠€鍗曠殑璇锋眰闄愰€熷櫒
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// NewRateLimiter 鍒涘缓闄愰€熷櫒
func NewRateLimiter(limitPerMinute int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limitPerMinute,
		window:   time.Minute,
	}
}

// Allow 妫€鏌ユ槸鍚﹀厑璁歌姹?
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// 娓呯悊杩囨湡璇锋眰
	if times, ok := r.requests[key]; ok {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		r.requests[key] = valid
	}

	// 妫€鏌ユ槸鍚﹁秴杩囬檺鍒?
	if len(r.requests[key]) >= r.limit {
		return false
	}

	// 璁板綍璇锋眰
	r.requests[key] = append(r.requests[key], now)
	return true
}

// RateLimitMiddleware 璇锋眰闄愰€熶腑闂翠欢
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
				"message": fmt.Sprintf("璇锋眰杩囦簬棰戠箒锛岄檺鍒朵负姣忓垎閽?d娆?, securityCfg.RateLimitPerMinute),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
