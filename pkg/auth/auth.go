package auth

import (
	"context"
	"errors"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"github.com/monzo/typhon"
)

// Token is a type alias for auth.Token from the Firebase auth package.
type Token auth.Token

// Auth is a struct that wraps the Firebase auth client and includes the project ID.
type Auth struct {
	*auth.Client
	ProjectID string
}

var (
	errAuthClient       = errors.New("error getting auth client")
	ErrTokenNotSupplied = errors.New("token not supplied")
	ErrInvalidToken     = errors.New("invalid token")
	ErrParsingToken     = errors.New("error parsing token")
)

// NewAuth initializes a new Auth struct. It takes a project ID as input and returns a pointer to an Auth struct and an error.
func NewAuth(projectID string) (*Auth, error) {

	cfg := &firebase.Config{
		ProjectID: projectID,
	}

	// Create a new Firebase app with the provided project ID.
	app, err := firebase.NewApp(context.Background(), cfg)
	if err != nil {
		return nil, errAuthClient
	}

	// Get the auth client from the Firebase app.
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, errAuthClient
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
		return "", ErrTokenNotSupplied
	}

	splitToken := strings.Split(reqToken, "Bearer ")

	if len(splitToken) != 2 {
		return "", ErrInvalidToken
	}

	return splitToken[1], nil
}

// GetToken verifies an ID token.
// It returns a pointer to a Token and an error.
func (a *Auth) GetToken(ctx context.Context, t string) (*Token, error) {

	token, err := a.VerifyIDToken(ctx, t)
	if err != nil {
		return nil, ErrInvalidToken
	}

	res := Token(*token)
	return &res, nil
}

// GetValidToken verifies an ID token and checks if it has been revoked.
// It returns a pointer to a Token and an error.
func (a *Auth) GetValidToken(ctx context.Context, t string) (*Token, error) {

	token, err := a.VerifyIDTokenAndCheckRevoked(ctx, t)
	if err != nil {
		return nil, ErrInvalidToken
	}

	res := Token(*token)
	return &res, nil
}

func (a *Auth) GetUserID(ctx context.Context, token Token) uuid.ID {
	return uuid.ID(token.UID)
}

func (a *Auth) IsAdmin(token Token) bool {
	if admin, ok := token.Claims["admin"]; ok {
		return admin.(bool)
	}
	return false
}
