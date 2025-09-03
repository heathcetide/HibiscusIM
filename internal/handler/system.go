package handlers

import (
	"HibiscusIM/pkg/middleware"
	"HibiscusIM/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateRateLimiterConfig 更新限流配置
func (h *Handlers) UpdateRateLimiterConfig(c *gin.Context) {
	var config middleware.RateLimiterConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Fail(c, "invalid request", nil)
		return
	}

	// 更新限流配置
	middleware.SetRateLimiterConfig(config)
	response.Success(c, "rate limiter config updated", nil)
}

// HealthCheck 健康检查接口
func (h *Handlers) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection failed"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
		return
	}

	// 返回健康状态
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
