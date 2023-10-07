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
	projectID string = "activity-tracker-bfaaa"
	dbName    string = "activity-tracker"
)

type MongoCredentials struct {
	username string
	password string
}

type MapsCredentials struct {
	key string
}

func getSecret(ctx context.Context, secrets *secretmanager.Client, key string, version *string) (*secretmanagerpb.AccessSecretVersionResponse, error) {

	var v string
	if version == nil {
		v = "latest"
	} else {
		v = *version
	}

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, key, v),
	}

	secret, err := secrets.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func getMapsCredentials(ctx context.Context, secrets *secretmanager.Client) (*MapsCredentials, error) {
	secret, err := getSecret(ctx, secrets, "mapsKey", nil)
	if err != nil {
		return nil, err
	}

	return &MapsCredentials{
		key: string(secret.Payload.Data),
	}, nil
}

func getMongoCredentials(ctx context.Context, secrets *secretmanager.Client) (*MongoCredentials, error) {

	dbUsername, err := getSecret(ctx, secrets, "dbUsername", nil)
	if err != nil {
		return nil, err
	}

	dbPassword, err := getSecret(ctx, secrets, "dbPassword", nil)
	if err != nil {
		return nil, err
	}

	return &MongoCredentials{
		username: string(dbUsername.Payload.Data),
		password: string(dbPassword.Payload.Data),
	}, nil
}

func main() {

	var env api.Environment
	if value, ok := os.LookupEnv("ENVIRONMENT"); ok {
		env = api.Environment(value)
	} else {
		env = api.DEV
	}

	ctx := context.Background()

	secretsClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup secrets client: %v", err)
	}

	mapsCreds, err := getMapsCredentials(ctx, secretsClient)
	if err != nil {
		log.Fatalf("failed to get maps credentials: %v", err)
	}

	var mongoURI string
	switch env {
	case api.PROD:
		dbCreds, err := getMongoCredentials(ctx, secretsClient)
		if err != nil {
			log.Fatalf("failed to get db credentials: %v", err)
		}
		mongoURI = fmt.Sprintf("mongodb+srv://%s:%s@activity-tracker.ur4pqgv.mongodb.net/?retryWrites=true&w=majority", dbCreds.username, dbCreds.password)
	default:
		mongoURI = os.Getenv("MONGODB_URI")
	}
	secretsClient.Close()

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalln(err)
	}

	cfg := api.NewConfig(env, mongoURI, dbName, port, projectID, mapsCreds.key)

	a, err := api.NewAPI(cfg)

	if err != nil {
		log.Fatalln(err)
	}

	a.Start()

}
