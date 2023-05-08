package users

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (u *Users) GetUser(ctx context.Context, id string) (User, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return User{}, err
	}

	var user User
	if err := u.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil

}

func (u *Users) PutUser(ctx context.Context, user User) error {

	_, err := u.InsertOne(ctx, user)

	return err

}

func (u *Users) DeleteUser(ctx context.Context, id string) (bool, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return false, err
	}

	res, err := u.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})

	if err != nil {
		return false, err
	}

	return res.DeletedCount == 0, nil

}
