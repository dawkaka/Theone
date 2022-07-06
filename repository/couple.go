package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
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

//Read Operations
func (c *CoupleMongo) Get(coupleName string) (entity.Couple, error) {
	var result entity.Couple

	err := c.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "couple_name", Value: coupleName}},
	).Decode(&result)

	return result, err
}

//Write Operations
func (c *CoupleMongo) Create(couple entity.Couple) error {
	_, err := c.collection.InsertOne(context.TODO(), couple)
	return err
}

func (c *CoupleMongo) Update(couple entity.Couple) error {

	result, err := c.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "user_name", Value: couple.CoupleName}},
		bson.D{{Key: "$set", Value: couple}},
	)

	if result.ModifiedCount != 1 {
		return entity.ErrNotFound
	}
	return err
}
