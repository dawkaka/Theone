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
			{Key: "following", Value: 0}, {Key: "likes", Value: 0},
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

func (u *UserMongo) Search(query string) ([]presentation.UserPreview, error) {
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

	results := []presentation.UserPreview{}

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (u *UserMongo) List(users []entity.ID) ([]presentation.UserPreview, error) {
	cursor, err := u.collection.Find(context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: users}}}},
	)
	if err != nil {
		return nil, err
	}

	results := []presentation.UserPreview{}

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

func (u *UserMongo) Login(param string) (entity.User, error) {
	opts := options.FindOne().SetProjection(
		bson.D{
			{Key: "following", Value: 0}, {Key: "likes", Value: 0},
		})

	user := entity.User{}
	err := u.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "$or", Value: bson.A{bson.D{{Key: "user_name", Value: param}}, bson.D{{Key: "email", Value: param}}}}},
		opts,
	).Decode(&user)
	fmt.Println(err)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, entity.ErrUserNotFound
		}
		return user, entity.ErrSomethingWentWrong
	}
	return user, nil
}

func (u *UserMongo) CheckSignup(userName, email string) (entity.User, error) {
	opts := options.FindOne().SetProjection(
		bson.D{
			{Key: "user_name", Value: 1}, {Key: "email", Value: 1},
		})

	user := entity.User{}
	err := u.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "$or", Value: bson.A{bson.M{"user_name": userName}, bson.M{"email": email}}}},
		opts,
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return user, entity.ErrUserNotFound
	}

	return user, entity.ErrSomethingWentWrong

}

func (u *UserMongo) Following(userName string, skip int) ([]entity.ID, error) {
	result := entity.User{}
	opts := options.FindOne().SetProjection(bson.M{"following": bson.M{"$slice": []int{skip, entity.Limit}}})
	err := u.collection.FindOne(
		context.TODO(),
		bson.M{"user_name": userName},
		opts,
	).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Following, nil
}

func (u *UserMongo) Notifications(userName string, page int) ([]entity.Notification, error) {
	opts := options.FindOne().SetProjection(bson.M{"notifications": bson.M{"$slice": []int{page, entity.Limit}}})

	user := entity.User{}

	err := u.collection.FindOne(
		context.TODO(),
		bson.M{"user_name": userName},
		opts,
	).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.Notifications, nil
}

//Write Methods
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
func (u *UserMongo) SendRequest(from, to entity.ID) error {
	_, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: from}},
		bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "pending_request", Value: entity.SENT_REQUEST},
					{Key: "partner_id", Value: to},
				},
			},
		},
	)
	return err
}

func (u *UserMongo) RecieveRequest(from, to entity.ID) error {
	_, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: from}},
		bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "pending_request", Value: entity.RECIEVED_REQUEST},
					{Key: "partner_id", Value: to},
				},
			},
		},
	)
	return err
}

func (u *UserMongo) NullifyRequest(userIDs [2]entity.ID) error {
	_, err := u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: userIDs}}}},
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "pending_request", Value: entity.NO_REQUEST}}},
			{Key: "$unset", Value: bson.D{{Key: "partner_id", Value: 1}}},
		},
	)

	return err
}

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
			{Key: "date_of_birth", Value: update.DateOfBirth},
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
		userID,
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
		bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "couple_id", Value: coupleId},
					{Key: "has_partner", Value: true},
					{Key: "open_to_request", Value: false},
					{Key: "pending_request", Value: entity.NO_REQUEST},
				},
			},
		},
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
				Key:   "$set",
				Value: bson.D{{Key: "show_pictures." + fmt.Sprint(index), Value: fileName}},
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
		bson.D{{Key: "$set", Value: bson.D{{Key: "open_to_request", Value: status == "ON"}}}},
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
		bson.D{{Key: "$set", Value: bson.D{{Key: setting, Value: value}}}},
	)
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return err
}

func (u *UserMongo) BreakedUp(couple [2]entity.ID) error {
	_, err := u.collection.UpdateMany(
		context.TODO(),
		bson.M{"_id": bson.M{"$in": couple}},
		bson.D{
			{Key: "$set", Value: bson.M{"has_partner": false, "open_to_request": true}},
			{Key: "$addToSet", Value: bson.M{"previous_relationships": bson.M{"$each": couple}}},
			{Key: "$unset", Value: bson.M{"partner_id": 1, "couple_id": 1}},
		},
	)
	return err
}
