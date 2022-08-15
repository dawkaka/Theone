package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	opts := options.FindOne().SetProjection(bson.M{"likes": 0, "comments": 0})
	err = p.collection.FindOne(
		context.TODO(),
		bson.D{
			{Key: "post_id", Value: postID},
			{Key: "couple_id", Value: ID},
		},
		opts,
	).Decode(&result)
	return &result, err
}
func (p *PostMongo) Comments(videoID string, skip int) ([]presentation.Comment, error) {
	ID, err := entity.StringToID(videoID)
	if err != nil {
		return nil, err
	}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: ID}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$comments"}}
	skipStage := bson.D{{Key: "$skip", Value: skip}}
	limitStage := bson.D{{Key: "$limit", Value: entity.Limit}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "comments.user_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "user"},
			},
		},
	}
	unwindStage2 := bson.D{{Key: "$unwind", Value: "$user"}}
	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"_id":         1,
				"comment":     "$comments.comment",
				"created_at":  "$comments.created_at",
				"likes_count": "$comments.likes_count",
				"user_name":   "$user.user_name",
				"has_partner": "$user.has_partner",
			},
		},
	}

	cursor, err := p.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, unwindStage, skipStage, limitStage, joinStage, unwindStage2, projectStage},
	)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	var comments []presentation.Comment
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
	fmt.Println(err)
	if err != nil {
		return primitive.ObjectID{}, err
	}
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

func (p *PostMongo) DeleteComment(postID, commentID string, userID entity.ID) error {
	pID, err1 := entity.StringToID(postID)
	cID, err2 := entity.StringToID(commentID)
	if err1 != nil || err2 != nil {
		return errors.New("invalid id")
	}
	_, err := p.collection.UpdateByID(
		context.TODO(),
		pID,
		bson.A{
			bson.D{{Key: "$set", Value: bson.M{"comments": bson.M{"$filter": bson.M{
				"input": "$comments",
				"as":    "comment",
				"cond":  bson.M{"$ne": []interface{}{"$$comment._id", cID}},
			}}}}},
			bson.D{{Key: "$set", Value: bson.M{"comments_count": bson.M{"$size": "$comments"}}}},
		},
	)
	return err
}

func (p *PostMongo) Like(postID, userID entity.ID) error {
	_, err := p.collection.UpdateByID(
		context.TODO(),
		postID,
		bson.A{
			bson.D{{
				Key: "$set", Value: bson.M{"likes": bson.M{"$setUnion": []interface{}{"$likes", []entity.ID{userID}}}},
			}},
			bson.D{{
				Key: "$set", Value: bson.M{"likes_count": bson.M{"$size": "$likes"}},
			}},
		},
	)

	return err
}

func (p *PostMongo) UnLike(postID, userID entity.ID) error {
	_, err := p.collection.UpdateByID(
		context.TODO(),
		postID,
		bson.A{
			bson.D{{
				Key: "$set", Value: bson.M{"likes": bson.M{"$setDifference": []interface{}{"$likes", []entity.ID{userID}}}},
			}},
			bson.D{{
				Key: "$set", Value: bson.M{"likes_count": bson.M{"$size": "$likes"}},
			}},
		},
	)

	return err
}

func (p *PostMongo) Edit(postID, coupleID entity.ID, newCaption string) error {
	res, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "couple_id", Value: coupleID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "caption", Value: newCaption}}}},
	)
	if res.ModifiedCount == 0 {
		return errors.New("no match found")
	}
	return err
}

func (p *PostMongo) Delete(coupleID, postID entity.ID) error {
	result, err := p.collection.DeleteOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "couple_id", Value: coupleID}},
	)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}
