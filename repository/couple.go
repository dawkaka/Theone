package repository

import (
	"context"
	"errors"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CoupleMongo struct {
	collection *mongo.Collection
}

//NewUserMongo create new repository
func NewCoupleMongo(col *mongo.Collection) *CoupleMongo {
	return &CoupleMongo{
		collection: col,
	}
}

//Read Operations
func (c *CoupleMongo) Get(coupleName string) (entity.Couple, error) {
	opts := options.FindOne().SetProjection(bson.M{"followers": 0, "posts": 0})
	result := entity.Couple{}
	err := c.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "couple_name", Value: coupleName}, {Key: "separated", Value: false}},
		opts,
	).Decode(&result)
	return result, err
}

func (c *CoupleMongo) List(IDs []entity.ID) ([]presentation.CouplePreview, error) {
	cursor, err := c.collection.Find(context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: IDs}}}, {Key: "separated", Value: false}},
	)
	if err != nil {
		return nil, err
	}
	results := []presentation.CouplePreview{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *CoupleMongo) GetCouplePosts(coupleName string, skip int) (entity.Couple, error) {
	var result entity.Couple
	opts := options.FindOne().SetProjection(bson.M{"followers": 0})
	//opts := options.Find().SetSkip(int64(skip)).SetLimit(15)
	// matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "couple_name", Value: coupleName}}}}

	// skipLimitStage := bson.D{{Key: "$skip", Value: int64(skip)}, {Key: "$limit", Value: entity.LimitP + 1}}
	// joinStage := bson.D{
	// 	{
	// 		Key: "$lookup",
	// 		Value: bson.D{
	// 			{Key: "from", Value: "posts"},
	// 			{Key: "localfield", Value: "_id"},
	// 			{Key: "foreignfield", Value: "couple_id"},
	// 			{Key: "as", Value: "couple_posts"},
	// 		},
	// 	},
	// }
	// unwindStage := bson.D{{Key: "$unwind", Value: "$couple_posts"}}

	// cursor, err := c.collection.Aggregate(
	// 	context.TODO(),
	// 	mongo.Pipeline{matchStage, skipLimitStage, joinStage, unwindStage},
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// if err = cursor.All(context.TODO(), &result); err != nil {
	// 	return nil, err
	// }
	// return result, err

	err := c.collection.FindOne(context.TODO(), bson.M{"couple_name": coupleName}, opts).Decode(&result)

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
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "couple_id"},
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

func (c *CoupleMongo) Followers(coupleName string, skip int) ([]entity.ID, error) {
	couple := entity.Couple{}
	opts := options.FindOne().SetProjection(bson.M{"followers": bson.M{"$slice": []int{skip, entity.Limit}}})
	err := c.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "couple_name", Value: coupleName}},
		opts,
	).Decode(&couple)
	if err != nil {
		return nil, err
	}
	return couple.Followers, nil
}

func (c *CoupleMongo) Follower(userID, coupleID entity.ID) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.A{
			bson.D{{
				Key: "$set", Value: bson.M{"followers": bson.M{"$setUnion": []interface{}{"$followers", []entity.ID{userID}}}},
			}},
			bson.D{{
				Key: "$set", Value: bson.M{"followers_count": bson.M{"$size": "$followers"}},
			}},
		},
	)
	return err
}

func (c *CoupleMongo) Unfollow(coupleID, userID entity.ID) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.A{
			bson.D{{
				Key: "$set", Value: bson.M{"followers": bson.M{"$setDifference": []interface{}{"$followers", []entity.ID{userID}}}},
			}},
			bson.D{{
				Key: "$set", Value: bson.M{"followers_count": bson.M{"$size": "$followers"}},
			}},
		},
	)
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

func (u *CoupleMongo) Update(coupleID entity.ID, update entity.UpdateCouple) error {

	_, err := u.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{{Key: "$set", Value: bson.M{
			"bio":            update.Bio,
			"updated_at":     update.UpdatedAt,
			"website":        update.Website,
			"date_commenced": update.DateCommenced,
		}}},
	)
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

func (c *CoupleMongo) ChangeName(coupleID entity.ID, coupleName string) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{{Key: "$set", Value: bson.D{{Key: "couple_name", Value: coupleName}}}},
	)
	return err
}

func (c *CoupleMongo) BreakUp(coupleId entity.ID) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleId,
		bson.D{
			{
				Key:   "$set",
				Value: bson.D{{Key: "separated", Value: true}},
			},
		},
	)
	return err
}

func (c *CoupleMongo) MakeUp(coupleID entity.ID) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{
			{
				Key:   "$set",
				Value: bson.D{{Key: "separated", Value: false}},
			},
		},
	)
	return err
}

func (c *CoupleMongo) Dated(userID, partnerID entity.ID) (entity.ID, error) {
	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	result := entity.Couple{}
	err := c.collection.FindOne(
		context.TODO(),
		bson.D{
			{
				Key: "$or",
				Value: bson.A{
					bson.D{
						{Key: "initiated", Value: userID},
						{Key: "accepted", Value: partnerID},
					},
					bson.D{
						{Key: "initiated", Value: partnerID},
						{Key: "accepted", Value: userID},
					},
				},
			},
		},
		opts,
	).Decode(&result)
	return result.ID, err
}

func (c *CoupleMongo) AddPost(coupleID entity.ID, postID string) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.M{"$push": bson.M{"posts": postID}},
	)
	return err
}

func (c *CoupleMongo) RemovePost(coupleID entity.ID, postID string) error {
	_, err := c.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.M{"$pull": bson.M{"posts": postID}},
	)
	return err
}
