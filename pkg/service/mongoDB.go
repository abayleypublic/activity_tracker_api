package service

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_ CRUDService[Resource] = (*MongoDBService[Resource])(nil)
)

type MongoDBService[T Resource] struct {
	*mongo.Collection
	IDKey string
}

func New[T Resource](collection *mongo.Collection) *MongoDBService[T] {
	return &MongoDBService[T]{collection, "_id"}
}

func (s *MongoDBService[T]) FindResource(ctx context.Context, resource interface{}, criteria interface{}) error {

	if err := s.FindOne(ctx, criteria).Decode(resource); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return ErrResourceNotFound
		}
		return ErrUnknownError
	}

	return nil
}

// Permitted checks if a user has permission to perform an operation on a resource.
// Electing to go this route instead of using queries to enforce permissions
// because it provides easier distinction of whether not found or not authorised.
// It will be a bit slower due to multiple db reads so may have to revisit this.
func (s *MongoDBService[T]) Permitted(ctx context.Context, criteria interface{}, op Operation) (bool, error) {

	var res T
	if err := s.FindResource(ctx, &res, criteria); err != nil {
		return false, ErrResourceNotFound
	}

	rc, err := GetActorContext(ctx)
	if err != nil {
		return false, nil
	}

	switch op {
	case READ:
		return res.CanBeReadBy(rc.UserID, rc.Admin), nil
	case UPDATE:
		return res.CanBeUpdatedBy(rc.UserID, rc.Admin), nil
	case DELETE:
		return res.CanBeDeletedBy(rc.UserID, rc.Admin), nil
	}

	return false, nil
}

func (s *MongoDBService[T]) Read(ctx context.Context, resourceID ID, resource *T) error {

	ok, err := s.Permitted(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, READ)
	if err != nil {
		log.Println(err)
		return err
	}
	if !ok {
		log.Println("Forbidden")
		return ErrForbidden
	}

	return s.FindResource(ctx, resource, bson.D{{Key: s.IDKey, Value: resourceID}})
}

func (s *MongoDBService[T]) ReadRaw(ctx context.Context, resourceID ID, resource interface{}) error {

	ok, err := s.Permitted(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, READ)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	return s.FindResource(ctx, resource, bson.D{{Key: s.IDKey, Value: resourceID}})
}

func (s *MongoDBService[T]) FindAll(ctx context.Context, resources interface{}, criteria interface{}) error {

	cur, err := s.Find(ctx, criteria)
	if err != nil {
		log.Println("Find")
		log.Println(err)
		return ErrUnknownError
	}

	if err = cur.All(ctx, resources); err != nil {
		log.Println("All")
		log.Println(err)
		return ErrUnknownError
	}

	return nil

}

// ReadAll retrieves all users from the MongoDB collection.
// It returns a slice of PartialUser instances and any error encountered.
func (s *MongoDBService[T]) ReadAll(ctx context.Context, resources *[]T) error {
	return s.FindAll(ctx, resources, bson.D{})
}

// ReadAllRaw is the same as ReadAll without the type checking
func (s *MongoDBService[T]) ReadAllRaw(ctx context.Context, resources interface{}) error {
	return s.FindAll(ctx, resources, bson.D{})
}

func (s *MongoDBService[T]) Create(ctx context.Context, resource T) (ID, error) {
	res, err := s.InsertOne(ctx, &resource)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrResourceAlreadyExists
		}
		return "", ErrUnknownError
	}

	return ID(res.InsertedID.(string)), nil
}

func (s *MongoDBService[T]) UpdateWithCriteria(ctx context.Context, resource T, criteria interface{}) error {

	ok, err := s.Permitted(ctx, criteria, UPDATE)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	opts := options.Update().SetUpsert(true)
	res, err := s.UpdateOne(ctx, criteria, bson.D{{Key: "$set", Value: resource}}, opts)
	if err != nil {
		return ErrUnknownError
	}

	if res.UpsertedCount != 1 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *MongoDBService[T]) Update(ctx context.Context, resource T) error {
	return s.UpdateWithCriteria(ctx, resource, bson.D{})
}

func (s *MongoDBService[T]) DeleteWithCriteria(ctx context.Context, criteria interface{}) error {

	ok, err := s.Permitted(ctx, criteria, DELETE)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	res, err := s.DeleteOne(ctx, criteria)
	if err != nil {
		return ErrUnknownError
	}

	if res.DeletedCount != 1 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *MongoDBService[T]) Delete(ctx context.Context, id ID) error {
	return s.DeleteWithCriteria(ctx, bson.D{{Key: s.IDKey, Value: id}})
}

func (s *MongoDBService[T]) ReadAttribute(ctx context.Context, resourceID ID, attributeKey string, attributes interface{}) error {

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

func (s *MongoDBService[T]) ReadSingleAttribute(ctx context.Context, resourceID ID, attributeKey string, attributeID ID, attribute interface{}) error {

	// Set the projection to only return the elements that match a condition & then only return one
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

func (s *MongoDBService[T]) AppendAttribute(ctx context.Context, resourceID ID, attributeKey string, attribute Attribute) (ID, error) {
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

func (s *MongoDBService[T]) RemoveAttribute(ctx context.Context, resourceID ID, attributeKey string, attributeID ID) error {

	res, err := s.UpdateOne(ctx, bson.D{{Key: s.IDKey, Value: resourceID}}, bson.D{{Key: "$pull", Value: bson.M{"$or": []bson.D{{{Key: attributeKey, Value: attributeID}}, {{Key: attributeKey, Value: bson.M{s.IDKey: attributeID}}}}}}})
	if err != nil {
		return ErrUnknownError
	}

	if res.ModifiedCount != 1 {
		return ErrResourceNotFound
	}

	return nil
}
