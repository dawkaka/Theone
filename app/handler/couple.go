package handler

import (
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/gin-gonic/gin"
)

func newCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func getCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}
}

func updateCouple(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeCoupleHandlers(r *gin.Engine, service couple.UseCase) {
	r.GET("/:coupleName", getCouple(service))
	r.POST("/couple/new", newCouple(service))
	r.PUT("/couple/update", updateCouple(service))

}
