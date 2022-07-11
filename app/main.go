package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dawkaka/theone/app/handler"
	"github.com/dawkaka/theone/config"
	"github.com/dawkaka/theone/repository"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/post"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/dawkaka/theone/usecase/video"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	r := gin.Default()
	store, err := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	if err != nil {
		panic(err)
	}
	r.Use(sessions.Sessions("session", store))

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.DB_HOST))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	usersRepo := repository.NewUserMongo(client.Database(config.DB_DATABASE).Collection("users"))
	couplesRepo := repository.NewCoupleMongo(client.Database(config.DB_DATABASE).Collection("couples"))
	postsRepo := repository.NewPostMongo(client.Database(config.DB_DATABASE).Collection("posts"))
	videosRepo := repository.NewVideoMongo(client.Database(config.DB_DATABASE).Collection("videos"))

	videoService := video.NewService(videosRepo)
	userService := user.NewService(usersRepo)
	postService := post.NewService(postsRepo)
	coupleService := couple.NewService(couplesRepo)

	handler.MakeUserHandlers(r, userService)
	handler.MakeCoupleHandlers(r, coupleService)
	handler.MakePostHandlers(r, postService)
	handler.MakeVideoHandlers(r, videoService)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run(fmt.Sprintf(":%d", config.API_PORT)) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
