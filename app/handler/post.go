package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/post"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func newPost(service post.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func getPost(service post.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, postID := ctx.Param("coupleName"), ctx.Param("postId")
		if !validator.IsUserName(coupleName) || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "UserNotFound"))
				return
			}
		}
		post, err := service.GetPost(couple.ID.String(), postID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"video": post})
	}
}

func newComment(service post.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(ctx.Request.Header, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}

		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		var comment entity.Comment
		err = ctx.ShouldBind(comment)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		comment.UserID = user.ID.String()
		comment.CreatedAt = time.Now()
		err = service.NewComment(postID, comment)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		notif := entity.Notification{
			Type:    "comment",
			Message: inter.LocalizeWithUserName(utils.GetLang(ctx.Request.Header), user.Name, "PostCommentNotif"),
		}
		_ = userService.NotifyCouple([2]entity.ID{post.InitiatedID, post.AcceptedID}, notif)

		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "CommentAdded"))
	}
}

func like(service post.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(ctx.Request.Header, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		err = service.LikePost(postID, user.ID.String())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		notif := entity.Notification{
			Type:    "like",
			Message: inter.LocalizeWithUserName(utils.GetLang(ctx.Request.Header), user.Name, "PostLikeNotif"),
		}
		_ = userService.NotifyCouple([2]entity.ID{post.InitiatedID, post.AcceptedID}, notif)
		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "PostLiked"))
	}
}

func updatePost(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func deletePost(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakePostHandlers(r *gin.Engine, service post.UseCase, coupleService couple.UseCase, userService user.UseCase) {
	r.GET("/post/:coupleName/:postID", getPost(service, coupleService))
	r.POST("/post/new", newPost(service))
	r.POST("/post/new-comment/:postID", newComment(service, userService))
	r.PATCH("/post/like/:postID", like(service, userService))
	r.PUT("/post/update", updatePost(service))
	r.DELETE("/post/delete", deletePost(service))
}
