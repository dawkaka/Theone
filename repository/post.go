package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

//Read opertions

//List(userName string) ([]*entity.Post, error)

func (p *PostMongo) Get(id entity.ID) (*entity.Post, error) {
	var result entity.Post
	err := p.collection.FindOne(context.TODO(), bson.D{{Key: "id", Value: id}}).Decode(&result)
	return &result, err
}

func (p *PostMongo) List(id entity.ID) ([]*entity.Post, error) {

	var result []*entity.Post
	cursor, err := p.collection.Find(context.TODO(), bson.D{{Key: "id", Value: id}})
	if err != nil {
		return result, err
	}
	err = cursor.All(context.TODO(), &result)
	return result, err
}

//Write Operations
// Create(e *entity.Post) (entity.ID, error)
// Update(e *entity.Post) error
// Delete(id entity.ID) error

func (p *PostMongo) Create(post entity.Post) (entity.ID, error) {

	result, err := p.collection.InsertOne(context.TODO(), post)

	return result.InsertedID.(primitive.ObjectID), err

}
