package middlewares

import (
	"net/http"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user")
		if user == nil {
			ctx.JSON(http.StatusUnauthorized, presentation.Error("Login required"))
			ctx.Abort()
		} else {
			ctx.Next()
		}
	}
}
