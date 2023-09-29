package users

import (
	"context"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReadUser retrieves a user from the database using the provided ID.
// It first converts the ID to an ObjectID, then attempts to find a user with that ID in the database.
// If the user is found, it is returned. If not, an error is returned.
func (u *Users) ReadUser(ctx context.Context, id uuid.ID) (*User, error) {

	user := &User{}
	if err := u.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(user); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

// CreateUser inserts a new user into the database.
// It takes a User object as input and inserts it into the database.
// If the operation is successful, it returns nil. If not, it returns the error.
func (u *Users) CreateUser(ctx context.Context, user User) error {
	opts := options.Update().SetUpsert(true)
	user.CreatedDate = time.Now().UTC().Format(time.UnixDate)
	res, err := u.UpdateOne(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$exists", Value: false}}}}, bson.D{{Key: "$set", Value: user}}, opts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return err

	}

	if res.UpsertedCount == 0 {
		return ErrUserAlreadyExists
	}

	return err
}

// UpdateUser upserts a user into the database.
// It takes a User object as input and inserts it into the database.
// If the operation is successful, it returns nil. If not, it returns the error.
func (u *Users) UpdateUser(ctx context.Context, user User) error {
	opts := options.Update().SetUpsert(true)
	res, err := u.UpdateOne(ctx, bson.D{{Key: "$set", Value: user}}, opts)
	if res.UpsertedCount != 1 {
		return ErrUserNotFound
	}

	return err
}

// DeleteUser removes a user from the database using the provided ID.
// It first converts the ID to an ObjectID, then attempts to delete a user with that ID from the database.
func (u *Users) DeleteUser(ctx context.Context, id uuid.ID) error {

	res, err := u.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if res.DeletedCount != 1 {
		return ErrUserNotFound
	}

	return err

}
