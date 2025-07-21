// Move this logic to the API me thinks

package users

import (
	"context"
	"fmt"
	"time"

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

func (svc *Service) Setup(ctx context.Context) error {
	if err := svc.users.Setup(ctx); err != nil {
		return fmt.Errorf("failed to setup users: %w", err)
	}
	return nil
}

func (svc *Service) Create(ctx context.Context, user *Detail) (service.ID, error) {
	if user.CreatedDate == nil {
		now := time.Now()
		user.CreatedDate = &now
	}

	id, err := svc.users.Create(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

func (svc *Service) Get(ctx context.Context, id service.ID, user interface{}) error {
	if err := svc.users.Get(ctx, id, user); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	opts := challenges.ListOptions{
		User: &id,
	}

	// Only add challenges if the user is of type User
	if u, ok := user.(*User); ok {
		cs := make([]challenges.Challenge, 0)
		if err := svc.challenges.List(ctx, opts, &cs); err != nil {
			return fmt.Errorf("failed to list challenges for user: %w", err)
		}

		for _, c := range cs {
			u.Challenges = append(u.Challenges, c.ID)
		}
	}

	return nil
}

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

func (svc *Service) List(ctx context.Context, opts ListOptions, users interface{}) error {
	los := NewDetailListOptions().SetLimit(opts.Limit).SetSkip(opts.Skip)
	if err := svc.users.List(ctx, *los, &users); err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	return nil
}

func (svc *Service) Update(ctx context.Context, user *Detail) error {
	if err := svc.users.Update(ctx, *user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// TODO - ensure this is done with a transaction
func (svc *Service) Delete(ctx context.Context, id service.ID) error {
	if err := svc.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	memsOpts := challenges.MembershipDeleteOpts{
		User: &id,
	}
	if err := svc.memberships.Delete(ctx, memsOpts); err != nil {
		return fmt.Errorf("failed to delete memberships for user: %w", err)
	}

	actOpts := activities.ActivityDeleteOpts{
		User: &id,
	}
	if err := svc.activities.Delete(ctx, actOpts); err != nil {
		return fmt.Errorf("failed to delete activities for user: %w", err)
	}

	return nil
}
