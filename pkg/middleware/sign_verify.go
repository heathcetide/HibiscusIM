package middleware

import (
	"HibiscusIM/pkg/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 生成 HMAC 签名
func generateSignature(data, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// API 签名验证中间件
func SignVerifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取签名
		signature := c.GetHeader("Signature")
		if signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Signature is missing"})
			c.Abort()
			return
		}

		// 获取请求的时间戳和请求体（例如：GET /api/resource?timestamp=xxx）
		timestamp := c.DefaultQuery("timestamp", "")
		if timestamp == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Timestamp is missing"})
			c.Abort()
			return
		}

		// 获取请求体，如果是 POST 请求可以读取其 Body 内容
		var requestBody string
		if c.Request.Method == http.MethodPost {
			bodyBytes, _ := c.GetRawData()
			requestBody = string(bodyBytes)
		}

		// 拼接用于签名的数据：请求路径 + 请求体 + 时间戳
		data := fmt.Sprintf("%s%s%s", c.Request.Method, c.Request.URL.Path, requestBody+timestamp)

		// 使用生成的签名与请求头中的签名进行比较
		expectedSignature := generateSignature(data, config.GlobalConfig.APISecretKey)
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			c.Abort()
			return
		}

		// 签名验证通过，继续处理请求
		c.Next()
	}
}
