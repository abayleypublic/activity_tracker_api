package auth

import (
	"context"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

// Token is a type alias for auth.Token from the Firebase auth package.
type Token auth.Token

// Auth is a struct that wraps the Firebase auth client and includes the project ID.
type Auth struct {
	*auth.Client
	ProjectID string
}

// NewAuth initializes a new Auth struct. It takes a project ID as input and returns a pointer to an Auth struct and an error.
func NewAuth(projectID string) (*Auth, error) {

	cfg := &firebase.Config{
		ProjectID: projectID,
	}

	// Create a new Firebase app with the provided project ID.
	app, err := firebase.NewApp(context.Background(), cfg)
	if err != nil {
		return nil, terrors.InternalService("", "error getting auth client", nil)
	}

	// Get the auth client from the Firebase app.
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, terrors.BadRequest("", "error getting auth client", nil)
	}

	// Return a new Auth struct.
	return &Auth{
		client,
		projectID,
	}, nil
}

// GetAuthToken extracts the auth token from the Authorization header of a request.
// It returns the token as a string and an error.
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

// GetValidToken verifies an ID token and checks if it has been revoked.
// It returns a pointer to a Token and an error.
func (a *Auth) GetValidToken(t string) (*Token, error) {

	token, err := a.VerifyIDTokenAndCheckRevoked(context.Background(), t)
	if err != nil {
		return nil, terrors.Forbidden("", "error verifying ID token", nil)
	}

	res := Token(*token)
	return &res, nil
}
