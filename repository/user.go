package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserMongo struct {
	collection *mongo.Collection
}

//NewUserMySQL create new repository
func NewUserMongo(col *mongo.Collection) *UserMongo {
	return &UserMongo{
		collection: col,
	}
}

//Read Methods
func (u *UserMongo) Get(userName string) (entity.User, error) {
	result := entity.User{}
	opts := options.FindOne().SetProjection(
		bson.D{
			{Key: "first_name", Value: 1}, {Key: "last_name", Value: 1},
			{Key: "user_name", Value: 1}, {Key: "profile_picture", Value: 1},
			{Key: "has_partner", Value: 1}, {Key: "show_pictures", Value: 1},
			{Key: "following_count", Value: 1}, {Key: "likes_count", Value: 1},
		})
	err := u.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "user_name", Value: userName}},
		opts,
	).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return result, entity.ErrUserNotFound
	}
	return result, err
}

func (u *UserMongo) Search(query string) ([]entity.User, error) {
	opts := options.Find().SetProjection(
		bson.D{
			{Key: "first_name", Value: 1}, {Key: "last_name", Value: 1},
			{Key: "user_name", Value: 1}, {Key: "profile_picture", Value: 1},
			{Key: "has_partner", Value: 1},
		})
	cursor, err := u.collection.Find(
		context.TODO(),
		bson.D{
			{
				Key: "$or",
				Value: bson.A{
					bson.D{
						{Key: "user_name", Value: bson.M{"$regex": primitive.Regex{Pattern: "^" + query, Options: "i"}}},
					},
					bson.D{
						{Key: "first_name", Value: bson.M{"$regex": primitive.Regex{Pattern: "^" + query, Options: "i"}}},
					},
					bson.D{
						{Key: "last_name", Value: bson.M{"$regex": primitive.Regex{Pattern: "^" + query, Options: "i"}}},
					},
				},
			},
		},
		opts,
	)
	if err != nil {
		return nil, err
	}

	results := []entity.User{}

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	fmt.Printf("%#v", results[0])
	return results, nil
}

func (u *UserMongo) List(users []entity.ID) ([]entity.User, error) {
	cursor, err := u.collection.Find(context.TODO(),
		bson.D{{Key: "id", Value: bson.D{{Key: "$in", Value: users}}}},
	)
	if err != nil {
		return nil, err
	}

	var results []entity.User

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

//Confirm if the parnter actually requested to be a couple
func (u *UserMongo) ConfirmCouple(userID, partnerID entity.ID) bool {
	err := u.collection.FindOne(
		context.TODO(),
		bson.D{
			{Key: "_id", Value: partnerID},
			{Key: "parnter_id", Value: userID},
		},
	)
	return err == nil
}

func (u *UserMongo) Following(userName string, skip int) ([]entity.Following, error) {
	var following []entity.Following

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "couple_name", Value: userName}}}}
	skipLimitStage := bson.D{{Key: "$skip", Value: int64(skip)}, {Key: "$limit", Value: entity.Limit + 1}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$following"}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "couples"},
				{Key: "localfield", Value: "following"},
				{Key: "foreignfield", Value: "_id"},
				{Key: "as", Value: "user_following"},
			},
		},
	}
	unwindStage2 := bson.D{{Key: "$unwind", Value: "$user_following"}}
	cursor, err := u.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, unwindStage, skipLimitStage, joinStage, unwindStage2},
	)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &following); err != nil {
		return nil, err
	}
	return following, nil
}

func (u *UserMongo) Request(from, to entity.ID) error {
	result, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: "from"}},
		bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "has_pending_request", Value: true},
					{Key: "partner_id", Value: to},
				},
			},
		},
	)
	if result.ModifiedCount != 1 {
		return errors.New("something went wrong")
	}
	return err
}

func (u *UserMongo) Notify(userName string, notif any) error {
	result, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "user_name", Value: userName}},
		bson.D{{Key: "$push", Value: bson.D{{Key: "notifications", Value: notif}}}},
	)
	if result.ModifiedCount != 1 {
		return errors.New("notify: couldn't update user notifications")
	}
	return err
}

//Write Methods
func (u *UserMongo) NotifyCouple(c [2]entity.ID, notif any) error {
	result, err := u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: c}}}},
		bson.D{{Key: "$push", Value: bson.D{{Key: "notifications", Value: notif}}}},
	)
	if result.ModifiedCount < 2 {
		return errors.New("notify: couldn't update user notifications")
	}
	return err
}

func (u *UserMongo) NotifyUsers(users []string, notif any) error {
	_, err := u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "user_name", Value: bson.D{{Key: "$in", Value: users}}}},
		bson.D{{Key: "$push", Value: bson.D{{Key: "notifications", Value: notif}}}},
	)

	return err
}

func (u *UserMongo) Create(e *entity.User) (entity.ID, error) {

	result, err := u.collection.InsertOne(context.TODO(), e)
	if err != nil {
		return primitive.NewObjectID(), err
	}
	return result.InsertedID.(entity.ID), nil
}

func (u *UserMongo) Update(userID entity.ID, update entity.UpdateUser) error {

	_, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: userID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "first_name", Value: update.FirstName},
			{Key: "last_name", Value: update.LastName},
			{Key: "bio", Value: update.Bio},
			{Key: "updated_at", Value: update.UpdatedAt},
			{Key: "pronouns", Value: update.Pronouns},
			{Key: "website", Value: update.Website},
		}}},
	)
	return err
}

func (u *UserMongo) Follow(coupleID entity.ID, userID entity.ID) error {
	result, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{
			{Key: "$inc", Value: bson.D{{Key: "following_count", Value: 1}}},
			{Key: "$push", Value: bson.D{{Key: "following", Value: coupleID}}},
		},
	)
	if result.ModifiedCount < 1 {
		return errors.New("user follow: something went wrong")
	}
	return err
}

func (u *UserMongo) Unfollow(coupleID, userID entity.ID) error {
	result, err := u.collection.UpdateByID(
		context.TODO(),
		coupleID,
		bson.D{
			{Key: "$inc", Value: bson.D{{Key: "following_count", Value: -1}}},
			{Key: "$pull", Value: bson.D{{Key: "following", Value: coupleID}}},
		},
	)
	if result.MatchedCount < 1 {
		return errors.New("unfollow: no match found")
	}
	return err
}

func (u *UserMongo) Delete(id entity.ID) error {
	_, err := u.collection.DeleteOne(
		context.TODO(),
		bson.D{{Key: "id", Value: id}},
	)
	return err
}

func (u *UserMongo) NewCouple(c [2]entity.ID, coupleId entity.ID) error {
	_, err := u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: c}}}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "couple_id", Value: coupleId}}}},
	)

	return err
}

func (u *UserMongo) UpdateProfilePic(fileName string, userID entity.ID) error {
	result, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{{Key: "$set", Value: bson.D{{Key: "profile_picture", Value: fileName}}}},
	)
	if result.MatchedCount == 0 {
		return errors.New("no match")
	}
	return err
}

func (u *UserMongo) UpdateShowPicture(userID entity.ID, index int, fileName string) error {
	result, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{
			{
				Key: "$push",
				Value: bson.D{{Key: "show_pictures", Value: bson.D{
					{Key: "$each", Value: []string{fileName}},
					{Key: "$position", Value: index},
				}}},
			},
		},
	)
	if result.MatchedCount < 1 {
		return entity.ErrNoMatch
	}

	return err
}

func (u *UserMongo) ChangeRequestStatus(userID entity.ID, status string) error {
	_, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{{Key: "open_to_request", Value: status == "ON"}},
	)
	return err
}

func (u *UserMongo) ChangeName(userID entity.ID, userName string) error {
	_, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{{Key: "$set", Value: bson.D{{Key: "user_name", Value: userName}}}},
	)
	return err
}

func (u *UserMongo) ChangeSettings(userID entity.ID, setting, value string) error {
	result, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.D{{Key: setting, Value: value}},
	)
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return err
}
