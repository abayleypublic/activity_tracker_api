// Move this logic to the API me thinks

package users

import (
	"context"
	"fmt"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
)

type User struct {
	Detail     `json:",inline" bson:",inline"`
	Challenges []service.ID `json:"challenges"`
}

type Service struct {
	users       *Details
	memberships *challenges.Memberships
	challenges  *challenges.Service
	activities  *activities.Service
}

func New(
	users *Details,
	memberships *challenges.Memberships,
	challenges *challenges.Service,
	activities *activities.Service,
) *Service {
	return &Service{
		users:       users,
		memberships: memberships,
		challenges:  challenges,
		activities:  activities,
	}
}

// Setup initializes the user service, setting up the underlying database and collections.
func (svc *Service) Setup(ctx context.Context) error {
	if err := svc.users.Setup(ctx); err != nil {
		return fmt.Errorf("failed to setup users: %w", err)
	}
	return nil
}

// Create adds a new user to the database and returns the user's ID.
func (svc *Service) Create(ctx context.Context, user *Detail) (service.ID, error) {
	id, err := svc.users.Create(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

// Get retrieves a user by their ID and populates the user's challenges if they are of type User.
func (svc *Service) Get(ctx context.Context, id service.ID, user interface{}) error {
	if err := svc.users.Get(ctx, id, user); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	opts := challenges.ListOptions{
		User: &id,
	}

	// Only add challenges if the user is of type User
	if u, ok := user.(*User); ok {
		cs := make([]challenges.Detail, 0)
		if err := svc.challenges.List(ctx, opts, &cs); err != nil {
			return fmt.Errorf("failed to list challenges for user: %w", err)
		}

		for _, c := range cs {
			u.Challenges = append(u.Challenges, c.ID)
		}
	}

	return nil
}

// GetByEmail retrieves a user by their email address.
func (svc *Service) GetByEmail(ctx context.Context, email string) (*Detail, error) {
	opts := NewDetailListOptions().
		SetEmail(email)

	users := make([]Detail, 0)
	if err := svc.users.List(ctx, *opts, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users found with email %s", email)
	}

	return &users[0], nil
}

type ListOptions struct {
	Limit int64
	Skip  int64
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

// List retrieves users based on the given options.
func (svc *Service) List(ctx context.Context, opts ListOptions, users interface{}) error {
	los := NewDetailListOptions().SetLimit(opts.Limit).SetSkip(opts.Skip)
	if err := svc.users.List(ctx, *los, &users); err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	return nil
}

// Update modifies an existing user in the database.
func (svc *Service) Update(ctx context.Context, user *Detail) error {
	if err := svc.users.Update(ctx, *user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user from the database and cleans up associated data.
func (svc *Service) Delete(ctx context.Context, id service.ID) error {
	session, err := svc.users.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sCtx context.Context) (interface{}, error) {
		if err := svc.users.Delete(sCtx, id); err != nil {
			return nil, fmt.Errorf("failed to delete user: %w", err)
		}

		memsOpts := challenges.MembershipDeleteOpts{
			User: &id,
		}
		if err := svc.memberships.Delete(sCtx, memsOpts); err != nil {
			return nil, fmt.Errorf("failed to delete memberships for user: %w", err)
		}

		actOpts := activities.ActivityDeleteOpts{
			User: &id,
		}
		if err := svc.activities.Delete(sCtx, actOpts); err != nil {
			return nil, fmt.Errorf("failed to delete activities for user: %w", err)
		}
		return nil, nil
	})

	return err
}
