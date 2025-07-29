package challenges

import (
	"context"
	"fmt"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func (svc *Service) Create(ctx context.Context, challenge *Challenge) (service.ID, error) {
	session, err := svc.challenges.Database().Client().StartSession()
	if err != nil {
		return "", fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	var cID service.ID
	_, err = session.WithTransaction(ctx, func(sCtx context.Context) (interface{}, error) {
		cID, err = svc.challenges.Create(sCtx, &challenge.Detail)
		if err != nil {
			return "", fmt.Errorf("failed to create challenge: %w", err)
		}

		for _, userID := range challenge.Members {
			membership := Membership{
				Challenge: challenge.ID,
				User:      userID,
				Created:   challenge.CreatedDate,
			}
			if err := svc.memberships.Create(sCtx, &membership); err != nil {
				return "", fmt.Errorf("failed to create memberships for challenge %s: %w", challenge.ID.ConvertID(), err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to create challenge in transaction: %w", err)
	}

	return cID, nil
}

func (svc *Service) Get(ctx context.Context, id service.ID, challenge interface{}) error {
	if err := svc.challenges.Get(ctx, id, challenge); err != nil {
		return fmt.Errorf("failed to get challenge: %w", err)
	}

	if ch, ok := challenge.(*Challenge); ok {
		memsOpts := MembershipListOptions{
			Challenge: &id,
		}
		mems := make([]Membership, 0)
		if err := svc.memberships.List(ctx, memsOpts, &mems); err != nil {
			return fmt.Errorf("failed to get memberships for challenge %s: %w", id.ConvertID(), err)
		}

		for _, m := range mems {
			ch.Members = append(ch.Members, m.User)
		}
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

func (svc *Service) List(ctx context.Context, opts ListOptions, challenges interface{}) error {
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
	Detail Detail
}

func (o SetDetailOperation) Execute(ctx context.Context, details *Details, _ *Memberships) error {
	if err := details.Update(ctx, o.Detail); err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
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
			Created:   now,
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

func (svc *Service) Update(ctx context.Context, operations ...Operation) error {
	session, err := svc.challenges.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sCtx context.Context) (interface{}, error) {
		for _, op := range operations {
			if err := op.Execute(sCtx, svc.challenges, svc.memberships); err != nil {
				return nil, fmt.Errorf("failed to execute operation: %w", err)
			}
		}
		return nil, nil
	})

	return err
}

func (svc *Service) Delete(ctx context.Context, challengeID service.ID) error {
	session, err := svc.challenges.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sCtx context.Context) (interface{}, error) {
		if err := svc.challenges.Delete(sCtx, challengeID); err != nil {
			return nil, fmt.Errorf("failed to delete challenge: %w", err)
		}

		opts := MembershipDeleteOpts{
			Challenge: &challengeID,
		}

		if err := svc.memberships.Delete(sCtx, opts); err != nil {
			return nil, fmt.Errorf("failed to delete memberships: %w", err)
		}

		return nil, nil
	})

	return err
}
