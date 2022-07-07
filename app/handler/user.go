package handler

import (
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-gonic/gin"
)

func signup(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func getUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}
}

func updateUser(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func deleteUser(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func searchUsers(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeUserHandlers(r *gin.Engine, service user.UseCase) {

	r.GET("/user/:userName", getUser(service))
	r.GET("/user/search/:query", searchUsers(service))
	r.POST("/user/signup", signup(service))
	r.PUT("/user/update", updateUser(service))
	r.DELETE("/user/delete-account", deleteUser(service))
}
