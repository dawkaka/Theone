package handler

import (
	"net/http"
	"strconv"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/pkg/validator"
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
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "Invalid couple name"))
			return
		}

		couple, err := service.GetCouple(coupleName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "Something went wrong"))
			return
		}
		pCouple := presentation.CoupleProfile{
			CoupleName:     couple.CoupleName,
			AcceptedAt:     couple.AcceptedAt,
			Bio:            couple.Bio,
			Status:         couple.Status,
			FollowersCount: couple.FollowersCount,
			ProfilePicture: couple.ProfilePicture,
			CoverPicture:   couple.CoverPicture,
			PostCount:      couple.PostCount,
		}
		ctx.JSON(http.StatusOK, gin.H{"couple": pCouple})
	}
}

func getCouplePosts(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipPosts, err := strconv.Atoi(skip)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		posts, err := service.GetCouplePosts(coupleName, skipPosts)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"posts": posts})
	}
}

func getCoupleVideos(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipPosts, err := strconv.Atoi(skip)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		posts, err := service.GetCoupleVideos(coupleName, skipPosts)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"videos": posts})
	}
}

func updateCouple(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeCoupleHandlers(r *gin.Engine, service couple.UseCase) {
	r.GET("/:coupleName", getCouple(service))
	r.GET("/:coupleName/posts/:skip", getCouplePosts(service))
	r.GET("/:coupleName/videos/:skip", getCoupleVideos(service))
	r.POST("/couple/new", newCouple(service))
	r.PUT("/couple/update", updateCouple(service))

}
