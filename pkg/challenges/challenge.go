package challenges

import (
	"context"
	"fmt"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
)

type Challenge struct {
	Detail  `json:",inline" bson:",inline"`
	Members []service.ID `json:"members" bson:"members"`
}

type Service struct {
	challenges  *Details
	memberships *Memberships
}

func New(
	challenges *Details,
	memberships *Memberships,
) *Service {
	return &Service{
		challenges:  challenges,
		memberships: memberships,
	}
}

func (svc *Service) Setup(ctx context.Context) error {
	if err := svc.challenges.Setup(ctx); err != nil {
		return fmt.Errorf("failed to setup challenges: %w", err)
	}

	if err := svc.memberships.Setup(ctx); err != nil {
		return fmt.Errorf("failed to setup memberships: %w", err)
	}

	return nil
}

// TODO - fix and make sure this is done as a transaction
func (svc *Service) Create(ctx context.Context, challenge *Challenge) error {
	if len(challenge.Members) == 0 {
		return fmt.Errorf("%w: challenge must have at least one member", ErrInvalid)
	}

	if challenge.CreatedDate == nil {
		now := time.Now()
		challenge.CreatedDate = &now
	}

	if err := svc.challenges.Create(ctx, &challenge.Detail); err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}

	for _, userID := range challenge.Members {
		membership := Membership{
			Challenge: challenge.ID,
			User:      userID,
			Created:   challenge.CreatedDate,
		}
		if err := svc.memberships.Create(ctx, &membership); err != nil {
			return fmt.Errorf("failed to create memberships for challenge %s: %w", challenge.ID.ConvertID(), err)
		}
	}

	return nil
}

func (svc *Service) Get(ctx context.Context, id service.ID, challenge *Challenge) error {
	if err := svc.challenges.Get(ctx, id, challenge); err != nil {
		return fmt.Errorf("failed to get challenge: %w", err)
	}

	memsOpts := MembershipListOptions{
		Challenge: &id,
	}
	mems := make([]Membership, 0)
	if err := svc.memberships.List(ctx, memsOpts, &mems); err != nil {
		return fmt.Errorf("failed to get memberships for challenge %s: %w", id.ConvertID(), err)
	}

	for _, m := range mems {
		challenge.Members = append(challenge.Members, m.User)
	}

	return nil
}

type ListOptions struct {
	Limit int64
	Skip  int64

	User *service.ID
}

func NewListOptions() ListOptions {
	return ListOptions{}
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

func (svc *Service) List(ctx context.Context, opts ListOptions, challenges *[]Challenge) error {
	mems := make([]Membership, 0, opts.Limit)

	memsOpts := MembershipListOptions{
		Limit: opts.Limit,
		Skip:  opts.Skip,
		User:  opts.User,
	}
	if err := svc.memberships.List(ctx, memsOpts, mems); err != nil {
		return fmt.Errorf("failed to list memberships: %w", err)
	}

	cs := make(bson.A, 0, len(mems))
	for _, m := range mems {
		var c bson.Raw
		if err := svc.challenges.Get(ctx, m.Challenge, &c); err != nil {
			return fmt.Errorf("failed to get challenge %s from membership: %w", m.Challenge.ConvertID(), err)
		}
		cs = append(cs, c)
	}

	data, err := bson.Marshal(cs)
	if err != nil {
		return fmt.Errorf("failed to marshal challenges: %w", err)
	}

	if err := bson.Unmarshal(data, challenges); err != nil {
		return fmt.Errorf("failed to unmarshal challenges: %w", err)
	}

	return nil
}

type Operation interface {
	Execute(ctx context.Context, details *Details, memberships *Memberships) error
}

type SetDetailOperation struct {
	ChallengeID service.ID
	Detail      Detail
}

func (o SetDetailOperation) Execute(ctx context.Context, details *Details, _ *Memberships) error {
	if err := details.Update(ctx, o.Detail); err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}
	return nil
}

type SetMemberOperation struct {
	Challenge service.ID
	User      service.ID
	Member    bool
}

func (o SetMemberOperation) Execute(ctx context.Context, _ *Details, memberships *Memberships) error {
	if o.Member {
		now := time.Now()
		membership := Membership{
			Challenge: o.Challenge,
			User:      o.User,
			Created:   &now,
		}

		if err := memberships.Create(ctx, &membership); err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		return nil
	}

	deleteOpts := MembershipDeleteOpts{
		Challenge: &o.Challenge,
		User:      &o.User,
	}

	if err := memberships.Delete(ctx, deleteOpts); err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	return nil
}

// TODO - ensure this is done as a transaction
func (svc *Service) Update(ctx context.Context, operations ...Operation) error {
	for _, op := range operations {
		if err := op.Execute(ctx, svc.challenges, svc.memberships); err != nil {
			return err
		}
	}

	return nil
}

// TODO - ensure this is done as a transaction
func (svc *Service) Delete(ctx context.Context, challengeID service.ID) error {
	if err := svc.challenges.Delete(ctx, challengeID); err != nil {
		return fmt.Errorf("failed to delete challenge: %w", err)
	}

	opts := MembershipDeleteOpts{
		Challenge: &challengeID,
	}

	if err := svc.memberships.Delete(ctx, opts); err != nil {
		return fmt.Errorf("failed to delete memberships: %w", err)
	}

	return nil
}
