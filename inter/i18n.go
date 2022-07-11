package inter

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func Initialize() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("../config/locales/en.json")
	bundle.MustLoadMessageFile("../config/locales/es.json")
	return bundle
}

func Localizer(lang, messageId string) string {
	bundle := Initialize()
	loc := i18n.NewLocalizer(bundle, lang)
	messagesCount := 10
	translation := loc.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageId,
		TemplateData: map[string]interface{}{
			"Name":  "Alex",
			"Count": messagesCount,
		},
		PluralCount: messagesCount,
	})

	return translation
}
