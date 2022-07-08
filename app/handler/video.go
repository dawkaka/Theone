package handler

import (
	"github.com/dawkaka/theone/usecase/video"
	"github.com/gin-gonic/gin"
)

func getVideo(service video.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

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

func MakeVideoHandlers(r *gin.Engine, service video.UseCase) {
	r.GET("/:coupleName/video/:videoId", getVideo(service))
	r.GET("/video/list", listVideos(service))
	r.POST("/video/new", newVideo(service))
	r.PUT("/video/update", updateVideo(service))
	r.DELETE("/video/delete", deleteVideo(service))
}
