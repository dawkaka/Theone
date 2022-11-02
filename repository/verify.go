package repository

import (
	"context"
	"time"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type VerifyMongo struct {
	col *mongo.Collection
}

func (v *VerifyMongo) NewUser(u entity.VerifySignup) error {
	_, err := v.col.InsertOne(context.TODO(), u)
	return err
}

func (v *VerifyMongo) RequestPasswordReset(email, linkID string) error {
	_, err := v.col.InsertOne(context.TODO(), bson.D{{Key: "id", Value: linkID}, {Key: "email", Value: email}, {Key: "type", Value: "password-reset"}})
	return err
}

func (v *VerifyMongo) GetNewUser(id string) (entity.Signup, error) {
	signup := entity.Signup{}
	sixHoursAgo := time.Now().UnixMilli() - (1000 * 60 * 60 * 6)
	err := v.col.FindOne(context.TODO(), bson.D{{Key: "type", Value: bson.M{"$ne": "password-reset"}}, {Key: "id", Value: id}, {Key: "date", Value: bson.M{"$gt": sixHoursAgo}}}).Decode(&signup)
	return signup, err
}

func (v *VerifyMongo) GetResetPasswordEmail(linkID string) (string, error) {
	res := struct{ Email string }{}
	err := v.col.FindOne(context.TODO(), bson.D{{Key: "id", Value: linkID}, {Key: "type", Value: "password-reset"}}).Decode(&res)
	return res.Email, err
}

func (v *VerifyMongo) Verified(id string) {
	v.col.DeleteOne(context.TODO(), bson.D{{Key: "id", Value: id}})
}

func NewVerifyMongo(collection *mongo.Collection) VerifyMongo {
	return VerifyMongo{
		col: collection,
	}
}
