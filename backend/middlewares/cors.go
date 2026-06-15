package middlewares

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 放开本地前端开发端口，保证 Vite dev server 能直接联调后端 API。
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc:  allowLocalFrontendOrigin,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// allowLocalFrontendOrigin 允许配置中的固定域名，以及本地 1200-1299 前端开发端口。
func allowLocalFrontendOrigin(origin string) bool {
	for _, allowed := range strings.Split(os.Getenv("TWO_TO_CORS_ALLOW_ORIGINS"), ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1" {
		return false
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		return false
	}
	return port >= 1200 && port <= 1299
}
