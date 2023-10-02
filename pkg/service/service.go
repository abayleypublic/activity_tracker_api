package service

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	timeFormat = time.RFC3339
)

type Resource interface {
	GetID() uuid.ID
}

type res struct{}

func (r *res) GetID() uuid.ID {
	return ""
}

func (r *res) SetCreated(string) {}

type Attribute interface {
	GetID() uuid.ID
}

type Attr struct{}

func (a Attr) GetID() uuid.ID {
	return ""
}

type CRUDService[T Resource] interface {
	Create(ctx context.Context, resource T) (uuid.ID, error)
	Read(ctx context.Context, id uuid.ID, resource *T) error
	ReadAll(ctx context.Context, resources *[]T) error
	Update(ctx context.Context, resource T) error
	Delete(ctx context.Context, id uuid.ID) error
	ReadAttribute(ctx context.Context, resourceID uuid.ID, attributeKey string, attributes interface{}) error
	AppendAttribute(ctx context.Context, resourceID uuid.ID, attributeKey string, attribute Attribute) (uuid.ID, error)
	RemoveAttribute(ctx context.Context, resourceID uuid.ID, attributeID uuid.ID, attributeKey string) error
}

var (
	ErrIDConversionError     = errors.New("error converting ID")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrUnknownError          = errors.New("unknown error")
	ErrInvalidPointer        = errors.New("invalid pointer")
)

var (
	_ Attribute             = (*Attr)(nil)
	_ Resource              = (*res)(nil)
	_ CRUDService[Resource] = (*Service[Resource])(nil)
)

type Service[T Resource] struct {
	*mongo.Collection
	IDKey string
}

func New[T Resource](collection *mongo.Collection) *Service[T] {
	return &Service[T]{collection, "_id"}
}

func (s *Service[T]) Read(ctx context.Context, resourceID uuid.ID, resource *T) error {

	if err := s.FindOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}).Decode(resource); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return ErrResourceNotFound
		}
		return ErrUnknownError
	}

	return nil
}

// ReadAll retrieves all users from the MongoDB collection.
// It returns a slice of PartialUser instances and any error encountered.
func (s *Service[T]) ReadAll(ctx context.Context, resources *[]T) error {

	cur, err := s.Find(ctx, bson.D{})
	if err != nil {
		return ErrUnknownError
	}

	if err = cur.All(ctx, resources); err != nil {
		return ErrUnknownError
	}

	return nil

}

func (s *Service[T]) Create(ctx context.Context, resource T) (uuid.ID, error) {

	opts := options.Update().SetUpsert(true)
	res, err := s.UpdateOne(ctx, bson.D{{Key: s.IDKey, Value: bson.D{{Key: "$exists", Value: false}}}}, bson.M{"$set": &resource}, opts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrResourceAlreadyExists
		}
		return "", ErrUnknownError
	}

	return uuid.ID(res.UpsertedID.(string)), nil
}

func (s *Service[T]) Update(ctx context.Context, resource T) error {
	opts := options.Update().SetUpsert(true)
	res, err := s.UpdateOne(ctx, bson.D{{Key: "$set", Value: resource}}, opts)
	if err != nil {
		return ErrUnknownError
	}

	if res.UpsertedCount != 1 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *Service[T]) Delete(ctx context.Context, id uuid.ID) error {

	res, err := s.DeleteOne(ctx, bson.D{{Key: s.IDKey, Value: id}})
	if err != nil {
		return ErrUnknownError
	}

	if res.DeletedCount != 1 {
		return ErrResourceNotFound
	}

	return nil

}

func (s *Service[T]) ReadAttribute(ctx context.Context, resourceID uuid.ID, attributeKey string, attributes any) error {

	rv := reflect.ValueOf(attributes)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrInvalidPointer
	}

	opts := options.FindOne().SetProjection(bson.M{s.IDKey: 0, attributeKey: 1})

	raw, err := s.FindOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, opts).DecodeBytes()
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return ErrResourceNotFound
		}
		return ErrUnknownError
	}

	element := raw.Lookup(attributeKey)
	err = element.Unmarshal(attributes)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service[T]) AppendAttribute(ctx context.Context, resourceID uuid.ID, attributeKey string, attribute Attribute) (uuid.ID, error) {

	res, err := s.UpdateOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, bson.D{{Key: "$push", Value: bson.M{attributeKey: attribute}}})
	if err != nil {
		return "", ErrUnknownError
	}

	if res.ModifiedCount != 1 {
		return "", ErrResourceNotFound
	}

	return attribute.GetID(), nil
}

func (s *Service[T]) RemoveAttribute(ctx context.Context, resourceID uuid.ID, attributeID uuid.ID, attributeKey string) error {

	res, err := s.UpdateOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, bson.D{{Key: "$pull", Value: bson.D{{Key: attributeKey, Value: bson.D{{Key: s.IDKey, Value: attributeID}}}}}})
	if err != nil {
		return ErrUnknownError
	}

	if res.ModifiedCount != 1 {
		return ErrResourceNotFound
	}

	return nil
}
