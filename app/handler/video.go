package handler

import (
	"net/http"
	"strings"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/video"
	"github.com/gin-gonic/gin"
)

func getVideo(service video.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, postID := ctx.Param("coupleName"), ctx.Param("postId")
		if !validator.IsUserName(coupleName) || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "NotFound"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		video, err := service.GetVideo(couple.ID.String(), postID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"post": video})
	}
}

func listVideos(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func newVideo(service video.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func updateVideo(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func deleteVideo(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeVideoHandlers(r *gin.Engine, service video.UseCase, coupleService couple.UseCase) {
	r.GET("/video/:coupleName/:videoId", getVideo(service, coupleService))
	r.GET("/video/list", listVideos(service))
	r.POST("/video/new", newVideo(service))
	r.PUT("/video/update", updateVideo(service))
	r.DELETE("/video/delete", deleteVideo(service))
}
