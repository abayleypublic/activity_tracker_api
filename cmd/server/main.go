package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/AustinBayley/activity_tracker_api/pkg/api"
)

const (
	DEV  string = "dev"
	STG  string = "stg"
	PROD string = "prod"
)

const (
	projectID string = "542135123656"
)

type MongoCredentials struct {
	username string
	password string
}

func getMongoCredentials() (*MongoCredentials, error) {

	// Create the client.
	ctx := context.Background()
	secretsClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
		return nil, err
	}
	defer secretsClient.Close()

	// Get username
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/dbUsername/versions/latest", projectID),
	}

	dbUsername, err := secretsClient.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Fatalf("failed to access secret version: %v", err)
		return nil, err
	}

	// Get password
	accessRequest = &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/dbPassword/versions/latest", projectID),
	}

	dbPassword, err := secretsClient.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Fatalf("failed to access secret version: %v", err)
		return nil, err
	}

	return &MongoCredentials{
		username: string(dbUsername.Payload.Data),
		password: string(dbPassword.Payload.Data),
	}, nil
}

func main() {

	env := os.Getenv("ENVIRONMENT")

	var mongoURI string
	switch env {
	// case STG:
	// 	creds, _ := getMongoCredentials()
	// 	mongoUri = fmt.Sprintf("mongodb+srv://%s:%s@activity-tracker-stg.ur4pqgv.mongodb.net/?retryWrites=true&w=majority", creds.username, creds.password)
	case PROD:
		creds, _ := getMongoCredentials()
		mongoURI = fmt.Sprintf("mongodb+srv://%s:%s@activity-tracker.ur4pqgv.mongodb.net/?retryWrites=true&w=majority", creds.username, creds.password)
	default:
		mongoURI = os.Getenv("MONGODB_URI")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}

	cfg := api.NewConfig(mongoURI, port, projectID)

	a, err := api.NewAPI(cfg)

	if err != nil {
		panic(err)
	}

	a.Start()

}
