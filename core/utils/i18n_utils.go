package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log"
	"strings"
	"sync"
)

const i18nPayloadPrefix = "__i18n__:"

type i18nPayload struct {
	Key  string                 `json:"key"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// I18n 结构体
type I18n struct {
	bundle *i18n.Bundle
}

var (
	I18nUtil *I18n
	once     sync.Once
)

func InitI18n() {
	once.Do(func() {
		bundle := i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
		// 加载本地化文件
		bundle.MustLoadMessageFile("core/locales/en.json")
		bundle.MustLoadMessageFile("core/locales/es-MX.json")
		bundle.MustLoadMessageFile("core/locales/pt-BR.json")
		bundle.MustLoadMessageFile("core/locales/id.json")
		I18nUtil = &I18n{bundle: bundle}
		log.Printf("i18n Succeed")
	})
}

// Translate 翻译函数
func (i *I18n) Translate(c *gin.Context, key string, data map[string]interface{}) string {
	if i == nil || i.bundle == nil {
		return key
	}
	lang := resolveI18nLanguage(c)

	localizer := i18n.NewLocalizer(i.bundle, lang)
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return message
}

func resolveI18nLanguage(c *gin.Context) string {
	if c == nil {
		return "en"
	}

	raw := strings.TrimSpace(strings.ToLower(c.GetHeader("Accept-Language")))
	if raw == "" {
		return "en"
	}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if idx := strings.Index(tag, ";"); idx >= 0 {
			tag = strings.TrimSpace(tag[:idx])
		}

		switch {
		case tag == "en" || strings.HasPrefix(tag, "en-"):
			return "en"
		case tag == "es" || tag == "es-mx" || tag == "mx" || strings.HasPrefix(tag, "es-"):
			return "es-MX"
		case tag == "pt" || tag == "pt-br" || tag == "br" || strings.HasPrefix(tag, "pt-"):
			return "pt-BR"
		case tag == "id" || tag == "id-id" || strings.HasPrefix(tag, "id-"):
			return "id"
		}
	}

	return "en"
}

func I18nMessage(key string, data map[string]interface{}) string {
	payload, err := json.Marshal(i18nPayload{
		Key:  key,
		Data: data,
	})
	if err != nil {
		return key
	}
	return i18nPayloadPrefix + string(payload)
}

func TranslateMessage(c *gin.Context, msg string) string {
	if msg == "" {
		return msg
	}
	if strings.HasPrefix(msg, i18nPayloadPrefix) {
		var payload i18nPayload
		if err := json.Unmarshal([]byte(strings.TrimPrefix(msg, i18nPayloadPrefix)), &payload); err == nil && payload.Key != "" {
			return I18nUtil.Translate(c, payload.Key, payload.Data)
		}
	}
	return I18nUtil.Translate(c, msg, nil)
}
