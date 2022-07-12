package utils

import (
	"net/http"
)

func GetLang(header http.Header) string {
	langArr := header["Accept-Language"]
	var lang string
	if len(langArr) > 0 {
		lang = langArr[0]
	} else {
		lang = "en"
	}
	return lang
}
