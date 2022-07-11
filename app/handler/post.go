package handler

import (
	"net/http"
	"strings"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/post"
	"github.com/gin-gonic/gin"
)

func newPost(service post.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func getPost(service post.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, postID := ctx.Param("coupleName"), ctx.Param("postId")
		if !validator.IsUserName(coupleName) || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "NotFound"))
			return
		}
		post, err := service.GetPost(coupleName, postID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"post": post})
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

func MakePostHandlers(r *gin.Engine, service post.UseCase) {
	r.GET("/:coupleName/:postID", getPost(service))
	r.POST("/post/new", newPost(service))
	r.PUT("/post/update", updatePost(service))
	r.DELETE("/post/delete", deletePost(service))
}
