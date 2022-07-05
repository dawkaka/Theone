package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type UserMongo struct {
	collection *mongo.Collection
}

//NewUserMySQL create new repository
func NewUserMongo(col *mongo.Collection) *UserMongo {
	return &UserMongo{
		collection: col,
	}
}
