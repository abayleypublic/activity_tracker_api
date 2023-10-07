package admin

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
)

// Admin is a struct that holds the auth object.
type Admin struct {
	auth *auth.Auth
}

// NewAdmin is a constructor function that creates a new Admin object.
// It takes an auth object as a parameter and returns a pointer to the Admin object.
func NewAdmin(auth *auth.Auth) *Admin {
	return &Admin{
		auth,
	}
}

// GetAdmin is a method on the Admin struct that checks if a user is an admin.
// It takes a context and a user ID as parameters.
// It returns a boolean indicating whether the user is an admin and an error if any occurred.
func (a *Admin) GetAdmin(ctx context.Context, id service.ID) (bool, error) {

	user, err := a.auth.GetUser(ctx, string(id))
	if err != nil {
		return false, err
	}

	if admin, ok := user.CustomClaims["admin"]; ok {
		return admin.(bool), nil
	}

	return false, nil

}

// SetAdmin is a method on the Admin struct that sets a user as an admin.
// It takes a context, a user ID, and a boolean indicating whether the user should be an admin as parameters.
// It returns an error if any occurred.
func (a *Admin) SetAdmin(ctx context.Context, id service.ID, admin bool) error {

	claims := map[string]interface{}{"admin": admin}
	if err := a.auth.SetCustomUserClaims(ctx, string(id), claims); err != nil {
		return err
	}

	return nil
}

func (a *Admin) RevokeTokens(ctx context.Context, id service.ID) error {
	return a.auth.RevokeRefreshTokens(ctx, string(id))
}
