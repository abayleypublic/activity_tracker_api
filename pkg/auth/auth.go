package auth

import (
	"context"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

// type Token struct {
// 	AccessToken string
// 	Admin       bool
// }

type Auth struct {
	*firebase.App
	ProjectID string
}

func NewAuth(projectID string) (*Auth, error) {

	cfg := &firebase.Config{
		ProjectID: projectID,
	}

	app, err := firebase.NewApp(context.Background(), cfg)
	if err != nil {
		return nil, terrors.InternalService("", "error getting auth client", nil)
	}

	return &Auth{
		app,
		projectID,
	}, nil
}

func (a *Auth) GetAuthToken(req typhon.Request) (string, error) {
	reqToken := req.Header.Get("Authorization")

	// If no token is supplied in the Authorization header, return error
	if reqToken == "" {
		return "", terrors.Unauthorized("", "token not supplied", nil)
	}

	splitToken := strings.Split(reqToken, "Bearer ")

	if len(splitToken) != 2 {
		return "", terrors.Forbidden("", "invalid token", nil)
	}

	return splitToken[1], nil
}

func (a *Auth) GetValidToken(t string) (*auth.Token, error) {

	client, err := a.Auth(context.Background())
	if err != nil {
		return nil, terrors.InternalService("", "error getting auth client", nil)
	}

	token, err := client.VerifyIDTokenAndCheckRevoked(context.Background(), t)
	if err != nil {
		return nil, terrors.Forbidden("", "error verifying ID token", nil)
	}

	return token, nil
}
