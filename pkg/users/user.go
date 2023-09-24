package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReadUser retrieves a user from the database using the provided ID.
// It first converts the ID to an ObjectID, then attempts to find a user with that ID in the database.
// If the user is found, it is returned. If not, an error is returned.
func (u *Users) ReadUser(ctx context.Context, id uuid.ID) (User, error) {

	oid, err := uuid.ConvertID(id)
	if err != nil {
		return User{}, err
	}

	user := User{}
	if err := u.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&user); err != nil {
		return user, err
	}

	return user, nil
}

// CreateOrUpdateUser inserts a new user into the database.
// It takes a User object as input and inserts it into the database.
// If the operation is successful, it returns nil. If not, it returns the error.
func (u *Users) CreateOrUpdateUser(ctx context.Context, user User) error {
	opts := options.Update().SetUpsert(true)
	_, err := u.UpdateOne(ctx, bson.D{{Key: "$set", Value: user}}, opts)
	return err
}

// DeleteUser removes a user from the database using the provided ID.
// It first converts the ID to an ObjectID, then attempts to delete a user with that ID from the database.
// If the operation is successful, it returns true. If not, it returns false and the error.
func (u *Users) DeleteUser(ctx context.Context, id uuid.ID) (bool, error) {

	oid, err := uuid.ConvertID(id)
	if err != nil {
		return false, err
	}

	res, err := u.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		return false, err
	}

	return res.DeletedCount == 1, nil

}
