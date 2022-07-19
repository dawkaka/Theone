package repository

import (
	"context"
	"errors"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (c *CoupleMongo) GetCouplePosts(coupleName string, skip int) ([]entity.Post, error) {
	var result []entity.Post
	//opts := options.Find().SetSkip(int64(skip)).SetLimit(15)
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "couple_name", Value: coupleName}}}}
	skipLimitStage := bson.D{{Key: "$skip", Value: int64(skip)}, {Key: "$limit", Value: entity.LimitP + 1}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "posts"},
				{Key: "localfield", Value: "_id"},
				{Key: "foreignfield", Value: "couple_id"},
				{Key: "as", Value: "couple_posts"},
			},
		},
	}
	unwindStage := bson.D{{Key: "$unwind", Value: "$couple_posts"}}

	cursor, err := c.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, skipLimitStage, joinStage, unwindStage},
	)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &result); err != nil {
		return nil, err
	}
	return result, err
}

func (c *CoupleMongo) GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error) {
	var result []entity.Video
	//opts := options.Find().SetSkip(int64(skip)).SetLimit(15)
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "couple_name", Value: coupleName}}}}
	skipLimitStage := bson.D{{Key: "$skip", Value: int64(skip)}, {Key: "$limit", Value: entity.LimitP + 1}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "videos"},
				{Key: "localfield", Value: "_id"},
				{Key: "foreignfield", Value: "couple_id"},
				{Key: "as", Value: "couple_videos"},
			},
		},
	}
	unwindStage := bson.D{{Key: "$unwind", Value: "$couple_videos"}}

	cursor, err := c.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, skipLimitStage, joinStage, unwindStage},
	)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &result); err != nil {
		return nil, err
	}
	return result, err
}

func (c *CoupleMongo) Followers(coupleName string, skip int) ([]entity.Follower, error) {
	var followers []entity.Follower

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "couple_name", Value: coupleName}}}}
	skipLimitStage := bson.D{{Key: "$skip", Value: int64(skip)}, {Key: "$limit", Value: entity.Limit + 1}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$followers"}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localfield", Value: "followers"},
				{Key: "foreignfield", Value: "_id"},
				{Key: "as", Value: "couple_followers"},
			},
		},
	}
	unwindStage2 := bson.D{{Key: "$unwind", Value: "$couple_followers"}}
	cursor, err := c.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, unwindStage, skipLimitStage, joinStage, unwindStage2},
	)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &followers); err != nil {
		return nil, err
	}
	return followers, nil
}

func (c *CoupleMongo) Follower(userID, coupleID entity.ID) error {
	result, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{
			{Key: "$inc", Value: bson.D{{Key: "followers_count", Value: 1}}},
			{Key: "$push", Value: bson.D{{Key: "followers", Value: userID}}},
		},
	)
	if result.ModifiedCount < 1 {
		return errors.New("user follow: something went wrong")
	}
	return err
}

func (c *CoupleMongo) Unfollow(userID, coupleId entity.ID) error {
	result, err := c.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{
			{Key: "$inc", Value: bson.D{{Key: "followers_count", Value: -1}}},
			{Key: "$push", Value: bson.D{{Key: "followers", Value: coupleId}}},
		},
	)
	if result.ModifiedCount < 1 {
		return errors.New("user follow: something went wrong")
	}
	return err

}

//Write Operations
func (c *CoupleMongo) Create(couple entity.Couple) (entity.ID, error) {
	var id primitive.ObjectID
	result, err := c.collection.InsertOne(context.TODO(), couple)
	if err == nil {
		id = result.InsertedID.(primitive.ObjectID)
	}
	return id, err
}

func (c *CoupleMongo) Update(couple entity.Couple) error {

	result, err := c.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "couple_name", Value: couple.CoupleName}},
		bson.D{{Key: "$set", Value: couple}},
	)

	if result.ModifiedCount != 1 {
		return entity.ErrNotFound
	}
	return err
}

func (c *CoupleMongo) UpdateProfilePic(fileName string, coupleID entity.ID) error {
	result, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{{Key: "$set", Value: bson.D{{Key: "profile_picture", Value: fileName}}}},
	)
	if result.MatchedCount == 0 {
		return errors.New("no match")
	}
	return err
}

func (c *CoupleMongo) UpdateCoverPic(fileName string, coupleID entity.ID) error {
	result, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{{Key: "$set", Value: bson.D{{Key: "cover_picture", Value: fileName}}}},
	)
	if result.MatchedCount == 0 {
		return errors.New("no match")
	}
	return err
}
