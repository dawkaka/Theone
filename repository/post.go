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

func (p *PostMongo) Get(coupleID, postID string) (*entity.Post, error) {
	var result entity.Post
	ID, err := entity.StringToID(coupleID)
	if err != nil {
		return nil, err
	}
	err = p.collection.FindOne(
		context.TODO(),
		bson.D{
			{Key: "post_id", Value: postID},
			{Key: "couple_id", Value: ID},
		},
	).Decode(&result)

	return &result, err
}
func (p *PostMongo) Comments(videoID string, skip int) ([]entity.Comment, error) {
	ID, err := entity.StringToID(videoID)
	if err != nil {
		return nil, err
	}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: ID}}}}
	sliceStage := bson.D{{Key: "$slice", Value: []interface{}{"$comments", skip, entity.Limit}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$comments"}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localfield", Value: "comments"},
				{Key: "foreignfield", Value: "_id"},
				{Key: "as", Value: "user"},
			},
		},
	}

	cursor, err := p.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, sliceStage, unwindStage, joinStage},
	)
	if err != nil {
		return nil, err
	}
	var comments []entity.Comment

	if err = cursor.All(context.TODO(), &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (p *PostMongo) GetByID(ID entity.ID) (entity.Post, error) {
	var result entity.Post
	err := p.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: ID}},
	).Decode((&result))
	return result, err
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

func (p *PostMongo) AddComment(postID entity.ID, comment entity.Comment) error {
	_, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}},
		bson.D{
			{Key: "$push", Value: bson.D{{Key: "comments", Value: comment}}},
			{Key: "$inc", Value: bson.D{{Key: "comments_count", Value: 1}}},
		},
	)
	return err
}

func (p *PostMongo) Like(postID, userID entity.ID) error {
	_, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}},
		bson.D{
			{Key: "$push", Value: bson.D{{Key: "likes", Value: userID}}},
			{Key: "$inc", Value: bson.D{{Key: "likes_count", Value: 1}}},
		},
	)
	return err
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
