// Package challenges provides functionality related to challenges in the activity tracker.
package challenges

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
)

// Member is a type that embeds the User type from the users package.
// It represents a user who is a member of a challenge.
type Member struct {
	users.User
}
