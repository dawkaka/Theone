package handler

import (
	"github.com/dawkaka/theone/usecase/post"
	"github.com/gin-gonic/gin"
)

func newPost(service post.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}

}

func getPost(service post.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}
}

func listsPosts(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

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
	r.GET("/:coupleName/post/:postID", getPost(service))
	r.GET("/posts/list", listsPosts(service))

	r.POST("/post/new", newPost(service))
	r.PUT("/post/update", updatePost(service))
	r.DELETE("/post/delete", deletePost(service))
}
