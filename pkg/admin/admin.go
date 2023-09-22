package admin

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
)

type Admin struct {
	auth *auth.Auth
}

func NewAdmin(auth *auth.Auth) *Admin {
	return &Admin{
		auth,
	}
}

func (a *Admin) GetAdmin(ctx context.Context, id uuid.ID) (bool, error) {

	user, err := a.auth.GetUser(ctx, string(id))
	if err != nil {
		return false, err
	}

	if admin, ok := user.CustomClaims["admin"]; ok {
		return admin.(bool), nil
	}

	return false, nil

}

func (a *Admin) SetAdmin(ctx context.Context, id uuid.ID, admin bool) error {

	claims := map[string]interface{}{"admin": admin}
	if err := a.auth.SetCustomUserClaims(ctx, string(id), claims); err != nil {
		return err
	}

	return nil
}
