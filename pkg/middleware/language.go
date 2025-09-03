package middleware

import (
	"HibiscusIM/pkg/i18n"

	"github.com/gin-gonic/gin"
)

func LanguageMiddleware(i18nSupport *i18n.I18nSupport) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求中的语言（从头部或者查询参数）
		lang := c.DefaultQuery("lang", "en") // 默认是英语
		if lang != "en" && lang != "zh" {
			lang = "en" // 如果传入的语言无效，则使用默认的英文
		}

		// 设置语言
		c.Set("lang", lang)
		// 继续处理
		c.Next()
	}
}
