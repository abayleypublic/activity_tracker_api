package service

import (
	"context"
	"errors"
)

type CtxKey string

type RequestContext struct {
	Admin  bool
	UserID ID
}

func GetActorContext(ctx context.Context) (RequestContext, error) {
	val := ctx.Value(UserCtxKey)
	if val == nil {
		return RequestContext{}, ErrInvalidContext
	}

	d, ok := val.(RequestContext)
	if !ok {
		return RequestContext{}, ErrInvalidContext
	}

	return d, nil
}

// A resource is a top level element in the database.
// A resource must be accessible via a unique ID.
type Resource interface {
	GetID() ID
	GetCreatedDate() Time
	CanBeReadBy(userID ID, admin bool) bool
	CanBeUpdatedBy(userID ID, admin bool) bool
	CanBeDeletedBy(userID ID, admin bool) bool
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
	FindResource(ctx context.Context, resource interface{}, criteria interface{}) error

	// Read retrieves a resource from the database based on an ID.
	Read(ctx context.Context, id ID, resource *T) error

	// Read retrieves a resource from the database based on an ID but does not parse it into a specific type.
	ReadRaw(ctx context.Context, id ID, resource interface{}) error

	// FindAll retrieves all resources from the database matching the provided criteria.
	// To enable parsing to other object types, this will accept an interface
	FindAll(ctx context.Context, resources interface{}, criteria interface{}) error

	// ReadAll retrieves all resources from the database.
	ReadAll(ctx context.Context, resources *[]T) error

	// ReadAll retrieves all resources from the database but does not parse them into a specific type.
	ReadAllRaw(ctx context.Context, resources interface{}) error

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
	ErrIDConversionError = errors.New("error converting ID")
	ErrInvalidPointer    = errors.New("invalid pointer")
	ErrInvalidContext    = errors.New("invalid context")

	ErrForbidden             = errors.New("user is not authorised to perform this action")
	ErrBadSyntax             = errors.New("bad syntax")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrUnknownError          = errors.New("unknown error")
)

const (
	UnknownUser = ID("unknown")
	UserCtxKey  = CtxKey("user")
)

type Operation string

const (
	READ   Operation = "read"
	DELETE Operation = "delete"
	UPDATE Operation = "update"
)
