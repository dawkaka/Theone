package middlewares

import (
	"net/http"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CheckBlocked(service couple.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := utils.GetSession(sessions.Default(c))
		coupleName := c.Param("coupleName")
		blocked, err := service.IsBlocked(coupleName, user.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "BadRequest"))
			c.Abort()
		}
		if blocked {
			c.JSON(http.StatusForbidden, presentation.Error(user.Lang, "Forbidden"))
			c.Abort()
		} else {
			c.Next()
		}
	}
}
