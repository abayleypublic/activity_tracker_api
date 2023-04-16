package admin

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
)

type Admin struct {
	*auth.Auth
}

func NewAdmin(auth *auth.Auth) *Admin {
	return &Admin{
		auth,
	}
}

func (a *Admin) GetAdmin(ctx context.Context, id string) (bool, error) {

	user, err := a.GetUser(ctx, id)
	if err != nil {
		return false, err
	}

	if admin, ok := user.CustomClaims["admin"]; ok {
		return admin.(bool), nil
	}

	return false, nil

}

func (a *Admin) SetAdmin(ctx context.Context, id string, admin bool) error {

	claims := map[string]interface{}{"admin": admin}
	if err := a.SetCustomUserClaims(ctx, id, claims); err != nil {
		return err
	}

	return nil
}
