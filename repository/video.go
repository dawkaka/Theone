package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type VideoMongo struct {
	collection *mongo.Collection
}

//NewUserMySQL create new repository
func NewVideoMongo(col *mongo.Collection) *VideoMongo {
	return &VideoMongo{
		collection: col,
	}
}
