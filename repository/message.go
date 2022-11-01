package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserCoupleMessage struct {
	col *mongo.Collection
}

func (m *UserCoupleMessage) Get(id entity.ID, skip int) ([]entity.UserCoupleMessage, error) {
	curr, err := m.col.Find(context.TODO(), bson.D{{Key: "$or", Value: bson.D{{Key: "from", Value: id}, {Key: "to", Value: id}}}})
	if err != nil {
		return nil, err
	}
	res := []entity.UserCoupleMessage{}

	if err = curr.All(context.TODO(), &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (m *UserCoupleMessage) GetToCouple(userID, coupleID entity.ID, skip int) ([]entity.UserCoupleMessage, error) {
	curr, err := m.col.Find(
		context.TODO(),
		bson.D{{
			Key: "$or",
			Value: bson.D{
				{Key: "$and", Value: bson.D{
					{Key: "from", Value: userID},
					{Key: "to", Value: coupleID},
				}},
				{Key: "$and", Value: bson.D{
					{Key: "from", Value: coupleID},
					{Key: "to", Value: userID},
				}},
			},
		}},
	)
	if err != nil {
		return nil, err
	}
	res := []entity.UserCoupleMessage{}

	if err = curr.All(context.TODO(), &res); err != nil {
		return nil, err
	}
	return res, nil
}

type CoupleMessage struct {
	col *mongo.Collection
}

func (m *CoupleMessage) Get(coupleID entity.ID, skip int) ([]entity.CoupleMessage, error) {
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetSkip(int64(skip)).SetLimit(entity.Limit)
	curr, err := m.col.Find(context.TODO(), bson.D{{Key: "couple_id", Value: coupleID.Hex()}}, opts)
	if err != nil {
		return nil, err
	}
	res := []entity.CoupleMessage{}

	if err = curr.All(context.TODO(), &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (m *CoupleMessage) NewMessages(userID, coupleID entity.ID) (int64, error) {
	return m.col.CountDocuments(context.TODO(), bson.D{{Key: "couple_id", Value: coupleID.Hex()}, {Key: "from", Value: bson.M{"$ne": userID.Hex()}}, {Key: "recieved", Value: false}})
}

func NewUserCoupleMessageRepo(col *mongo.Collection) UserCoupleMessage {
	return UserCoupleMessage{
		col: col,
	}
}

func NewCoupleMessageRepo(col *mongo.Collection) CoupleMessage {
	return CoupleMessage{
		col: col,
	}
}
