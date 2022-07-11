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

func (p *PostMongo) Get(coupleName, postID string) (*entity.Post, error) {
	var result entity.Post
	err := p.collection.FindOne(
		context.TODO(),
		bson.D{
			{Key: "post_id", Value: postID},
			{Key: "couple_name", Value: coupleName},
		},
	).Decode(&result)

	return &result, err
}

func (p *PostMongo) List(ids []entity.ID) ([]*entity.Post, error) {

	var result []*entity.Post
	cursor, err := p.collection.Find(
		context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}},
	)
	if err != nil {
		return result, err
	}
	err = cursor.All(context.TODO(), &result)
	return result, err
}

//Write Operations

func (p *PostMongo) Create(post *entity.Post) (entity.ID, error) {
	result, err := p.collection.InsertOne(context.TODO(), post)
	return result.InsertedID.(primitive.ObjectID), err
}

func (p *PostMongo) Update(post *entity.Post) error {
	result, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: post.ID}},
		bson.D{{Key: "$set", Value: post}},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount < 1 {
		return entity.ErrNotFound
	}
	return nil
}

func (p *PostMongo) Delete(id entity.ID) error {
	result, err := p.collection.DeleteOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: id}},
	)
	if err != nil {
		return err
	}
	if result.DeletedCount < 1 {
		return entity.ErrNotFound
	}

	return nil
}
