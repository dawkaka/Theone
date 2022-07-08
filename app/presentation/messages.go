package presentation

import (
	"github.com/gin-gonic/gin"
)

func Error(message string) gin.H {
	return gin.H{"type": "ERROR", "message": message}
}

func Success(message string) gin.H {
	return gin.H{"type": "SUCCESS", "message": message}
}
