package util

import (
	"context"
	"errors"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

var bundle *i18n.Bundle

func RegisterLanguages(languages ...string) {
	if languages == nil {
		panic("at least one language pack is required")
	}

	bundle = i18n.NewBundle(language.English)
	ext := "yaml"
	bundle.RegisterUnmarshalFunc(ext, yaml.Unmarshal)
	for _, v := range languages {
		filePath := fmt.Sprintf("./etc/i18n/%s.%s", v, ext)
		bundle.MustLoadMessageFile(filePath)
	}

}

func Localize(lang string, msgId string, templateData ...interface{}) string {
	if lang == "" {
		lang = "en"
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	var msg string
	var err error

	if templateData == nil {
		msg, err = localizer.Localize(&i18n.LocalizeConfig{
			MessageID: msgId,
		})
	} else {
		msg, err = localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    msgId,
			TemplateData: templateData[0],
		})
	}

	if err != nil {
		fmt.Println("unspecified", templateData[0], err)
		return "unspecified error occured"
	}

	return msg
}

// A Lang represents a lang.
type Lang interface {
	Localize(msgId string, templateData ...interface{}) string
	Error(removedPrefixMsgId string, templateData ...interface{}) error
}

type lang struct {
	ctx context.Context
	// lang string
}

// WithContext sets ctx to lang, for keeping tracing information.
func WithContext(ctx context.Context) Lang {
	return &lang{
		ctx: ctx,
	}
}

func (l *lang) Localize(msgId string, templateData ...interface{}) string {
	md, ok := metadata.FromIncomingContext(l.ctx)
	if !ok {
		return Localize("en", "error.missingBasicParameter", map[string]string{"param": "lang"})
	}

	// 获取lang
	lang := ""
	envArr := md.Get("lang")
	if len(envArr) >= 1 {
		lang = envArr[0]
	}

	return Localize(lang, msgId, templateData...)
}

func (l *lang) Error(removedPrefixMsgId string, templateData ...interface{}) error {
	prefix := "error."
	msg := l.Localize(prefix+removedPrefixMsgId, templateData...)
	return errors.New(msg)
}
