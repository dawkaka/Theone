package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/dawkaka/theone/usecase/video"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func getVideo(service video.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, videoID := ctx.Param("coupleName"), ctx.Param("videoId")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsUserName(coupleName) || strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "NotFound"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		video, err := service.GetVideo(couple.ID.String(), videoID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
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
func videoComment(service video.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		videoID := ctx.Param("videoID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		video, err := service.GetVideoByID(videoID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		var comment entity.Comment
		err = ctx.ShouldBind(comment)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
		}
		comment.UserID = user.ID.String()
		comment.CreatedAt = time.Now()
		err = service.NewComment(videoID, comment)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		notif := entity.Notification{
			Type:    "comment",
			Message: inter.LocalizeWithUserName(lang, user.Name, "VideoCommentNotif"),
		}
		_ = userService.NotifyCouple([2]entity.ID{video.InitiatedID, video.AcceptedID}, notif)

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "CommentAdded"))
	}
}

func likeVideo(service video.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		videoID := ctx.Param("videoID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		video, err := service.GetVideoByID(videoID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.LikeVideo(videoID, user.ID.String())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		notif := entity.Notification{
			Type:    "like",
			Message: inter.LocalizeWithUserName(lang, user.Name, "VideoLikeNotif"),
		}
		_ = userService.NotifyCouple([2]entity.ID{video.InitiatedID, video.AcceptedID}, notif)
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "VideoLiked"))
	}
}

func videoComments(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		skip, err := strconv.Atoi(ctx.Param("skip"))
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		videoID := ctx.Param("videoID")
		comments, err := service.GetComments(videoID, skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}

		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(comments) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"comments": comments, "pagination": page})
	}
}

func deleteVideoComment(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		commentID, videoID := ctx.Param("commentID"), ctx.Param("videoID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(commentID) == "" || strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		userID := sessions.Default(ctx).Get("user").(entity.UserSession).ID
		err := service.DeleteComment(videoID, commentID, userID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "CommentDeleted"))
	}
}
func unLikeVideo(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		videoID := ctx.Param("videoID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(videoID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		userID := user.ID
		err := service.UnLikeVideo(videoID, userID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "UnlikeVideo"))
	}
}

func editVideoCaption(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		var Caption struct {
			Caption string `json:"caption"`
		}
		err := ctx.ShouldBindJSON(&Caption)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		err = service.EditCaption(postID, user.CoupleID, Caption.Caption)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "PostEdited"))
	}
}

func deleteVideo(service video.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeVideoHandlers(r *gin.Engine, service video.UseCase, coupleService couple.UseCase, userService user.UseCase) {
	r.GET("/video/:coupleName/:videoId", getVideo(service, coupleService))
	r.GET("/video/list", listVideos(service))
	r.GET("/video/comments/:videoID", videoComments(service))
	r.DELETE("/video/comment/:videoID/:commentID", deleteVideoComment(service))
	r.POST("/video/new-comment/:videoID", videoComment(service, userService))
	r.PATCH("/video/like/:videoID", likeVideo(service, userService))
	r.PATCH("/video/unlike/:videoID", unLikeVideo(service))
	r.PATCH("/video/edit/:videoID", editVideoCaption(service))
	r.POST("/video/new", newVideo(service))
	r.PUT("/video/update", updateVideo(service))
	r.DELETE("/video/delete", deleteVideo(service))
}
