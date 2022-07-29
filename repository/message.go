package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	curr, err := m.col.Find(context.TODO(), bson.D{{Key: "couple_id", Value: coupleID}})
	if err != nil {
		return nil, err
	}
	res := []entity.CoupleMessage{}

	if err = curr.All(context.TODO(), &res); err != nil {
		return nil, err
	}
	return res, nil
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
