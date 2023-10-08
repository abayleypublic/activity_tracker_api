package challenges

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
)

// This is currently quite an expensive operation as it needs to read user, read their activities & then read the challenge
func (c *Challenges) GetProgress(ctx context.Context, challengeID service.ID, memberID service.ID) (targets.Progress, error) {

	isMember, err := c.users.IsUserMember(ctx, memberID, challengeID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, service.ErrResourceNotFound
	}

	activities, err := c.users.ReadUserActivities(ctx, memberID)
	if err != nil {
		return nil, err
	}

	challenge := Challenge{}
	if err := c.Read(ctx, challengeID, &challenge); err != nil {
		return nil, err
	}

	progress, err := challenge.Target.Evaluate(ctx, activities)
	if err != nil {
		return nil, err
	}

	return progress, nil

}
