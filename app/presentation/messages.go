package presentation

import (
	"github.com/dawkaka/theone/inter"
	"github.com/gin-gonic/gin"
)

func Error(lang, message string) gin.H {
	translation := translate(lang, message)
	return gin.H{"type": "ERROR", "message": translation}
}

func translate(lang, message string) string {
	translation := inter.Localize(lang, message)
	return translation
}

func Success(lang, message string) gin.H {
	translation := translate(lang, message)
	return gin.H{"type": "SUCCESS", "message": translation}
}
