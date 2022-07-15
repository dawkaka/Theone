package repository

import (
	"context"
	"errors"

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
func (v *VideoMongo) Get(coupleID, videoID string) (*entity.Video, error) {
	var video *entity.Video
	ID, err := entity.StringToID(coupleID)
	if err != nil {
		return video, err
	}
	err = v.collection.FindOne(
		context.TODO(),
		bson.D{
			{Key: "video_id", Value: videoID},
			{Key: "couple_id", Value: ID},
		},
	).Decode(&video)
	return video, err
}

func (v *VideoMongo) GetByID(id string) (entity.Video, error) {
	ID, err := entity.StringToID(id)
	if err != nil {
		return entity.Video{}, errors.New("parsing id: failed to convert string to id")
	}
	var video entity.Video
	err = v.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: ID}},
	).Decode(&video)
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

func (v *VideoMongo) AddComment(videoID entity.ID, comment entity.Comment) error {
	_, err := v.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: videoID}},
		bson.D{
			{Key: "$push", Value: bson.D{{Key: "comments", Value: comment}}},
			{Key: "$inc", Value: "comments_count"},
		},
	)
	return err
}

func (v *VideoMongo) Like(videoID, userID entity.ID) error {
	_, err := v.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: videoID}},
		bson.D{
			{Key: "$push", Value: bson.D{{Key: "likes", Value: userID}}},
			{Key: "$inc", Value: "likes_count"},
		},
	)
	return err
}
