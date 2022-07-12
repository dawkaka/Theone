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

func Localize(lang, messageId string) string {
	bundle := Initialize()
	loc := i18n.NewLocalizer(bundle, lang)
	translation := loc.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageId,
	})

	return translation
}

func LocalizeWithFullName(lang, firstName, lastName, messageId string) string {

	bundle := Initialize()
	loc := i18n.NewLocalizer(bundle, lang)
	translation := loc.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageId,
		TemplateData: map[string]interface{}{
			"FirstName": firstName,
			"LastName":  lastName,
		},
	})

	return translation
}
