package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
func (u *UserMongo) ListFollowers(flws []entity.ID) ([]entity.Follower, error) {
	opts := options.Find().SetProjection(
		bson.D{
			{Key: "first_name", Value: 1}, {Key: "last_name", Value: 1},
			{Key: "user_name", Value: 1}, {Key: "profile_picture", Value: 1},
			{Key: "has_partner", Value: 1},
		})
	cursor, err := u.collection.Find(context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: flws}}}},
		opts,
	)
	if err != nil {
		return nil, err
	}
	results := []entity.Follower{}
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
			{Key: "following", Value: 0}, {Key: "likes", Value: 0}, {Key: "feed_posts", Value: 0},
			{Key: "previous_relationships", Value: 0}, {Key: "exempted", Value: 0},
		})

	user := entity.User{}
	err := u.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "$or", Value: bson.A{bson.D{{Key: "user_name", Value: param}}, bson.D{{Key: "email", Value: param}}}}},
		opts,
	).Decode(&user)
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

	if err != nil {
		return user, entity.ErrSomethingWentWrong
	}
	return user, err
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

func (u *UserMongo) Notifications(userName string, page int) (presentation.Notification, error) {
	opts := options.FindOne().SetProjection(bson.M{"notifications": bson.M{"$slice": []int{page, entity.Limit}}})

	user := entity.User{}
	notif := presentation.Notification{Notifications: []entity.Notification{}, NewCount: 0}

	err := u.collection.FindOne(
		context.TODO(),
		bson.M{"user_name": userName},
		opts,
	).Decode(&user)

	if err != nil {
		return notif, err
	}

	notif.Notifications = user.Notifications
	notif.NewCount = user.NewNotificationsCount

	return notif, nil
}

//Write Methods
func (u *UserMongo) Notify(userName string, notif entity.Notification) error {
	// matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "user_name", Value: userName}}}}
	// projectStage := bson.D{{
	// 	Key: "$project",
	// 	Value: bson.M{
	// 		"notifications": bson.M{"$filter": bson.M{
	// 			"input": "$notifications",
	// 			"as":    "notif",
	// 			"cond":  bson.M{"$eq": []interface{}{"$$notif.user", notif.User}},
	// 		}},
	// 	},
	// }}
	// cursor, err := u.collection.Aggregate(context.TODO(), mongo.Pipeline{matchStage, projectStage})

	// if err != nil {
	// 	return entity.ErrNoMatch
	// }

	// if err != nil {
	// 	return entity.ErrNoMatch
	// }

	// notifs := []entity.User{}
	// if err = cursor.All(context.TODO(), &notifs); err != nil {
	// 	return err
	// }
	// if len(notifs) != 0 && len(notifs[0].Notifications) != 0 {
	// 	for _, val := range notifs[0].Notifications {
	// 		switch val.Type {
	// 		case "Mentioned":
	// 			fallthrough
	// 		case "like":
	// 			if val.PostID == notif.PostID {
	// 				return nil
	// 			}
	// 		case "comment":
	// 			if val.PostID == notif.PostID && val.Message == notif.Message {
	// 				return nil
	// 			}
	// 		case "follow":
	// 			return nil
	// 		}
	// 	}
	// }

	_, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "user_name", Value: userName}},
		bson.A{
			bson.D{
				{
					Key: "$set", Value: bson.D{
						{
							Key: "notifications",
							Value: bson.M{
								"$concatArrays": bson.A{
									bson.A{notif},
									bson.M{
										"$slice": bson.A{"$notifications", 0, 98},
									},
								},
							},
						},
						{
							Key:   "new_notifications_count",
							Value: bson.M{"$add": bson.A{"$new_notifications_count", 1}},
						},
					},
				},
			},
		},
	)

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

func (u *UserMongo) NotifyCouple(c [2]entity.ID, notif entity.Notification) error {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.M{"$in": c}}}}}
	projectStage := bson.D{{
		Key: "$project",
		Value: bson.M{
			"notifications": bson.M{"$filter": bson.M{
				"input": "$notifications",
				"as":    "notif",
				"cond":  bson.M{"$eq": []interface{}{"$$notif.user", notif.User}},
			}},
		},
	}}
	cursor, err := u.collection.Aggregate(context.TODO(), mongo.Pipeline{matchStage, projectStage})

	if err != nil {
		fmt.Println(err)
		return entity.ErrNoMatch
	}

	notifs := []entity.User{}
	if err = cursor.All(context.TODO(), &notifs); err != nil {
		fmt.Println(err)
		return err
	}
	if len(notifs) != 0 && len(notifs[0].Notifications) != 0 {
		for _, val := range notifs[0].Notifications {
			switch val.Type {
			case "Mentioned":
				fallthrough
			case "like":
				if val.PostID == notif.PostID && notif.Type == "like" || notif.Type == "Mentioned" {
					return nil
				}
			case "comment":
				if val.PostID == notif.PostID && val.Message == notif.Message {
					return nil
				}
			case "follow":
				if notif.Type == "follow" {
					return nil
				}
			default:
				continue
			}
		}
	}

	_, err = u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: c}}}},
		bson.A{
			bson.D{
				{
					Key: "$set", Value: bson.D{
						{
							Key: "notifications",
							Value: bson.M{
								"$concatArrays": bson.A{
									bson.A{notif},
									bson.M{
										"$slice": bson.A{"$notifications", 0, 98},
									},
								},
							},
						},
						{
							Key:   "new_notifications_count",
							Value: bson.M{"$add": bson.A{"$new_notifications_count", 1}},
						},
					},
				},
			},
		},
	)
	return err
}

func (u *UserMongo) NotifyUsers(users []string, notif entity.Notification) error {
	_, err := u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "user_name", Value: bson.D{{Key: "$in", Value: users}}}},
		bson.A{
			bson.D{
				{
					Key: "$set", Value: bson.D{
						{
							Key: "notifications",
							Value: bson.M{
								"$concatArrays": bson.A{
									bson.A{notif},
									bson.M{
										"$slice": bson.A{"$notifications", 0, 98},
									},
								},
							},
						},
						{
							Key:   "new_notifications_count",
							Value: bson.M{"$add": bson.A{"$new_notifications_count", 1}},
						},
					},
				},
			},
		},
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
	_, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.A{
			bson.D{{
				Key: "$set", Value: bson.M{"following": bson.M{"$setUnion": []interface{}{"$following", []entity.ID{coupleID}}}},
			}},
			bson.D{{
				Key: "$set", Value: bson.M{"following_count": bson.M{"$size": "$following"}},
			}},
		},
	)
	return err
}

func (u *UserMongo) Unfollow(coupleID, userID entity.ID) error {
	_, err := u.collection.UpdateByID(
		context.TODO(),
		userID,
		bson.A{
			bson.D{{
				Key: "$set", Value: bson.M{"following": bson.M{"$setDifference": []interface{}{"$following", []entity.ID{coupleID}}}},
			}},
			bson.D{{
				Key: "$set", Value: bson.M{"following_count": bson.M{"$size": "$following"}},
			}},
		},
	)
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
		bson.M{"_id": bson.M{"$in": []entity.ID{couple[0], couple[1]}}},
		bson.D{
			{Key: "$set", Value: bson.M{"has_partner": false, "open_to_request": true}},
			{Key: "$addToSet", Value: bson.M{"previous_relationships": bson.M{"$each": []entity.ID{couple[0], couple[1]}}}},
			{Key: "$unset", Value: bson.M{"partner_id": 1, "couple_id": 1}},
		},
	)
	return err
}

func (u *UserMongo) Startup(userID entity.ID) (presentation.StartupInfo, error) {
	result := presentation.StartupInfo{}
	opts := options.FindOne().SetProjection(bson.M{"has_partner": 1, "new_notifications_count": 1, "user_name": 1, "new_feed_post_count": 1, "pending_request": 1})
	err := u.collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: userID}}, opts).Decode(&result)
	return result, err
}

func (u *UserMongo) ClearNotifsCount(userID entity.ID) error {
	_, err := u.collection.UpdateByID(context.TODO(), userID, bson.D{{Key: "$set", Value: bson.M{"new_notifications_count": 0}}})
	return err
}
func (u *UserMongo) ClearFeedPostsCount(userID entity.ID) error {
	_, err := u.collection.UpdateByID(context.TODO(), userID, bson.D{{Key: "$set", Value: bson.M{"new_feed_post_count": 0}}})

	return err
}

func (u *UserMongo) UsageMonitoring(userID entity.ID) error {
	_, err := u.collection.UpdateByID(context.TODO(), userID, bson.D{{Key: "$set", Value: bson.M{"last_visited": time.Now()}}})
	return err
}

func (u *UserMongo) NewFeedPost(postID entity.ID, userIDs []entity.ID) error {
	_, err := u.collection.UpdateMany(
		context.TODO(),
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: userIDs}}}},
		bson.A{
			bson.D{
				{
					Key: "$set", Value: bson.D{
						{
							Key: "feed_posts",
							Value: bson.M{
								"$concatArrays": bson.A{
									bson.A{postID},
									bson.M{
										"$slice": bson.A{"$feed_posts", 0, 1000},
									},
								},
							},
						},
						{
							Key:   "new_feed_post_count",
							Value: bson.M{"$add": bson.A{"$new_feed_post_count", 1}},
						},
					},
				},
			},
		},
	)
	return err
}

func (u *UserMongo) GetFeedPosts(userID entity.ID, skip int) ([]presentation.Post, error) {
	result := []presentation.Post{}
	matchStage := bson.D{{Key: "$match", Value: bson.M{"_id": userID}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$feed_posts"}}
	pro := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"feed_posts": 1,
			},
		},
	}
	skipStage := bson.D{{Key: "$skip", Value: skip}}
	limitStage := bson.D{{Key: "$limit", Value: entity.Limit}}
	joinStage := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "posts"},
				{Key: "localField", Value: "feed_posts"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "post"},
			},
		},
	}
	unwindStage2 := bson.D{{Key: "$unwind", Value: "$post"}}
	joinStage2 := bson.D{
		{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "couples"},
				{Key: "localField", Value: "post.couple_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "couple"},
			},
		},
	}
	unwindStage3 := bson.D{{Key: "$unwind", Value: "$couple"}}
	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"_id":             "$post._id",
				"post_id":         "$post.post_id",
				"couple_id":       "$post.couple_id",
				"files":           "$post.files",
				"caption":         "$post.caption",
				"location":        "$post.location",
				"likes_count":     "$post.likes_count",
				"comments_count":  "$post.comments_count",
				"created_at":      "$post.created_at",
				"couple_name":     "$couple.couple_name",
				"profile_picture": "$couple.profile_picture",
				"comments_closed": "$post.comments_closed",
				"verified":        "$couple.verified",
				"married":         "$couple.married",
				"has_liked":       bson.M{"$in": bson.A{userID, "$post.likes"}},
			},
		},
	}

	cursor, err := u.collection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, pro, unwindStage, skipStage, joinStage, unwindStage2, limitStage, joinStage2, unwindStage3, projectStage},
	)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &result)
	return result, err
}

func (u *UserMongo) CheckNameAvailability(name string) bool {
	opts := options.FindOne().SetProjection(bson.D{{Key: "user_name", Value: 1}})
	err := u.collection.FindOne(
		context.TODO(),
		bson.D{{Key: "user_name", Value: bson.M{"$regex": primitive.Regex{Pattern: "^" + name + "$", Options: "i"}}}},
		opts,
	)
	return err.Err() != nil
}

func (u *UserMongo) ExemptedFromSuggestedAccounts(userID entity.ID, addExempt bool) ([]entity.ID, error) {
	opts := options.FindOne().SetProjection(bson.D{{Key: "following", Value: 1}, {Key: "exempted", Value: 1}})
	user := entity.User{}
	err := u.collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: userID}}, opts).Decode(&user)
	if err != nil {
		return nil, err
	}
	coupleIDs := user.Following
	if addExempt {
		coupleIDs = append(coupleIDs, user.Exempted...)
	}
	return coupleIDs, nil
}

func (u *UserMongo) Exempt(userID, coupleID entity.ID) error {
	_, err := u.collection.UpdateByID(context.TODO(), userID, bson.D{{Key: "$push", Value: bson.M{"exempted": coupleID}}})
	return err
}

func (u *UserMongo) ResetPassword(email, password string) error {
	_, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "email", Value: email}},
		bson.D{{Key: "$set", Value: bson.M{"password": password}}},
	)
	return err
}
