package middlewares

import (
	"time"

	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func UsageMonitoring(service user.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userSession := session.Get("user")
		if userSession != nil {
			user := userSession.(entity.UserSession)
			if time.Since(user.LastVisited).Hours() > 24 {
				go func() {
					err := service.UsageMonitoring(user.ID)
					if err != nil {
						user.LastVisited = time.Now()
						session.Set("user", userSession)
						session.Save()
					}
				}()
			}
			c.Set("user", user)
		}
		c.Next()
	}
}
