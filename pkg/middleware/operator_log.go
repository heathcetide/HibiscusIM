package middleware

import (
	constants "HibiscusIM/pkg/constant"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
	"gorm.io/gorm"
)

// OperationLogMiddleware 记录操作日志
func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet(constants.DbField).(*gorm.DB)
		// 获取用户 ID 和用户名（假设已通过认证中间件获取）
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		// 获取请求的操作和目标
		action := c.Request.Method   // 操作类型：POST、GET、PUT、DELETE
		target := c.Request.URL.Path // 操作目标：API 路径

		// 获取请求的 IP 地址
		ipAddress := c.ClientIP()

		// 获取用户代理信息
		userAgent := c.GetHeader("User-Agent")

		// 获取请求来源页面
		referer := c.GetHeader("Referer")

		ua := user_agent.New(c.GetHeader("User-Agent"))
		device := ua.Platform()
		browser, version := ua.Browser()
		os := ua.OS()

		// 获取请求方法
		requestMethod := c.Request.Method

		// 获取地理位置信息（根据 IP 获取）
		location := getGeoLocation(ipAddress)

		// 记录操作日志
		err := CreateOperationLog(db, userID.(int64), username.(string), action, target, "User action recorded", ipAddress, userAgent, referer, device, browser+version, os, location.(string), requestMethod)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record operation log"})
			c.Abort()
			return
		}

		// 继续执行后续处理
		c.Next()
	}
}

// OperationLog 记录用户操作日志
type OperationLog struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	UserID          int64     `gorm:"not null" json:"user_id"`          // 操作的用户 ID
	Username        string    `gorm:"not null" json:"username"`         // 操作的用户名
	Action          string    `gorm:"not null" json:"action"`           // 操作类型（如：创建、删除、更新等）
	Target          string    `gorm:"not null" json:"target"`           // 操作目标（如：用户、订单等）
	Details         string    `gorm:"not null" json:"details"`          // 操作详细描述
	IPAddress       string    `gorm:"not null" json:"ip_address"`       // 用户 IP 地址
	UserAgent       string    `gorm:"not null" json:"user_agent"`       // 用户的浏览器信息
	Referer         string    `gorm:"not null" json:"referer"`          // 请求来源页面
	Device          string    `gorm:"not null" json:"device"`           // 用户设备（手机、桌面等）
	Browser         string    `gorm:"not null" json:"browser"`          // 浏览器信息（如 Chrome, Firefox 等）
	OperatingSystem string    `gorm:"not null" json:"operating_system"` // 操作系统（如 Windows, MacOS 等）
	Location        string    `gorm:"not null" json:"location"`         // 用户的地理位置
	RequestMethod   string    `gorm:"not null" json:"request_method"`   // HTTP 请求方法（GET、POST等）
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"` // 操作时间
}

// CreateOperationLog 创建操作日志
func CreateOperationLog(db *gorm.DB, userID int64, username, action, target, details, ipAddress, userAgent, referer, device, browser, operatingSystem, location, requestMethod string) error {
	log := OperationLog{
		UserID:          userID,
		Username:        username,
		Action:          action,
		Target:          target,
		Details:         details,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Referer:         referer,
		Device:          device,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		Location:        location,
		RequestMethod:   requestMethod,
		CreatedAt:       time.Now(),
	}

	// 保存操作日志到数据库
	if err := db.Create(&log).Error; err != nil {
		return err
	}
	return nil
}

func getGeoLocation(address string) interface{} {
	// 使用 GeoIP 获取位置信息
	reader, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	record, err := reader.City(net.ParseIP(address))
	if err != nil {
		log.Fatal(err)
	}
	return record.City.Names["en"] // 返回城市名
}
