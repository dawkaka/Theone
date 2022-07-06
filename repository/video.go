package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

//Read operations
// List(videos []entity.ID) ([]*entity.Video, error)
// Get(id entity.ID) (*entity.Video, error)

func (v *VideoMongo) Get(id entity.ID) (*entity.Video, error) {
	var video *entity.Video
	err := v.collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&video)
	return video, err
}

func (v *VideoMongo) List(ids []entity.ID) ([]*entity.Video, error) {
	var videos []*entity.Video
	cursor, err := v.collection.Find(context.TODO(), bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})

	if err != nil {
		return videos, err
	}
	err = cursor.All(context.TODO(), &videos)
	return videos, err
}

//Write Operations
// Create(e *entity.Video) (entity.ID, error)
// 	Update(e *entity.Video) error
// 	Delete(id entity.ID) error

func (v *VideoMongo) Create(video *entity.Video) (entity.ID, error) {
	result, err := v.collection.InsertOne(context.TODO(), video)
	return result.InsertedID.(primitive.ObjectID), err
}

func (v *VideoMongo) Update(video *entity.Video) error {
	result, err := v.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: video.ID}},
		bson.D{{Key: "$set", Value: video}},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount < 1 {
		return entity.ErrNotFound
	}
	return nil
}

func (v *VideoMongo) Delete(id entity.ID) error {
	result, err := v.collection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount < 1 {
		return entity.ErrNotFound
	}
	return nil
}
