package service

import (
	"context"
	"errors"
)

type CtxKey string

// A resource is a top level element in the database.
// A resource must be accessible via a unique ID.
type Resource interface {
	GetID() ID
	GetCreatedDate() Time
}

// An attribute is a sub element of a resource.
// An attribute must be accessible via a unique ID.
type Attribute interface {
	GetID() ID
}

// CRUDService is a generic interface for a service that can perform CRUD operations on a resource.
// The intention is to provide a means to abstract away the underlying database & enforce rules & results that are easily managed by an API.
type CRUDService[T Resource] interface {
	// Create creates a new resource in the database.
	Create(ctx context.Context, resource T) (ID, error)

	// FindResource queries the database for a resource matching the provided criteria.
	FindResource(ctx context.Context, resource *T, criteria interface{}) error

	// Read retrieves a resource from the database based on an ID.
	Read(ctx context.Context, id ID, resource *T) error

	// FindAll retrieves all resources from the database matching the provided criteria.
	FindAll(ctx context.Context, resources *[]T, criteria interface{}) error

	// ReadAll retrieves all resources from the database.
	ReadAll(ctx context.Context, resources *[]T) error

	// UpdateWithCriteria updates a resource in the database matching the provided criteria.
	UpdateWithCriteria(ctx context.Context, resource T, criteria interface{}) error

	// Update updates a resource in the database based on an ID.
	Update(ctx context.Context, resource T) error

	// DeleteWithCriteria deletes a resource from the database matching the provided criteria.
	DeleteWithCriteria(ctx context.Context, criteria interface{}) error

	// Delete deletes a resource from the database based on an ID.
	Delete(ctx context.Context, id ID) error

	// ReadAttribute retrieves an attribute from a resource with given ID.
	ReadAttribute(ctx context.Context, resourceID ID, attributeKey string, attributes interface{}) error

	// AppendAttribute appends an attribute to an array contained in a resource with given ID.
	AppendAttribute(ctx context.Context, resourceID ID, attributeKey string, attribute Attribute) (ID, error)

	// RemoveAttribute removes an attribute from an array contained in a resource with given ID.
	RemoveAttribute(ctx context.Context, resourceID ID, attributeKey string, attributeID ID) error
}

var (
	ErrIDConversionError     = errors.New("error converting ID")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrUnknownError          = errors.New("unknown error")
	ErrInvalidPointer        = errors.New("invalid pointer")
	ErrBadSyntax             = errors.New("bad syntax")
)
