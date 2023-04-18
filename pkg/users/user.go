package users

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (u *Users) GetUser(ctx context.Context, id string) (User, error) {

	var user User
	err := u.FindOne(ctx, bson.D{{"_id", id}}).Decode(&user)

	return user, err

}

func (u *Users) DeleteUser(ctx context.Context, id string) (bool, error) {

	res, err := u.DeleteOne(ctx, bson.D{{"_id", id}})

	return res.DeletedCount == 0, err

}
