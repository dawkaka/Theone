package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dawkaka/theone/config"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	r := gin.Default()

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.DB_HOST))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	usersCollection := client.Database(config.DB_DATABASE).Collection("users")
	couplesCollection := client.Database(config.DB_DATABASE).Collection("couples")
	postsCollection := client.Database(config.DB_DATABASE).Collection("posts")
	videosCollection := client.Database(config.DB_DATABASE).Collection("videos")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run(fmt.Sprintf(":%d", config.API_PORT)) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
