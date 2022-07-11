package presentation

import (
	"net/http"

	"github.com/dawkaka/theone/inter"
	"github.com/gin-gonic/gin"
)

func Error(header http.Header, message string) gin.H {
	translation := translate(header, message)
	return gin.H{"type": "ERROR", "message": translation}
}

func translate(header http.Header, message string) string {
	langArr := header["Accept-Language"]
	var lang string
	if len(langArr) > 0 {
		lang = langArr[0]
	} else {
		lang = "en"
	}
	translation := inter.Localizer(lang, message)
	return translation
}

func Success(header http.Header, message string) gin.H {
	translation := translate(header, message)
	return gin.H{"type": "SUCCESS", "message": translation}
}
