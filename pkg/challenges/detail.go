package challenges

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Detail represents a full challenge, including its members.
type Detail struct {
	ID          service.ID     `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string         `json:"name" bson:"name"`
	Description string         `json:"description" bson:"description"`
	StartDate   time.Time      `json:"startDate" bson:"startDate"`
	EndDate     time.Time      `json:"endDate" bson:"endDate"`
	Public      bool           `json:"public" bson:"public"`
	InviteOnly  bool           `json:"inviteOnly" bson:"inviteOnly"`
	CreatedBy   service.ID     `json:"createdBy" bson:"createdBy"`
	CreatedDate *time.Time     `json:"createdDate" bson:"createdDate"`
	Target      targets.Target `json:"target" bson:"target"`
}

// Details wraps a MongoDB collection of challenge details.
type Details struct {
	*mongo.Collection
}

// NewDetails creates a new Challenges instance with the provided MongoDB collection.
func NewDetails(c *mongo.Collection) *Details {
	return &Details{c}
}

func (svc *Details) Setup(ctx context.Context) error {
	if err := svc.Database().CreateCollection(ctx, svc.Name()); err != nil {
		return fmt.Errorf("failed to create challenge detail collection: %w", err)
	}
	return nil
}

// Create adds a new challenge to the database.
func (svc *Details) Create(ctx context.Context, challenge *Detail) error {
	res, err := svc.InsertOne(ctx, challenge)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	challenge.ID = service.ID(res.InsertedID.(string))

	return nil
}

// Get retrieves a challenge by its ID from the database.
func (svc *Details) Get(ctx context.Context, id service.ID, challenge interface{}) error {
	if err := svc.
		FindOne(ctx, bson.D{{Key: "_id", Value: id.ConvertID()}}).
		Decode(challenge); err != nil {
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
}

func NewDetailListOptions() DetailListOptions {
	return DetailListOptions{}
}

func (opts *DetailListOptions) SetLimit(limit int64) *DetailListOptions {
	opts.Limit = limit
	return opts
}

func (opts *DetailListOptions) SetSkip(skip int64) *DetailListOptions {
	opts.Skip = skip
	return opts
}

// List retrieves challenges based on the given criteria.
func (svc *Details) List(ctx context.Context, opts DetailListOptions, challenges interface{}) error {
	options := options.Find()

	if opts.Limit > 0 {
		options = options.SetLimit(opts.Limit)
	}

	if opts.Skip > 0 {
		options = options.SetSkip(opts.Skip)
	}

	cursor, err := svc.Find(ctx, bson.D{}, options)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if err := cursor.All(ctx, challenges); err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

// Update updates a challenge in the database based on the provided criteria.
func (svc *Details) Update(ctx context.Context, challenge Detail) error {
	opts := options.UpdateOne().SetUpsert(true)
	res, err := svc.UpdateOne(
		ctx,
		bson.D{},
		bson.D{{Key: "$set", Value: challenge}},
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

// Delete removes a challenge from the database by its ID.
func (svc *Details) Delete(ctx context.Context, challengeID service.ID) error {
	res, err := svc.DeleteOne(ctx, bson.D{{Key: "_id", Value: challengeID}})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if res.DeletedCount != 1 {
		return ErrNotFound
	}

	return nil
}
