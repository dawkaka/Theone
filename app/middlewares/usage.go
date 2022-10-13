package middlewares

import (
	"fmt"
	"time"

	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func UsageMonitoring(service user.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c).Get("user")
		if session != nil {
			user := session.(entity.UserSession)
			fmt.Println(time.Since(user.LastVisited).Hours())
			if time.Since(user.LastVisited).Hours() > 24 {
				err := service.UsageMonitoring(user.ID)
				fmt.Println(err)
			}
		}
		c.Next()
	}
}
