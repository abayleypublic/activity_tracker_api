package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Detail represents the detailed information of a user.
type Detail struct {
	ID          service.ID `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName   string     `json:"firstName,omitempty" bson:"firstName"`
	LastName    string     `json:"lastName,omitempty" bson:"lastName"`
	Email       string     `json:"email,omitempty" bson:"email"`
	CreatedDate *time.Time `json:"createdDate" bson:"createdDate"`
	Bio         string     `json:"bio,omitempty" bson:"bio"`
}

// Users is a wrapper around a MongoDB collection of users.
type Details struct {
	*mongo.Collection
}

// NewUsers creates a new Users instance with the provided MongoDB collection.
func NewDetails(c *mongo.Collection) *Details {
	return &Details{c}
}

func (svc *Details) Setup(ctx context.Context) error {
	if err := svc.Database().CreateCollection(ctx, svc.Name()); err != nil {
		return fmt.Errorf("failed to create user detail collection: %w", err)
	}
	return nil
}

// Create adds a new user to the database.
func (svc *Details) Create(ctx context.Context, user *Detail) (service.ID, error) {
	user.ID = service.NewID()
	now := time.Now()
	user.CreatedDate = &now
	res, err := svc.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrAlreadyExists
		}
		return "", fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return service.ID(res.InsertedID.(string)), nil
}

// Get retrieves a user by its ID from the database.
func (svc *Details) Get(ctx context.Context, id service.ID, user interface{}) error {
	if err := svc.
		FindOne(ctx, bson.D{{Key: "_id", Value: id.ConvertID()}}).
		Decode(user); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return ErrNotFound
		}
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

type DetailListOptions struct {
	Limit int64
	Skip  int64

	Email string
}

func NewDetailListOptions() *DetailListOptions {
	return &DetailListOptions{}
}

func (opts *DetailListOptions) SetLimit(limit int64) *DetailListOptions {
	opts.Limit = limit
	return opts
}

func (opts *DetailListOptions) SetSkip(skip int64) *DetailListOptions {
	opts.Skip = skip
	return opts
}

func (opts *DetailListOptions) SetEmail(email string) *DetailListOptions {
	opts.Email = email
	return opts
}

// List retrieves users based on the given criteria.
func (svc *Details) List(ctx context.Context, opts DetailListOptions, users interface{}) error {
	options := options.Find()

	if opts.Limit > 0 {
		options = options.SetLimit(opts.Limit)
	}

	if opts.Skip > 0 {
		options = options.SetSkip(opts.Skip)
	}

	filter := bson.D{}
	if opts.Email != "" {
		filter = append(filter, bson.E{Key: "email", Value: opts.Email})
	}

	cursor, err := svc.Find(ctx, filter, options)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if err := cursor.All(ctx, users); err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

// Update updates a user in the database based on the provided criteria.
func (svc *Details) Update(ctx context.Context, user Detail) error {
	opts := options.UpdateOne().SetUpsert(true)
	res, err := svc.UpdateOne(
		ctx,
		bson.D{},
		bson.D{{Key: "$set", Value: user}},
		opts,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if res.UpsertedCount != 1 {
		return ErrNotFound
	}

	return nil
}

// Delete removes a user from the database by its ID.
func (svc *Details) Delete(ctx context.Context, userID service.ID) error {
	res, err := svc.DeleteOne(ctx, bson.D{{Key: "_id", Value: userID}})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if res.DeletedCount != 1 {
		return ErrNotFound
	}

	return nil
}
