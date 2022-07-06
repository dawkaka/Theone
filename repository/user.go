package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
func (u *UserMongo) Get(userName string) (*entity.User, error) {
	var result entity.User

	err := u.collection.FindOne(context.TODO(), bson.D{{Key: "user_name", Value: userName}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (u *UserMongo) Search(query string) ([]*entity.User, error) {

	cursor, err := u.collection.Find(
		context.TODO(),
		bson.D{
			{Key: "$or", Value: bson.D{
				{Key: "user_name", Value: bson.D{{Key: "$regex", Value: "/^" + query + "/i"}}},
				{Key: "first_name", Value: bson.D{{Key: "$regex", Value: "/^" + query + "/i"}}},
				{Key: "last_name", Value: bson.D{{Key: "$regex", Value: "/^" + query + "/i"}}},
			}}},
	)
	if err != nil {
		return nil, err
	}

	var results []*entity.User

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (u *UserMongo) List(users []entity.ID) ([]*entity.User, error) {
	cursor, err := u.collection.Find(context.TODO(),
		bson.D{{Key: "id", Value: bson.D{{Key: "$in", Value: users}}}},
	)
	if err != nil {
		return nil, err
	}

	var results []*entity.User

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

//Write Methods
func (u *UserMongo) Create(e *entity.User) error {

	_, err := u.collection.InsertOne(context.TODO(), e)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserMongo) Update(e *entity.User) error {

	_, err := u.collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "user_name", Value: e.UserName}},
		bson.D{{Key: "$set", Value: e}},
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
