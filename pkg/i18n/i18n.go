package i18n

import (
	"encoding/json"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// I18nSupport 国际化支持结构体
type I18nSupport struct {
	bundle *i18n.Bundle
}

// NewI18nSupport 初始化国际化支持
func NewI18nSupport(defaultLang string) (*I18nSupport, error) {
	// 创建翻译器实例
	bundle := i18n.NewBundle(language.MustParse(defaultLang))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 加载语言文件（中文和英文）
	_, err := bundle.LoadMessageFile("locales/zh.json")
	if err != nil {
		log.Printf("failed to load zh.json: %v", err)
		// 不返回错误，因为可能只需要英文
	}

	_, err = bundle.LoadMessageFile("locales/en.json")
	if err != nil {
		log.Printf("failed to load en.json: %v", err)
		// 不返回错误，因为可能只需要中文
	}

	return &I18nSupport{
		bundle: bundle,
	}, nil
}

// T 获取翻译文本
func (i *I18nSupport) T(languageTag, key string, templateData map[string]interface{}) string {
	localizer := i18n.NewLocalizer(i.bundle, languageTag)

	translation, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})

	if err != nil {
		log.Printf("Error translating key %s: %v", key, err)
		return key // 返回键名作为默认值
	}

	return translation
}

// TWithDefaultLang 使用默认语言获取翻译文本
func (i *I18nSupport) TWithDefaultLang(key string, templateData map[string]interface{}) string {
	// 使用bundle的默认语言
	localizer := i18n.NewLocalizer(i.bundle)

	translation, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})

	if err != nil {
		log.Printf("Error translating key %s: %v", key, err)
		return key // 返回键名作为默认值
	}

	return translation
}
