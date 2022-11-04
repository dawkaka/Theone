package repository

import (
	"context"
	"errors"

	"github.com/dawkaka/theone/app/presentation"
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

func (p *PostMongo) Get(coupleID, userID, postID string) (entity.Post, error) {
	var result entity.Post
	ID, err := entity.StringToID(coupleID)
	u, _ := entity.StringToID(userID)
	if err != nil {
		return result, err
	}

	matchStage := bson.D{{Key: "$match", Value: bson.M{"post_id": postID, "couple_id": ID}}}
	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"_id":             1,
				"post_id":         1,
				"couple_id":       1,
				"files":           1,
				"caption":         1,
				"location":        1,
				"likes_count":     1,
				"comments_count":  1,
				"comments_closed": 1,
				"created_at":      1,
				"has_liked":       bson.M{"$in": bson.A{u, "$likes"}},
			},
		},
	}

	cursor, err := p.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, projectStage},
	)

	if err != nil {
		return result, err
	}
	r := []entity.Post{}
	if err = cursor.All(context.TODO(), &r); err != nil {
		return result, err
	}
	if len(r) < 1 {
		return result, mongo.ErrNoDocuments
	}
	return r[0], err
}

func (p *PostMongo) Comments(postID, userID string, skip int) ([]presentation.Comment, error) {
	uID, _ := entity.StringToID(userID)
	ID, err := entity.StringToID(postID)
	if err != nil {
		return nil, err
	}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: ID}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$comments"}}
	sortStage := bson.D{{Key: "$sort", Value: bson.M{"likes_count": -1}}}
	skipStage := bson.D{{Key: "$skip", Value: skip}}
	limitStage := bson.D{{Key: "$limit", Value: entity.Limit}}

	projectStage := bson.D{{
		Key: "$project",
		Value: bson.M{
			"comments":    1,
			"likes_count": bson.M{"$size": "$comments.likes"},
		},
	}}

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
	projectStage2 := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"comment":         "$comments.comment",
				"_id":             "$comments._id",
				"created_at":      "$comments.created_at",
				"likes_count":     1,
				"user_id":         "$comments.user_id",
				"user_name":       "$user.user_name",
				"has_partner":     "$user.has_partner",
				"profile_picture": "$user.profile_picture",
				"has_liked":       bson.M{"$in": bson.A{uID, "$comments.likes"}},
			},
		},
	}

	cursor, err := p.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, unwindStage, projectStage, sortStage, skipStage, limitStage, joinStage, unwindStage2, projectStage2},
	)
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
	res, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "comments_closed", Value: false}},
		bson.D{
			{Key: "$push", Value: bson.D{{Key: "comments", Value: comment}}},
			{Key: "$inc", Value: bson.D{{Key: "comments_count", Value: 1}}},
		},
	)
	if res.ModifiedCount == 0 {
		return entity.ErrNoMatch
	}
	return err
}

func (p *PostMongo) DeleteComment(postID, commentID string, userID entity.ID) error {
	pID, err1 := entity.StringToID(postID)
	cID, err2 := entity.StringToID(commentID)
	if err1 != nil || err2 != nil {
		return errors.New("invalid id")
	}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: pID}}}}
	projectStage := bson.D{{
		Key: "$project",
		Value: bson.M{
			"comments": bson.M{"$filter": bson.M{
				"input": "$comments",
				"as":    "comment",
				"cond":  bson.M{"$eq": []interface{}{"$$comment._id", cID}},
			}},
		},
	}}
	cursor, err := p.collection.Aggregate(context.TODO(), mongo.Pipeline{matchStage, projectStage})

	if err != nil {
		return entity.ErrNoMatch
	}

	post := []entity.Post{}
	if err = cursor.All(context.TODO(), &post); err != nil {
		return err
	}
	if len(post) == 0 || len(post[0].Comments) == 0 || post[0].Comments[0].UserID != userID {
		return entity.ErrNoMatch
	}
	_, err = p.collection.UpdateByID(
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

func (p *PostMongo) Edit(postID, coupleID entity.ID, edit entity.EditPost) error {

	_, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "couple_id", Value: coupleID}},
		bson.M{"$set": bson.M{"caption": edit.Caption, "location": edit.Location}},
	)

	return err
}

func (p *PostMongo) LikeComment(postID, commentID, userID entity.ID) error {
	_, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "comments._id", Value: commentID}},
		bson.M{"$addToSet": bson.M{"comments.$.likes": userID}},
	)
	return err
}

func (p *PostMongo) UnLikeComment(postID, commentID, userID entity.ID) error {
	_, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "comments._id", Value: commentID}},
		bson.M{"$pull": bson.M{"comments.$.likes": userID}},
	)
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

func (p *PostMongo) GetPosts(coupleID entity.ID, userID entity.ID, postIDs []string) ([]presentation.Post, error) {
	result := []presentation.Post{}
	matchStage := bson.D{{Key: "$match", Value: bson.M{"couple_id": coupleID, "post_id": bson.M{"$in": postIDs}}}}
	sortStage := bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}}

	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"_id":             1,
				"created_at":      1,
				"likes_count":     1,
				"comments_count":  1,
				"caption":         1,
				"files":           1,
				"location":        1,
				"comments_closed": 1,
				"post_id":         1,
				"has_liked":       bson.M{"$in": bson.A{userID, "$likes"}},
			},
		},
	}

	cursor, err := p.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, sortStage, projectStage},
	)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.TODO(), &result)

	return result, err
}

func (p *PostMongo) SetClosedComments(postID, coupleID entity.ID, state bool) error {
	_, err := p.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: postID}, {Key: "couple_id", Value: coupleID}},
		bson.M{"$set": bson.M{"comments_closed": state}},
	)
	return err
}
