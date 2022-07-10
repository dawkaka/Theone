package handler

import (
	"net/http"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/pkg/password/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/gin-gonic/gin"
)

func newCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func getCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error("Invalid couple name"))
			return
		}

		couple, err := service.GetCouple(coupleName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error("Something went wrong"))
			return
		}
		pCouple := presentation.CoupleProfile{
			CoupleName:     couple.CoupleName,
			AcceptedAt:     couple.AcceptedAt,
			Bio:            couple.Bio,
			Status:         couple.Status,
			Followers:      len(couple.Followers),
			ProfilePicture: couple.ProfilePicture,
			CoverPicture:   couple.CoverPicture,
			PostCount:      couple.PostCount,
		}
		ctx.JSON(http.StatusOK, gin.H{"couple": pCouple})
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
