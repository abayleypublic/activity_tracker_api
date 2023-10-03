package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Resource interface {
	GetID() ID
	GetCreatedDate() Time
}

type Attribute interface {
	GetID() ID
}

type CRUDService[T Resource] interface {
	Create(ctx context.Context, resource T) (ID, error)
	Read(ctx context.Context, id ID, resource *T) error
	ReadAll(ctx context.Context, resources *[]T) error
	Update(ctx context.Context, resource T) error
	Delete(ctx context.Context, id ID) error
	ReadAttribute(ctx context.Context, resourceID ID, attributeKey string, attributes interface{}) error
	AppendAttribute(ctx context.Context, resourceID ID, attributeKey string, attribute Attribute) (ID, error)
	RemoveAttribute(ctx context.Context, resourceID ID, attributeKey string, attributeID ID) error
}

var (
	ErrIDConversionError     = errors.New("error converting ID")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrUnknownError          = errors.New("unknown error")
	ErrInvalidPointer        = errors.New("invalid pointer")
)

var (
	_ CRUDService[Resource] = (*Service[Resource])(nil)
)

type Service[T Resource] struct {
	*mongo.Collection
	IDKey string
}

func New[T Resource](collection *mongo.Collection) *Service[T] {
	return &Service[T]{collection, "_id"}
}

func (s *Service[T]) Read(ctx context.Context, resourceID ID, resource *T) error {

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

func (s *Service[T]) Create(ctx context.Context, resource T) (ID, error) {
	opts := options.Update().SetUpsert(true)
	res, err := s.UpdateOne(ctx, bson.D{{Key: s.IDKey, Value: bson.D{{Key: "$exists", Value: false}}}}, bson.M{"$set": &resource}, opts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrResourceAlreadyExists
		}
		log.Println(err)
		return "", ErrUnknownError
	}

	return ID(res.UpsertedID.(string)), nil
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

func (s *Service[T]) Delete(ctx context.Context, id ID) error {

	res, err := s.DeleteOne(ctx, bson.D{{Key: s.IDKey, Value: id}})
	if err != nil {
		return ErrUnknownError
	}

	if res.DeletedCount != 1 {
		return ErrResourceNotFound
	}

	return nil

}

func (s *Service[T]) ReadAttribute(ctx context.Context, resourceID ID, attributeKey string, attributes interface{}) error {

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
		return ErrInvalidPointer
	}

	return nil
}

func (s *Service[T]) ReadSingleAttribute(ctx context.Context, resourceID ID, attributeKey string, attributeID ID, attribute interface{}) error {

	opts := options.FindOne().SetProjection(bson.M{s.IDKey: 0, attributeKey: bson.M{"$elemMatch": bson.M{s.IDKey: attributeID}}})
	raw, err := s.FindOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, opts).DecodeBytes()

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return ErrResourceNotFound
		}
		return ErrUnknownError
	}

	element := raw.Lookup(attributeKey)
	arr, ok := element.ArrayOK()
	if !ok {
		return ErrResourceNotFound
	}

	err = arr.Index(0).Value().Unmarshal(attribute)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service[T]) AppendAttribute(ctx context.Context, resourceID ID, attributeKey string, attribute Attribute) (ID, error) {

	res, err := s.UpdateOne(ctx,
		bson.D{
			{Key: s.IDKey, Value: resourceID},
			{Key: attributeKey, Value: bson.M{"$ne": string(attribute.GetID())}},
			{Key: fmt.Sprintf("%s.%s", attributeKey, s.IDKey), Value: bson.M{"$ne": string(attribute.GetID())}},
		},
		bson.D{{Key: "$push", Value: bson.M{attributeKey: attribute}}},
	)
	if err != nil {
		return "", ErrUnknownError
	}

	if res.ModifiedCount != 1 {
		if res.MatchedCount == 0 {
			return "", ErrResourceAlreadyExists
		}

		return "", ErrResourceNotFound
	}

	return attribute.GetID(), nil
}

func (s *Service[T]) RemoveAttribute(ctx context.Context, resourceID ID, attributeKey string, attributeID ID) error {

	res, err := s.UpdateOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, bson.D{{Key: "$pull", Value: bson.D{{Key: attributeKey, Value: bson.M{"$or": bson.D{{Key: s.IDKey, Value: attributeID}, {Key: "$eq", Value: attributeID}}}}}}})
	if err != nil {
		return ErrUnknownError
	}

	if res.ModifiedCount != 1 {
		return ErrResourceNotFound
	}

	return nil
}
