package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/utils"
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
func (c *CoupleMongo) Get(coupleName string, userID entity.ID) (entity.Couple, error) {
	result := entity.Couple{}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "couple_name", Value: coupleName}, {Key: "separated", Value: false}}}}
	projectStage := bson.D{{
		Key: "$project",
		Value: bson.M{
			"couple_name":     1,
			"profile_picture": 1,
			"married":         1,
			"cover_picture":   1,
			"date_commenced":  1,
			"verified":        1,
			"_id":             1,
			"initiated":       1,
			"accepted":        1,
			"bio":             1,
			"website":         1,
			"followers_count": 1,
			"is_following":    bson.M{"$in": bson.A{userID, "$followers"}},
			"post_count":      bson.M{"$size": "$posts"}},
	}}

	cursor, err := c.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, projectStage},
	)
	if err != nil {
		return result, err
	}
	results := []entity.Couple{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return result, err
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return result, entity.ErrCoupleNotFound
}

func (c *CoupleMongo) List(IDs []entity.ID, userID entity.ID) ([]presentation.CouplePreview, error) {
	matchStage := bson.D{
		{
			Key: "$match", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$in", Value: IDs}}},
				{Key: "separated", Value: false},
				{Key: "$expr", Value: bson.M{"$eq": bson.A{bson.M{"$size": bson.M{
					"$filter": bson.M{"input": "$blocked", "as": "cbl", "cond": bson.M{"$eq": bson.A{"$cbl", userID}}},
				}}, 0}}},
			},
		},
	}
	projectStage := bson.D{{
		Key: "$project",
		Value: bson.M{
			"couple_name":     1,
			"profile_picture": 1,
			"married":         1,
			"verified":        1,
			"is_following":    bson.M{"$in": bson.A{userID, "$followers"}},
		},
	}}

	cursor, err := c.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, projectStage},
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

func (c *CoupleMongo) Search(query string, userID entity.ID) ([]presentation.CouplePreview, error) {
	opts := options.Find().SetSort(bson.D{{Key: "followers_count", Value: -1}}).SetProjection(bson.M{"couple_name": 1, "profile_picture": 1, "verified": 1, "married": 1})
	results := []presentation.CouplePreview{}
	cursor, err := c.collection.Find(
		context.TODO(),
		bson.M{"couple_name": bson.M{"$regex": primitive.Regex{Pattern: "^" + query, Options: "i"}},
			"$expr": bson.M{"$eq": bson.A{bson.M{"$size": bson.M{
				"$setIntersection": []interface{}{"$blocked", []entity.ID{userID}},
			}}, 0}},
		},
		opts,
	)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return results, err

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

func (c *CoupleMongo) UpdateStatus(coupleID entity.ID, married bool) error {
	_, err := c.collection.UpdateByID(context.TODO(), coupleID, bson.D{{Key: "$set", Value: bson.M{"married": married}}})
	return err
}

func (c *CoupleMongo) FollowersToNotify(coupleID entity.ID, skip int) ([]entity.ID, error) {
	couple := entity.Couple{}
	opts := options.FindOne().SetProjection(bson.M{"followers": bson.M{"$slice": []int{skip, 1000}}})
	err := c.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: coupleID}},
		opts,
	).Decode(&couple)
	if err != nil {
		return nil, err
	}
	return couple.Followers, nil
}

func (c *CoupleMongo) SuggestedAccounts(exempted []entity.ID, country string) ([]presentation.CouplePreview, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: exempted}}}, {Key: "separated", Value: false}, {Key: "country", Value: country}}}}
	sortStage := bson.D{{Key: "$sort", Value: bson.M{"followers_count": -1}}}
	limitStage := bson.D{{Key: "$limit", Value: 20}}

	projectStage := bson.D{{
		Key: "$project",
		Value: bson.M{
			"couple_name":     1,
			"profile_picture": 1,
			"married":         1,
			"verified":        1,
		},
	}}

	cursor, err := c.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, sortStage, limitStage, projectStage},
	)
	if err != nil {
		return nil, err
	}
	results := []presentation.CouplePreview{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	res2 := []presentation.CouplePreview{}
	if len(results) < 20 {
		category := utils.GetCategory(1, country)
		fechted := []string{}
		for _, val := range results {
			fechted = append(fechted, val.CoupleName)
		}
		matchStage := bson.D{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$nin", Value: exempted}}},
			{Key: "separated", Value: false},
			{Key: "country", Value: bson.M{"$in": category}},
			{Key: "couple_name", Value: bson.M{"$nin": fechted}},
		},
		}}
		sortStage := bson.D{{Key: "$sort", Value: bson.M{"followers_count": -1}}}
		limitStage := bson.D{{Key: "$limit", Value: 20}}

		projectStage := bson.D{{
			Key: "$project",
			Value: bson.M{
				"couple_name":     1,
				"profile_picture": 1,
				"married":         1,
				"verified":        1,
			},
		}}

		cursor, err := c.collection.Aggregate(
			context.TODO(),
			mongo.Pipeline{matchStage, sortStage, limitStage, projectStage},
		)
		if err != nil {
			return nil, err
		}
		cursor.All(context.TODO(), &res2)
	}
	return append(results, res2...), nil
}

func (c *CoupleMongo) Block(coupleID, userID entity.ID) error {
	_, err := c.collection.UpdateByID(context.TODO(), coupleID, bson.D{{Key: "$addToSet", Value: bson.M{"blocked": userID}}})
	return err
}

func (c *CoupleMongo) IsBlocked(coupleName string, userID entity.ID) (bool, error) {
	opts := options.FindOne().SetProjection(bson.D{{Key: "couple_name", Value: 1}})
	err := c.collection.FindOne(context.TODO(), bson.D{{Key: "couple_name", Value: coupleName}, {Key: "blocked", Value: userID}}, opts)
	if err.Err() != nil {
		if err.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err.Err()
	}
	return true, nil
}
