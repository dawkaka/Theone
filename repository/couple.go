package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type CoupleMongo struct {
	collection *mongo.Collection
}

//NewUserMySQL create new repository
func NewCoupleMongo(col *mongo.Collection) *CoupleMongo {
	return &CoupleMongo{
		collection: col,
	}
}
