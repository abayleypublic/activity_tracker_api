package challenges

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Membership struct {
	Challenge service.ID `json:"challenge" bson:"challenge"`
	User      service.ID `json:"user" bson:"user"`
	Created   *time.Time `json:"created,omitempty" bson:"created,omitempty"`
}

// Memberships wraps a MongoDB collection of challenges.
type Memberships struct {
	*mongo.Collection
}

// NewMemberships creates a new Memberships instance with the provided MongoDB collection.
func NewMemberships(c *mongo.Collection) *Memberships {
	return &Memberships{c}
}

func (svc *Memberships) Setup(ctx context.Context) error {
	if err := svc.Database().CreateCollection(ctx, svc.Name()); err != nil {
		return fmt.Errorf("failed to create challenge members collection: %w", err)
	}

	_, err := svc.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "challenge", Value: 1}, {Key: "user", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("challenge_user_unique_index"),
		},
	})

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create unique index for challenge members")
	}

	return nil
}

// Create adds a new membership to the database.
func (svc *Memberships) Create(ctx context.Context, membership *Membership) error {
	_, err := svc.InsertOne(ctx, membership)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

// Get retrieves a membership by its ID from the database.
func (svc *Memberships) Get(ctx context.Context, id service.ID, membership interface{}) error {
	if err := svc.
		FindOne(ctx, bson.D{{Key: "_id", Value: id.ConvertID()}}).
		Decode(membership); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return ErrNotFound
		}
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

type MembershipListOptions struct {
	Limit int64
	Skip  int64

	User      *service.ID
	Challenge *service.ID
}

func NewMembershipListOptions() MembershipListOptions {
	return MembershipListOptions{}
}

func (opts *MembershipListOptions) SetLimit(limit int64) *MembershipListOptions {
	opts.Limit = limit
	return opts
}

func (opts *MembershipListOptions) SetSkip(skip int64) *MembershipListOptions {
	opts.Skip = skip
	return opts
}

func (opts *MembershipListOptions) SetUser(id service.ID) *MembershipListOptions {
	opts.User = &id
	return opts
}

// List retrieves memberships based on the given criteria.
func (svc *Memberships) List(ctx context.Context, opts MembershipListOptions, memberships interface{}) error {
	options := options.Find()

	if opts.Limit > 0 {
		options = options.SetLimit(opts.Limit)
	}

	if opts.Skip > 0 {
		options = options.SetSkip(opts.Skip)
	}

	filter := bson.D{}
	if opts.User != nil {
		filter = append(filter, bson.E{Key: "user", Value: opts.User.ConvertID()})
	}
	if opts.Challenge != nil {
		filter = append(filter, bson.E{Key: "challenge", Value: opts.Challenge.ConvertID()})
	}

	cursor, err := svc.Find(ctx, filter, options)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if err := cursor.All(ctx, memberships); err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

type MembershipDeleteOpts struct {
	Challenge *service.ID
	User      *service.ID
}

// Delete removes memberships based on the provided criteria.
// This can be used to delete memberships for a specific user or a whole challenge.
func (svc *Memberships) Delete(ctx context.Context, opts MembershipDeleteOpts) error {
	filter := bson.D{}
	if opts.User != nil {
		filter = append(filter, bson.E{Key: "user", Value: opts.User.ConvertID()})
	}
	if opts.Challenge != nil {
		filter = append(filter, bson.E{Key: "challenge", Value: opts.Challenge.ConvertID()})
	}

	res, err := svc.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if res.DeletedCount >= 1 {
		return ErrNotFound
	}

	return nil
}
