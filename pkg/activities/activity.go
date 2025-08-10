package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/validate"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrAlreadyExists = errors.New("activity already exists")
	ErrNotFound      = errors.New("activity not found")
	ErrUnknown       = errors.New("unknown error")
	ErrInvalid       = errors.New("invalid")
	ErrValidation    = errors.New("validation error")
)

type ActivityType string

const (
	Any      ActivityType = "any"
	Walking  ActivityType = "walking"
	Running  ActivityType = "running"
	Swimming ActivityType = "swimming"
	Cycling  ActivityType = "cycling"
)

type ActivityCategory map[ActivityType]struct{}

var (
	Moving = ActivityCategory{
		Walking:  {},
		Running:  {},
		Swimming: {},
		Cycling:  {},
	}
)

type Activity struct {
	ID          service.ID   `json:"id" bson:"_id"`
	UserID      service.ID   `json:"user_id" bson:"userID" validate:"required"`
	CreatedDate time.Time    `json:"created_date" bson:"createdDate" validate:"required"`
	Type        ActivityType `json:"type" bson:"type" validate:"required"`
	Value       float64      `json:"value" bson:"value" validate:"required"`
	Start       time.Time    `json:"start" bson:"start" validate:"required"`
	End         time.Time    `json:"end,omitempty" bson:"end" validate:"required,gtfield=Start"`
}

func NewActivity(Type ActivityType, Value float64) Activity {
	return Activity{
		ID:    service.NewID(),
		Type:  Type,
		Value: Value,
	}
}

type Service struct {
	*mongo.Collection
}

func New(c *mongo.Collection) *Service {
	return &Service{c}
}

// Setup initializes the activity service, setting up the underlying database and collections.
func (svc *Service) Setup(ctx context.Context) error {
	if err := svc.Database().CreateCollection(ctx, svc.Name()); err != nil {
		return fmt.Errorf("failed to create activity collection: %w", err)
	}
	return nil
}

// Create adds a new activity to the database.
func (svc *Service) Create(ctx context.Context, activity *Activity) (service.ID, error) {
	activity.ID = service.NewID()
	activity.CreatedDate = time.Now()

	if err := validate.Struct(activity); err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidation, err)
	}

	res, err := svc.InsertOne(ctx, activity)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrAlreadyExists
		}
		return "", fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return service.ID(res.InsertedID.(string)), nil
}

// Get retrieves an activity by its ID from the database.
func (svc *Service) Get(ctx context.Context, id service.ID, activity interface{}) error {
	if err := svc.
		FindOne(ctx, bson.D{{Key: "_id", Value: id.ConvertID()}}).
		Decode(activity); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return ErrNotFound
		}
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

type ListOptions struct {
	Limit int64
	Skip  int64

	User *service.ID
}

func NewListOptions() *ListOptions {
	return &ListOptions{}
}

func (opts *ListOptions) SetLimit(limit int64) *ListOptions {
	opts.Limit = limit
	return opts
}

func (opts *ListOptions) SetSkip(skip int64) *ListOptions {
	opts.Skip = skip
	return opts
}

func (opts *ListOptions) SetUser(id service.ID) *ListOptions {
	opts.User = &id
	return opts
}

// List retrieves activities based on the given criteria.
func (svc *Service) List(ctx context.Context, opts ListOptions, activities interface{}) error {
	options := options.Find()

	if opts.Limit > 0 {
		options = options.SetLimit(opts.Limit)
	}

	if opts.Skip > 0 {
		options = options.SetSkip(opts.Skip)
	}

	filter := bson.D{}
	if opts.User != nil {
		filter = append(filter, bson.E{Key: "userID", Value: opts.User.ConvertID()})
	}

	cursor, err := svc.Find(ctx, filter, options)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if err := cursor.All(ctx, activities); err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}

// Update updates an activity in the database based on the provided criteria.
func (svc *Service) Update(ctx context.Context, activity Activity) error {
	if err := validate.Struct(activity); err != nil {
		return fmt.Errorf("%w: %w", ErrValidation, err)
	}

	opts := options.UpdateOne().SetUpsert(true)
	res, err := svc.UpdateOne(
		ctx,
		bson.D{},
		bson.D{{Key: "$set", Value: activity}},
		opts,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	if res.MatchedCount != 1 {
		return ErrNotFound
	}

	return nil
}

type ActivityDeleteOpts struct {
	ID   *service.ID
	User *service.ID
}

// Delete removes an activity from the database by its ID.
func (svc *Service) Delete(ctx context.Context, opts ActivityDeleteOpts) error {
	if opts.ID == nil && opts.User == nil {
		return fmt.Errorf("%w: activity ID or user ID must be supplied", ErrInvalid)
	}

	filter := bson.D{}
	if opts.ID != nil {
		filter = append(filter, bson.E{Key: "_id", Value: opts.ID.ConvertID()})
	}
	if opts.User != nil {
		filter = append(filter, bson.E{Key: "userID", Value: opts.User.ConvertID()})
	}

	_, err := svc.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	return nil
}
