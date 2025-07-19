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
	Detail     `json:",inline"`
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

func (svc *Service) Get(ctx context.Context, id service.ID, user *User) error {
	if err := svc.users.Get(ctx, id, user); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	opts := challenges.ListOptions{
		User: &id,
	}

	cs := make([]challenges.Challenge, 0)
	if err := svc.challenges.List(ctx, opts, &cs); err != nil {
		return fmt.Errorf("failed to list challenges for user: %w", err)
	}

	for _, c := range cs {
		user.Challenges = append(user.Challenges, c.ID)
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
