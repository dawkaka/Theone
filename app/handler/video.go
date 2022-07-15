package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/video"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func getVideo(service video.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, videoID := ctx.Param("coupleName"), ctx.Param("videoId")
		if !validator.IsUserName(coupleName) || strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "NotFound"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		video, err := service.GetVideo(couple.ID.String(), videoID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"video": video})
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
func videoComment(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		videoID := ctx.Param("videoID")
		if strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
			return
		}
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		var comment entity.Comment
		err := ctx.ShouldBind(comment)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		comment.UserID = user.ID.String()
		comment.CreatedAt = time.Now()
		err = service.NewComment(videoID, comment)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "CommentAdded"))
	}
}

func likeVideo(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		videoID := ctx.Param("videoID")
		if strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
			return
		}
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		err := service.LikeVideo(videoID, user.ID.String())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "VideoLiked"))
	}
}

func deleteVideo(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeVideoHandlers(r *gin.Engine, service video.UseCase, coupleService couple.UseCase) {
	r.GET("/video/:coupleName/:videoId", getVideo(service, coupleService))
	r.GET("/video/list", listVideos(service))
	r.POST("/video/new-comment/:videoID", videoComment(service))
	r.PATCH("/video/like/:videoID", likeVideo(service))
	r.POST("/video/new", newVideo(service))
	r.PUT("/video/update", updateVideo(service))
	r.DELETE("/video/delete", deleteVideo(service))
}
