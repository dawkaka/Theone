package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type PostMongo struct {
	collection *mongo.Collection
}

//NewUserMySQL create new repository
func NewPostMongo(col *mongo.Collection) *PostMongo {
	return &PostMongo{
		collection: col,
	}
}
