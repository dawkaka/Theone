package main

import (
	"context"
	"encoding/gob"
	"fmt"

	"github.com/dawkaka/theone/app/handler"
	"github.com/dawkaka/theone/app/middlewares"
	"github.com/dawkaka/theone/config"
	"github.com/dawkaka/theone/entity"
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

	db := client.Database(config.DB_DATABASE)

	usersRepo := repository.NewUserMongo(db.Collection("users"))
	couplesRepo := repository.NewCoupleMongo(db.Collection("couples"))
	postsRepo := repository.NewPostMongo(db.Collection("posts"))
	videosRepo := repository.NewVideoMongo(db.Collection("videos"))
	userMessageRepo := repository.NewUserCoupleMessageRepo(db.Collection(("group-messages")))
	coupleMessageRepo := repository.NewCoupleMessageRepo(db.Collection("couple-messages"))
	reportsRepo := repository.NewReportRepo(db.Collection("reports"))

	videoService := video.NewService(videosRepo)
	userService := user.NewService(usersRepo)
	postService := post.NewService(postsRepo)
	coupleService := couple.NewService(couplesRepo)
	gob.Register(entity.UserSession{})
	r.Use(sessions.Sessions("session", store))
	r.Use(middlewares.CORSMiddleware())
	handler.MakeUserHandlers(r, userService, coupleService, userMessageRepo)
	handler.MakeCoupleHandlers(r, coupleService, userService, coupleMessageRepo, userMessageRepo)
	handler.MakePostHandlers(r, postService, coupleService, userService, reportsRepo)
	handler.MakeVideoHandlers(r, videoService, coupleService, userService)
	r.Run(fmt.Sprintf(":%d", config.API_PORT))
}
