package middlewares

import (
	"encoding/gob"
	"net/http"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gob.Register(entity.UserSession{})
		user := sessions.Default(ctx).Get("user")
		if user == nil || user == "" {
			ctx.JSON(http.StatusUnauthorized, presentation.Error(utils.GetLang("", ctx.Request.Header), "LoginRequired"))
			ctx.Abort()
		} else {
			ctx.Next()
		}
	}
}
