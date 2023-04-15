package auth

import (
	"context"
	"errors"
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

const (
	ErrNoAuth      string = "no token supplied"
	ErrInvalidAuth string = "invalid token supplied"
)

func GetAuthToken(req typhon.Request) (string, error) {
	reqToken := req.Header.Get("Authorization")

	// If no token is supplied in the Authorization header, return error
	if reqToken == "" {
		return "", errors.New(ErrNoAuth)
	}

	splitToken := strings.Split(reqToken, "Bearer ")

	if len(splitToken) != 2 {
		return "", errors.New(ErrInvalidAuth)
	}

	return splitToken[1], nil
}

func GetValidToken(t string) (*auth.Token, error) {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return nil, terrors.InternalService("", "error getting auth client", nil)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, terrors.InternalService("", "error getting auth client", nil)
	}

	token, err := client.VerifyIDTokenAndCheckRevoked(context.Background(), t)
	if err != nil {
		return nil, terrors.Forbidden("", "error verifying ID token", nil)
	}

	return token, nil
}
